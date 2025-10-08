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

func scaleImage(i image.Image) image.Image {
	bounds := i.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	maxSide := max(width, height)
	scaleFactor := min(detMaxLength/float64(maxSide), 1)

	newWidth := max(int(math.Round(float64(width)*scaleFactor/blockSize)), 1) * blockSize
	newHeight := max(int(math.Round(float64(height)*scaleFactor/blockSize)), 1) * blockSize

	if newWidth != width || newHeight != height {
		inew := image.NewNRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.BiLinear.Scale(inew, inew.Bounds(), i, i.Bounds(), draw.Src, nil)
		return inew
	} else {
		return i
	}
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

func prepareForDet(i image.Image) ([]float32, []int64) {
	scaled := scaleImage(i)
	bounds := scaled.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	area := bounds.Dx() * bounds.Dy()
	data := make([]float32, area*3)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			ir, ig, ib, _ := scaled.At(x, y).RGBA()
			r, g, b := float32(ir)/0xffff, float32(ig)/0xffff, float32(ib)/0xffff
			data[(0*height+y)*width+x] = (b - 0.485) / 0.229
			data[(1*height+y)*width+x] = (g - 0.456) / 0.224
			data[(2*height+y)*width+x] = (r - 0.406) / 0.225
		}
	}

	return data, []int64{1, 3, int64(height), int64(width)}
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
}

func (o *OCR) Close() {
	o.rec.Close()
	o.det.Close()
	o.memInfo.Close()
	o.env.Close()
}

func (o *OCR) RunOnImage(img image.Image) ([]*LabeledBox, error) {
	detInputData, detInputShape := prepareForDet(img)
	detInputTensor, err := o.memInfo.NewTensorF32(detInputData, detInputShape)
	if err != nil {
		return nil, err
	}
	defer detInputTensor.Close()

	outputs, err := o.det.Run(nil, map[string]*ort.Tensor{inputName: detInputTensor}, []string{outputName})
	if err != nil {
		return nil, err
	}

	output := outputs[0]
	defer output.Close()

	dims, err := output.Dims()
	if err != nil {
		return nil, err
	}

	outputData, err := ort.GetTensorData[float32](output)
	if err != nil {
		return nil, err
	}

	contours := findContours(outputData, int(dims[3]), int(dims[2]), img.Bounds())

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

		recOutputs, err := o.rec.Run(nil, map[string]*ort.Tensor{inputName: recInputTensor}, []string{outputName})
		if err != nil {
			return nil, err
		}

		recOutput := recOutputs[0]

		dims, err := recOutput.Dims()
		if err != nil {
			return nil, err
		}

		data, err := ort.GetTensorData[float32](recOutput)
		if err != nil {
			return nil, err
		}

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

func NewOCR(name, detModelPath, recModelPath string) (*OCR, error) {
	env, err := ort.NewEnv(ort.LoggingLevelFatal, "t7")
	if err != nil {
		return nil, err
	}

	memInfo, err := ort.NewMemoryInfo("Cpu", ort.ArenaAllocator, 0, ort.MemTypeDefault)
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
