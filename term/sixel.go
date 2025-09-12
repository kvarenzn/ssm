// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package term

import (
	"bufio"
	"container/heap"
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strings"
)

type WuQuantizer struct {
	side, square, cube, bins int
	counts                   []float64
	sumR                     []float64
	sumG                     []float64
	sumB                     []float64
	sumSq                    []float64
}

func NewWuQuantizer(bins int) *WuQuantizer {
	side := bins + 1
	cube := side * side * side
	return &WuQuantizer{
		side:   side,
		square: side * side,
		cube:   cube,
		bins:   bins,
		counts: make([]float64, cube),
		sumR:   make([]float64, cube),
		sumG:   make([]float64, cube),
		sumB:   make([]float64, cube),
		sumSq:  make([]float64, cube),
	}
}

func (w *WuQuantizer) idx(r, g, b int) int {
	return (r*w.side+g)*w.side + b
}

func (w *WuQuantizer) buildHistogram(img image.Image) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r32, g32, b32, _ := img.At(x, y).RGBA()
			r := int(r32 >> 8)
			g := int(g32 >> 8)
			b := int(b32 >> 8)

			ir := (r >> 3) + 1
			ig := (g >> 3) + 1
			ib := (b >> 3) + 1
			id := w.idx(ir, ig, ib)
			w.counts[id]++
			w.sumR[id] += float64(r)
			w.sumG[id] += float64(g)
			w.sumB[id] += float64(b)
			w.sumSq[id] += float64(r*r + g*g + b*b)
		}
	}
}

func (w *WuQuantizer) buildMoments() {
	for r := 1; r <= w.bins; r++ {
		for g := 1; g <= w.bins; g++ {
			for b := 1; b <= w.bins; b++ {
				i := w.idx(r, g, b)
				i1 := w.idx(r-1, g, b)
				i2 := w.idx(r, g-1, b)
				i3 := w.idx(r, g, b-1)
				i4 := w.idx(r-1, g-1, b)
				i5 := w.idx(r-1, g, b-1)
				i6 := w.idx(r, g-1, b-1)
				i7 := w.idx(r-1, g-1, b-1)

				w.counts[i] += w.counts[i1] + w.counts[i2] + w.counts[i3] - w.counts[i4] - w.counts[i5] - w.counts[i6] + w.counts[i7]
				w.sumR[i] += w.sumR[i1] + w.sumR[i2] + w.sumR[i3] - w.sumR[i4] - w.sumR[i5] - w.sumR[i6] + w.sumR[i7]
				w.sumG[i] += w.sumG[i1] + w.sumG[i2] + w.sumG[i3] - w.sumG[i4] - w.sumG[i5] - w.sumG[i6] + w.sumG[i7]
				w.sumB[i] += w.sumB[i1] + w.sumB[i2] + w.sumB[i3] - w.sumB[i4] - w.sumB[i5] - w.sumB[i6] + w.sumB[i7]
				w.sumSq[i] += w.sumSq[i1] + w.sumSq[i2] + w.sumSq[i3] - w.sumSq[i4] - w.sumSq[i5] - w.sumSq[i6] + w.sumSq[i7]
			}
		}
	}
}

func (w *WuQuantizer) integrate(box *Box, moment []float64) float64 {
	r0, r1 := box.r0, box.r1
	g0, g1 := box.g0, box.g1
	b0, b1 := box.b0, box.b1
	return moment[w.idx(r1, g1, b1)] -
		moment[w.idx(r0-1, g1, b1)] -
		moment[w.idx(r1, g0-1, b1)] +
		moment[w.idx(r0-1, g0-1, b1)] -
		moment[w.idx(r1, g1, b0-1)] +
		moment[w.idx(r0-1, g1, b0-1)] +
		moment[w.idx(r1, g0-1, b0-1)] -
		moment[w.idx(r0-1, g0-1, b0-1)]
}

func (w *WuQuantizer) volumeCount(box *Box) float64 {
	return w.integrate(box, w.counts)
}

func (w *WuQuantizer) volumeSumR(box *Box) float64 {
	return w.integrate(box, w.sumR)
}

func (w *WuQuantizer) volumeSumG(box *Box) float64 {
	return w.integrate(box, w.sumG)
}

func (w *WuQuantizer) volumeSumB(box *Box) float64 {
	return w.integrate(box, w.sumB)
}

func (w *WuQuantizer) volumeSumSq(box *Box) float64 {
	return w.integrate(box, w.sumSq)
}

