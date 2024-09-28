package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

type GraphConfig struct {
	Duration     float64 // unit: seconds
	Height       float64
	TrackWidth   float64
	TrackPadding float64
	ColorsMap    map[int]color.Color
}

type Parallelograms struct {
	Height float64
	Offset float64
	Points []Point2D
}

func NewParallelograms(width, height float64) Parallelograms {
	return Parallelograms{
		Height: height,
		Offset: width / 2,
		Points: []Point2D{},
	}
}

func (pgs *Parallelograms) AddPoint(p Point2D) {
	pgs.Points = append(pgs.Points, p)
}

func (pgs *Parallelograms) drawPgs(points []Point2D, ctx *canvas.Context) {
	if len(points) <= 0 {
		return
	}

	p0 := points[0]
	points = points[1:]

	// left points
	ctx.MoveTo(p0.X-pgs.Offset, pgs.Height-p0.Y)

	for _, p := range points {
		ctx.LineTo(p.X-pgs.Offset, pgs.Height-p.Y)
	}

	for i := len(points) - 1; i >= 0; i-- {
		p := points[i]
		ctx.LineTo(p.X+pgs.Offset, pgs.Height-p.Y)
	}

	ctx.LineTo(p0.X+pgs.Offset, pgs.Height-p0.Y)
	ctx.Close()
	ctx.Fill()
}

func (pgs *Parallelograms) Render(ctx *canvas.Context) {
	// break paths
	paths := [][]Point2D{{pgs.Points[0]}}
	pgs.Points = pgs.Points[1:]

	for _, p := range pgs.Points {
		lastPath := paths[len(paths)-1]
		lastPoint := lastPath[len(lastPath)-1]
		if p.Y <= lastPoint.Y {
			paths[len(paths)-1] = append(paths[len(paths)-1], p)
			continue
		}

		xb := lastPoint.X + (p.X-lastPoint.X)*lastPoint.Y/(lastPoint.Y+pgs.Height-p.X)
		lastPath = append(lastPath, Point2D{xb, 0})
		paths = append(paths, []Point2D{{xb, pgs.Height}})
	}

	for _, path := range paths {
		pgs.drawPgs(path, ctx)
	}
}

type PathItem struct {
	Seconds float64
	OffsetX float64
}

