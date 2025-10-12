// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package ocr

import (
	"cmp"
	"image"
	"math"
	"slices"

	"github.com/kvarenzn/ssm/ort"
	"golang.org/x/image/draw"
)

const (
	inputName  = "x"
	outputName = "fetch_name_0"

	detMaxLength = 960
	blockSize    = 32

	unclipRatio     = 1.5
	binaryThreshold = 0.3

	recBatch         = 6
	recDefaultHeight = 48
	recDefaultWidth  = 320
)

func scaledSizeOf(width, height int) (int, int) {
	maxSide := max(width, height)
	scaleFactor := min(detMaxLength/float64(maxSide), 1)

	newWidth := max(int(math.Round(float64(width)*scaleFactor/blockSize)), 1) * blockSize
	newHeight := max(int(math.Round(float64(height)*scaleFactor/blockSize)), 1) * blockSize
	return newWidth, newHeight
}

func clamp[T cmp.Ordered](x, lo, hi T) T {
	if x < lo {
		return lo
	} else if x > hi {
		return hi
	} else {
		return x
	}
}

func findContours(data []float32, width, height int, original image.Rectangle) []image.Rectangle {
	visited := make([]bool, width*height)
	rects := []image.Rectangle{}

	ow := original.Dx()
	oh := original.Dy()
	rw := float64(ow) / float64(width)
	rh := float64(oh) / float64(height)

	for y := range height {
		for x := range width {
			pos := y*width + x
			if data[pos] >= binaryThreshold && !visited[pos] {
				minX, maxX := x, x
				minY, maxY := y, y

				stack := []struct{ x, y int }{{x, y}}
				visited[pos] = true

				for len(stack) > 0 {
					c := stack[len(stack)-1]
					stack = stack[:len(stack)-1]

					minX = min(c.x, minX)
					maxX = max(c.x, maxX)
					minY = min(c.y, minY)
					maxY = max(c.y, maxY)

					for _, d := range []struct{ x, y int }{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
						nx := c.x + d.x
						ny := c.y + d.y
						npos := ny*width + nx
						if nx < 0 || nx >= width || ny < 0 || ny >= height || data[npos] < binaryThreshold || visited[npos] {
							continue
						}

						visited[npos] = true
						stack = append(stack, struct {
							x int
							y int
						}{nx, ny})
					}
				}

				if maxX+1-minX <= 3 || maxY+1-minY <= 3 {
					continue
				}

				w := float64(maxX + 1 - minX)
				h := float64(maxY + 1 - minY)

				distance := w * h * unclipRatio / (w + h) / 2
				rects = append(rects, image.Rect(
					clamp(int(math.Round((float64(minX)-distance)*rw)), 0, ow),
					clamp(int(math.Round((float64(minY)-distance)*rh)), 0, oh),
					clamp(int(math.Round((float64(maxX)+distance)*rw)), 0, ow),
					clamp(int(math.Round((float64(maxY)+distance)*rh)), 0, oh),
				))
			}
		}
	}

	slices.SortFunc(rects, func(x, y image.Rectangle) int {
		bx, by := x.Size(), y.Size()
		rx := float64(bx.X) / float64(bx.Y)
		ry := float64(by.X) / float64(by.Y)

		if rx < ry {
			return -1
		} else if rx == ry {
			return 0
		} else {
			return 1
		}
	})

	return rects
}

func ratioOf(i image.Rectangle) float64 {
	sz := i.Size()
	return float64(sz.X) / float64(sz.Y)
}

type LabeledBox struct {
	Rect  image.Rectangle
	Label string
}

type OCR struct {
	env     *ort.Env
	memInfo *ort.MemoryInfo
	det     *ort.Session
	rec     *ort.Session

	detImageBuffer            *image.NRGBA
	detInputTensor            *ort.Tensor[float32]
	detOutputTensor           *ort.Tensor[float32]
	detWidth, detHeight       int
	originWidth, originHeight int
}

func (o *OCR) scaleImage(i image.Image) image.Image {
	if o.detImageBuffer == nil {
		return i
	}

	draw.BiLinear.Scale(o.detImageBuffer, image.Rect(0, 0, o.detWidth, o.detHeight), i, i.Bounds(), draw.Src, nil)
	return o.detImageBuffer
}

func (o *OCR) Det(img image.Image) ([]image.Rectangle, error) {
	b := img.Bounds().Size()
	w, h := b.X, b.Y
	var err error
	if w != o.originWidth || h != o.originHeight {
		o.detWidth, o.detHeight = scaledSizeOf(w, h)
		if w == o.detWidth && h == o.detHeight {
			o.detImageBuffer = nil
		} else {
			o.detImageBuffer = image.NewNRGBA(image.Rect(0, 0, o.detWidth, o.detHeight))
		}

		if o.detInputTensor != nil {
			o.detInputTensor.Close()
		}

		input := make([]float32, o.detWidth*o.detHeight*3)
		o.detInputTensor, err = o.memInfo.NewTensorF32(input, []int64{1, 3, int64(o.detHeight), int64(o.detWidth)})
		if err != nil {
			return nil, err
		}

		if o.detOutputTensor != nil {
			o.detOutputTensor.Close()
		}

		output := make([]float32, o.detWidth*o.detHeight)
		o.detOutputTensor, err = o.memInfo.NewTensorF32(output, []int64{1, 1, int64(o.detHeight), int64(o.detWidth)})
		if err != nil {
			return nil, err
		}

		o.originWidth, o.originHeight = w, h
	}

	scaled := o.scaleImage(img)
	width, height := o.detWidth, o.detHeight
	data := o.detInputTensor.Data
	for y := range height {
		for x := range width {
			ir, ig, ib, _ := scaled.At(x, y).RGBA()
			r, g, b := float32(ir)/0xffff, float32(ig)/0xffff, float32(ib)/0xffff
			data[(0*height+y)*width+x] = (b - 0.485) / 0.229
			data[(1*height+y)*width+x] = (g - 0.456) / 0.224
			data[(2*height+y)*width+x] = (r - 0.406) / 0.225
		}
	}

	err = o.det.RunOnOutput(nil, map[string]ort.Value{inputName: o.detInputTensor}, map[string]ort.Value{outputName: o.detOutputTensor})
	if err != nil {
		return nil, err
	}

	contours := findContours(o.detOutputTensor.Data, o.detWidth, o.detHeight, img.Bounds())
	return contours, nil
}