func (w *WuQuantizer) variance(box *Box) float64 {
	n := w.volumeCount(box)
	if n <= 0 {
		return 0.0
	}
	sumR := w.volumeSumR(box)
	sumG := w.volumeSumG(box)
	sumB := w.volumeSumB(box)
	sumSq := w.volumeSumSq(box)
	return sumSq - (sumR*sumR+sumG*sumG+sumB*sumB)/n
}

type Box struct {
	r0, r1 int
	g0, g1 int
	b0, b1 int
}

func (b *Box) String() string {
	return fmt.Sprintf("Box((%02d, %02d, %02d) -> (%02d, %02d, %02d))", b.r0, b.g0, b.b0, b.r1, b.g1, b.b1)
}

func splitBox(box *Box, dir int, at int, box1 *Box, box2 *Box) {
	c := *box
	box1.r0, box1.r1, box1.g0, box1.g1, box1.b0, box1.b1 = c.r0, c.r1, c.g0, c.g1, c.b0, c.b1
	box2.r0, box2.r1, box2.g0, box2.g1, box2.b0, box2.b1 = c.r0, c.r1, c.g0, c.g1, c.b0, c.b1
	switch dir {
	case 0:
		box1.r1 = at
		box2.r0 = at + 1
	case 1:
		box1.g1 = at
		box2.g0 = at + 1
	case 2:
		box1.b1 = at
		box2.b0 = at + 1
	default:
		panic("?")
	}
}

func (w *WuQuantizer) maximize(box *Box, dir int) (bestDiff float64, bestCut int) {
	bestDiff = 0
	bestCut = -1

	var s, e int
	switch dir {
	case 0: // R
		s, e = box.r0, box.r1
	case 1: // G
		s, e = box.g0, box.g1
	case 2: // B
		s, e = box.b0, box.b1
	}

	box1 := new(Box)
	box2 := new(Box)
	for i := s; i < e; i++ {
		splitBox(box, dir, i, box1, box2)
		wt1 := w.volumeCount(box1)
		wt2 := w.volumeCount(box2)
		if wt1 == 0 || wt2 == 0 {
			continue
		}

		diff := w.variance(box) - (w.variance(box1) + w.variance(box2))
		if diff > bestDiff {
			bestDiff = diff
			bestCut = i
		}
	}

	return
}

func (w *WuQuantizer) cut(box *Box, newBox *Box) bool {
	var dir int
	cut := -1
	maxDiff := -1.0

	for i := range 3 {
		diff, c := w.maximize(box, i)
		if c >= 0 && diff > maxDiff {
			dir = i
			maxDiff = diff
			cut = c
		}
	}

	if cut < 0 {
		return false
	}

	splitBox(box, dir, cut, box, newBox)
	return true
}

func (w *WuQuantizer) averageColor(box *Box) color.NRGBA {
	n := w.volumeCount(box)
	if n <= 0 {
		return color.NRGBA{0, 0, 0, 255}
	}
	r := w.volumeSumR(box) / n
	g := w.volumeSumG(box) / n
	b := w.volumeSumB(box) / n

	ri := uint8(max(0, min(255, math.Round(r))))
	gi := uint8(max(0, min(255, math.Round(g))))
	bi := uint8(max(0, min(255, math.Round(b))))
	return color.NRGBA{ri, gi, bi, 255}
}

type PQItem struct {
	box   *Box
	score float64
	idx   int
}
type PriorityQueue []*PQItem

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].score > pq[j].score
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].idx = i
	pq[j].idx = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*PQItem)
	item.idx = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.idx = -1
	*pq = old[0 : n-1]
	return item
}

func (w *WuQuantizer) Quantize(colors int) ([]color.NRGBA, []*Box) {
	boxes := make([]*Box, 0, colors)
	initial := &Box{r0: 1, r1: w.bins, g0: 1, g1: w.bins, b0: 1, b1: w.bins}
	boxes = append(boxes, initial)

	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &PQItem{box: initial, score: w.variance(initial)})

	for len(boxes) < colors && pq.Len() > 0 {
		item := heap.Pop(pq).(*PQItem)
		box := item.box
		newBox := &Box{}
		ok := w.cut(box, newBox)
		if !ok {
			heap.Push(pq, item)
			break
		}
		boxes = append(boxes, newBox)
		heap.Push(pq, &PQItem{box: box, score: w.variance(box)})
		heap.Push(pq, &PQItem{box: newBox, score: w.variance(newBox)})
	}

	finalBoxes := make([]*Box, 0, colors)
	for pq.Len() > 0 && len(finalBoxes) < colors {
		item := heap.Pop(pq).(*PQItem)
		finalBoxes = append(finalBoxes, item.box)
	}
	for _, b := range boxes {
		if len(finalBoxes) >= colors {
			break
		}
		dup := slices.Contains(finalBoxes, b)
		if !dup {
			finalBoxes = append(finalBoxes, b)
		}
	}

	palette := []color.NRGBA{}
	for _, b := range finalBoxes {
		palette = append(palette, w.averageColor(b))
	}
	return palette, finalBoxes
}