func Draw(chart Chart, events RawVirtualEvents, config GraphConfig) *canvas.Canvas {
	font := canvas.NewFontFamily("fira")
	if err := font.LoadSystemFont("Fira Code", canvas.FontRegular); err != nil {
		panic(err)
	}
	tickFace := font.Face(14.0, canvas.Black, canvas.FontRegular, canvas.FontNormal)
	bpmFace := font.Face(14.0, canvas.Magenta, canvas.FontRegular, canvas.FontNormal)

	posOf := func(secs float64) Point2D {
		column := math.Floor(secs / config.Duration)
		height := (secs - column*config.Duration) / config.Duration * config.Height
		return Point2D{column*(config.TrackWidth+config.TrackPadding*2) + config.TrackPadding, config.Height - height}
	}

	columnsCount := int(math.Ceil(float64(events[len(events)-1].Timestamp) / 1000 / config.Duration))
	width := float64(columnsCount) * (config.TrackWidth + config.TrackPadding*2)

	c := canvas.New(width, config.Height)
	ctx := canvas.NewContext(c)

	// draw vertical lines
	ctx.SetStrokeColor(color.Black)
	leftMost := config.TrackPadding
	for i := 0; i < columnsCount; i++ {
		ctx.SetStrokeWidth(2)
		ctx.MoveTo(leftMost, 0)
		ctx.LineTo(leftMost, config.Height)
		ctx.Stroke()
		ctx.MoveTo(leftMost+config.TrackWidth, 0)
		ctx.LineTo(leftMost+config.TrackWidth, config.Height)
		ctx.Stroke()

		// split lines
		ctx.SetStrokeWidth(0.5)
		for j := 1; j <= 6; j++ {
			x := leftMost + float64(j)*config.TrackWidth/7
			ctx.MoveTo(x, 0)
			ctx.LineTo(x, config.Height)
			ctx.Stroke()
		}

		leftMost += config.TrackPadding*2 + config.TrackWidth
	}

	// draw horizontal lines
	bpmEvents := append([]BPMEvent{{0, chart.Header.BPM}}, chart.BPMEvents...)
	secStart := 0.0
	p0 := posOf(0)
	// first bpm line
	ctx.SetStrokeWidth(1)
	ctx.SetStrokeColor(canvas.Magenta)
	ctx.SetDashes(0, config.TrackWidth/15)
	ctx.MoveTo(p0.X, config.Height-p0.Y)
	ctx.LineTo(p0.X+config.TrackWidth, config.Height-p0.Y)
	ctx.Stroke()
	ctx.SetStrokeColor(color.Black)
	ctx.DrawText(p0.X-config.TrackPadding/5, config.Height-p0.Y, canvas.NewTextLine(bpmFace, fmt.Sprintf("%f", chart.Header.BPM), canvas.Right))
	for qTick := 1; ; qTick++ {
		tick := float64(qTick) / 4
		if len(bpmEvents) > 1 && tick >= bpmEvents[1].Tick {
			secStart += 240.0 / bpmEvents[0].BPM * (bpmEvents[1].Tick - bpmEvents[0].Tick)
			// bpm line
			p0 = posOf(secStart)
			ctx.SetStrokeWidth(1)
			ctx.SetStrokeColor(canvas.Magenta)
			ctx.SetDashes(0, config.TrackWidth/15)
			ctx.MoveTo(p0.X, config.Height-p0.Y)
			ctx.LineTo(p0.X+config.TrackWidth, config.Height-p0.Y)
			ctx.Stroke()
			ctx.SetStrokeColor(color.Black)
			ctx.DrawText(p0.X-config.TrackPadding/5, config.Height-p0.Y, canvas.NewTextLine(bpmFace, fmt.Sprintf("%f", bpmEvents[1].BPM), canvas.Right))
			bpmEvents = bpmEvents[1:]
		}

		if qTick%4 == 0 {
			ctx.SetStrokeWidth(1)
			ctx.SetDashes(0)
		} else {
			ctx.SetStrokeWidth(0.5)
			ctx.SetDashes(0, config.TrackWidth/29)
		}

		secs := secStart + 240.0/bpmEvents[0].BPM*(tick-bpmEvents[0].Tick)
		if secs >= float64(columnsCount)*config.Duration {
			break
		}
		p := posOf(secs)
		ctx.MoveTo(p.X, config.Height-p.Y)
		ctx.LineTo(p.X+config.TrackWidth, config.Height-p.Y)
		ctx.Stroke()

		if qTick%4 == 0 {
			ctx.DrawText(p.X+config.TrackWidth+config.TrackPadding/5, config.Height-p.Y, canvas.NewTextLine(tickFace, fmt.Sprintf("%d", qTick/4), canvas.Left))
		}
	}

	// draw traces
	traces := map[int][][]PathItem{}
	poses := map[int]*Point2D{}

	for _, item := range events {
		tick := item.Timestamp
		for _, event := range item.Events {
			switch event.Action {
			case TouchDown:
				if poses[event.PointerID] != nil {
					panic("failed")
				}
				traces[event.PointerID] = append(traces[event.PointerID], []PathItem{
					{float64(tick) / 1000, event.X},
				})
				poses[event.PointerID] = &Point2D{event.X, event.Y}
			case TouchMove:
				if poses[event.PointerID] == nil {
					panic("failed")
				}
				ptr := traces[event.PointerID]
				ptr[len(ptr)-1] = append(ptr[len(ptr)-1], PathItem{float64(tick) / 1000, event.X})
				poses[event.PointerID] = &Point2D{event.X, event.Y}
			case TouchUp:
				if poses[event.PointerID] == nil {
					panic("failed")
				}
				ptr := traces[event.PointerID]
				ptr[len(ptr)-1] = append(ptr[len(ptr)-1], PathItem{float64(tick) / 1000, event.X})
				poses[event.PointerID] = nil
			}
		}
	}

	ctx.SetStrokeColor(color.Transparent)
	for ptrID, paths := range traces {
		ctx.SetFillColor(config.ColorsMap[ptrID])
		for _, path := range paths {
			p0 := path[0]
			path = path[1:]
			p := posOf(p0.Seconds)
			p.X += ((p0.OffsetX * 6 / 7) + 1.0/14) * config.TrackWidth
			pgs := NewParallelograms(config.TrackWidth/7, config.Height)
			pgs.AddPoint(p)

			for _, item := range path {
				p := posOf(item.Seconds)
				p.X += ((item.OffsetX * 6 / 7) + 1.0/14) * config.TrackWidth
				pgs.AddPoint(p)
			}

			pgs.Render(ctx)
		}
	}

	return c
}

func drawMain(chart Chart, events RawVirtualEvents, outPath string) error {
	graphConf := GraphConfig{
		Duration:     7.6,
		Height:       800,
		TrackWidth:   70,
		TrackPadding: 20,
		ColorsMap: map[int]color.Color{
			// tracks
			0: color.RGBA{0xff, 0xff, 0xff, 0xbb},
			1: color.RGBA{0xff, 0x00, 0x00, 0xbb},
			2: color.RGBA{0xff, 0xff, 0x00, 0xbb},
			3: color.RGBA{0x00, 0xff, 0x00, 0xbb},
			4: color.RGBA{0x00, 0xff, 0xff, 0xbb},
			5: color.RGBA{0x00, 0x00, 0xff, 0xbb},
			6: color.RGBA{0xff, 0x00, 0xff, 0xbb},
		},
	}
	c := Draw(chart, events, graphConf)
	if err := renderers.Write(outPath, c); err != nil {
		return err
	}

	return nil
}