func (o *OCR) Rec(img image.Image, contours []image.Rectangle) ([]*LabeledBox, error) {
	clipResizeAndNorm := func(contour image.Rectangle, maxWidth int, out []float32) {
		resizedWidth := min(int(math.Ceil(recDefaultHeight*ratioOf(contour))), maxWidth)

		scaled := image.NewNRGBA(image.Rect(0, 0, resizedWidth, recDefaultHeight))
		draw.BiLinear.Scale(scaled, scaled.Bounds(), img, contour, draw.Src, nil)

		for y := range recDefaultHeight {
			for x := range resizedWidth {
				ir, ig, ib, _ := scaled.At(x, y).RGBA()
				r, g, b := float32(ir)/0xffff, float32(ig)/0xffff, float32(ib)/0xffff
				out[(0*recDefaultHeight+y)*maxWidth+x] = (b - 0.5) * 2
				out[(1*recDefaultHeight+y)*maxWidth+x] = (g - 0.5) * 2
				out[(2*recDefaultHeight+y)*maxWidth+x] = (r - 0.5) * 2
			}
		}
	}

	results := []*LabeledBox{}

	for begin := 0; begin < len(contours); begin += recBatch {
		maxRatio := float64(recDefaultWidth) / float64(recDefaultHeight)
		end := min(len(contours), begin+recBatch)

		for i := begin; i < end; i++ {
			maxRatio = max(maxRatio, ratioOf(contours[i]))
		}

		maxWidth := int(math.Ceil(float64(recDefaultHeight) * maxRatio))
		chunkSize := 3 * recDefaultHeight * maxWidth
		chunks := end - begin
		recInputData := make([]float32, chunks*chunkSize)

		for i := begin; i < end; i++ {
			contour := contours[i]
			clipResizeAndNorm(contour, maxWidth, recInputData[chunkSize*(i-begin):])
		}

		recInputTensor, err := o.memInfo.NewTensorF32(recInputData, []int64{int64(chunks), 3, recDefaultHeight, int64(maxWidth)})
		if err != nil {
			return nil, err
		}

		recOutputs, err := o.rec.Run(nil, map[string]ort.Value{inputName: recInputTensor}, []string{outputName})
		if err != nil {
			return nil, err
		}

		recOutput, err := ort.AsTensor[float32](recOutputs[0])
		if err != nil {
			return nil, err
		}

		dims, err := recOutput.Dims()
		if err != nil {
			return nil, err
		}

		data := recOutput.Data

		ptr := 0
		for i := range int(dims[0]) {
			lastIdx := int64(-2)
			s := ""
			for range dims[1] {
				maxValue := float32(0)
				maxIdx := int64(0)
				for x := range dims[2] {
					if data[ptr] > maxValue {
						maxValue = data[ptr]
						maxIdx = x
					}
					ptr++
				}

				idx := maxIdx - 1
				if idx != lastIdx && idx >= 0 {
					s += chars[idx]
				}

				lastIdx = idx
			}

			if len(s) > 0 {
				results = append(results, &LabeledBox{
					Rect:  contours[begin+i],
					Label: s,
				})
			}
		}

		recOutput.Close()
	}

	return results, nil
}

func (o *OCR) Close() {
	if o.detInputTensor != nil {
		o.detInputTensor.Close()
	}

	if o.detOutputTensor != nil {
		o.detOutputTensor.Close()
	}

	o.rec.Close()
	o.det.Close()
	o.memInfo.Close()
	o.env.Close()
}

func NewOCR(identifier, detModelPath, recModelPath string) (*OCR, error) {
	env, err := ort.NewEnv(ort.LoggingLevelFatal, identifier)
	if err != nil {
		return nil, err
	}

	memInfo, err := ort.NewMemoryInfo(ort.DeviceCPU, ort.ArenaAllocator, 0, ort.MemTypeDefault)
	if err != nil {
		env.Close()
		return nil, err
	}

	det, err := env.NewSession(detModelPath, nil)
	if err != nil {
		env.Close()
		memInfo.Close()
		return nil, err
	}

	rec, err := env.NewSession(recModelPath, nil)
	if err != nil {
		env.Close()
		memInfo.Close()
		det.Close()
		return nil, err
	}

	return &OCR{
		env:     env,
		memInfo: memInfo,
		det:     det,
		rec:     rec,
	}, nil
}