func (w *WuQuantizer) mapImageToPalette(img image.Image, boxes []*Box, palette []color.NRGBA) []int {
	bounds := img.Bounds()
	out := make([]int, bounds.Dx()*bounds.Dy())

	lookup := make([]int, w.cube)
	for i := range lookup {
		lookup[i] = -1
	}
	for pi, b := range boxes {
		for r := b.r0; r <= b.r1; r++ {
			for g := b.g0; g <= b.g1; g++ {
				for bb := b.b0; bb <= b.b1; bb++ {
					lookup[w.idx(r, g, bb)] = pi
				}
			}
		}
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r32, g32, b32, _ := img.At(x, y).RGBA()
			r := int(r32 >> 8)
			g := int(g32 >> 8)
			b := int(b32 >> 8)
			ir := (r >> 3) + 1
			ig := (g >> 3) + 1
			ib := (b >> 3) + 1
			pi := lookup[w.idx(ir, ig, ib)]
			if pi < 0 {
				best := 0
				bestDist := math.MaxFloat64
				for i, c := range palette {
					dr := float64(int(c.R) - r)
					dg := float64(int(c.G) - g)
					db := float64(int(c.B) - b)
					d := dr*dr + dg*dg + db*db
					if d < bestDist {
						bestDist = d
						best = i
					}
				}
				pi = best
			}
			out[(y-bounds.Min.Y)*bounds.Dx()+(x-bounds.Min.X)] = pi
		}
	}
	return out
}

func sixelOutput(buffer *bufio.Writer, bounds image.Rectangle, palette []color.NRGBA, img []int) {
	// into sixel mode
	fmt.Fprintf(buffer, "\x1bP0;1;8q\"1;1;%d;%d", bounds.Dx(), bounds.Dy())

	// set palette
	for i, c := range palette {
		fmt.Fprintf(buffer, "#%d;2;%d;%d;%d", i, int(c.R)*100/0xff, int(c.G)*100/0xff, int(c.B)*100/0xff)
	}

	buffer.WriteByte('\n')

	sz := bounds.Dx() * bounds.Dy()

	for y := bounds.Min.Y; y < bounds.Max.Y; y += 6 {
		masks := map[int][]byte{}
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			v := byte(1)
			for dy := range 6 {
				i := (y+dy-bounds.Min.Y)*bounds.Dx() + (x - bounds.Min.X)
				var idx int
				if i < sz {
					idx = img[i]
				}

				if _, ok := masks[idx]; !ok {
					masks[idx] = nil
				}

				for len(masks[idx]) < x+1 {
					masks[idx] = append(masks[idx], 0)
				}

				masks[idx][x] |= v
				v <<= 1
			}
		}

		cnt := 0
		n := len(masks)
		for idx, mask := range masks {
			fmt.Fprintf(buffer, "#%d", idx)
			last := byte(255)
			count := 0
			for _, m := range mask {
				if m != last {
					if last != 255 {
						if count > 3 {
							fmt.Fprintf(buffer, "!%d%c", count, rune('?'+last))
						} else {
							buffer.WriteString(strings.Repeat(string(rune('?'+last)), count))
						}
					}

					last = m
					count = 0
				}
				count++
			}

			if count > 0 {
				if count > 3 {
					fmt.Fprintf(buffer, "!%d%c", count, rune('?'+last))
				} else {
					buffer.WriteString(strings.Repeat(string(rune('?'+last)), count))
				}
			}

			if cnt < n-1 {
				buffer.WriteByte('$')
			} else {
				buffer.WriteByte('-')
			}
			buffer.WriteByte('\n')

			cnt++
		}
	}

	// exit sixel mode
	buffer.WriteString("\x1b\\\n")
}
