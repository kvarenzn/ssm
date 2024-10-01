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

type columnInfo struct {
	column int
	points []Point2D
}

func RenderTrace(config GraphConfig, trace []PathItem, ctx *canvas.Context) {
	begin := trace[0]
	trace = trace[1:]
	base := 0.0
	column := 0
	columns := []*columnInfo{}
	for base+config.Duration < begin.Seconds {
		column++
		base += config.Duration
	}
	columns = append(columns, &columnInfo{
		column: column,
		points: []Point2D{
			{
				X: begin.OffsetX,
				Y: begin.Seconds - base,
			},
		},
	})

	for _, item := range trace {
		if item.Seconds <= base+config.Duration {
			ptr := columns[len(columns)-1]
			ptr.points = append(ptr.points, Point2D{
				X: item.OffsetX,
				Y: item.Seconds - base,
			})
		} else {
			for base+config.Duration < item.Seconds {
				base += config.Duration
				x := begin.OffsetX + (item.OffsetX-begin.OffsetX)*(base-begin.Seconds)/(item.Seconds-begin.Seconds)
				ptr := columns[len(columns)-1]
				ptr.points = append(ptr.points, Point2D{
					X: x,
					Y: config.Duration,
				})
				column++
				columns = append(columns, &columnInfo{
					column: column,
					points: []Point2D{
						{
							X: x,
							Y: 0,
						},
					},
				})
			}
			ptr := columns[len(columns)-1]
			ptr.points = append(ptr.points, Point2D{
				X: item.OffsetX,
				Y: item.Seconds - base,
			})
		}

		begin = item
	}

	hlw := config.TrackWidth / 14 // half of lane width
	for _, column := range columns {
		offset := float64(column.column)*(config.TrackPadding*2+config.TrackWidth) + config.TrackPadding + hlw
		p0 := column.points[0]
		column.points = column.points[1:]
		ctx.MoveTo(offset+p0.X*12*hlw-hlw, config.Height*(p0.Y/config.Duration))
		ctx.LineTo(offset+p0.X*12*hlw+hlw, config.Height*(p0.Y/config.Duration))
		for _, p := range column.points {
			ctx.LineTo(offset+p.X*12*hlw+hlw, config.Height*(p.Y/config.Duration))
		}

		for i := len(column.points) - 1; i >= 0; i-- {
			p := column.points[i]
			ctx.LineTo(offset+p.X*12*hlw-hlw, config.Height*(p.Y/config.Duration))
		}

		ctx.Close()
		ctx.Fill()
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
	tickFace := font.Face(16.0, canvas.Black, canvas.FontRegular, canvas.FontNormal)
	bpmFace := font.Face(16.0, canvas.Magenta, canvas.FontRegular, canvas.FontNormal)

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
	ctx.DrawText(p0.X-config.TrackPadding/7, config.Height-p0.Y, canvas.NewTextLine(bpmFace, fmt.Sprintf("%.2f", chart.Header.BPM), canvas.Right))
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
			ctx.DrawText(p0.X-config.TrackPadding/7, config.Height-p0.Y, canvas.NewTextLine(bpmFace, fmt.Sprintf("%.2f", bpmEvents[1].BPM), canvas.Right))
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
			RenderTrace(config, path, ctx)
		}
	}

	return c
}

func drawMain(chart Chart, events RawVirtualEvents, outPath string) error {
	alpha := uint8(0xff)
	graphConf := GraphConfig{
		Duration:     7.6,
		Height:       800,
		TrackWidth:   70,
		TrackPadding: 20,
		ColorsMap: map[int]color.Color{
			// mygo
			0: color.RGBA{0x77, 0xbb, 0xdd, alpha},
			1: color.RGBA{0xff, 0x88, 0x99, alpha},
			2: color.RGBA{0x77, 0xdd, 0x77, alpha},
			3: color.RGBA{0xff, 0xdd, 0x88, alpha},
			4: color.RGBA{0x77, 0x77, 0xaa, alpha},
			// ave mujica
			5: color.RGBA{0xbb, 0x99, 0x55, alpha},
			6: color.RGBA{0x77, 0x99, 0x77, alpha},
			7: color.RGBA{0x33, 0x55, 0x66, alpha},
			8: color.RGBA{0xaa, 0x44, 0x77, alpha},
			9: color.RGBA{0x77, 0x99, 0xcc, alpha},
		},
	}
	c := Draw(chart, events, graphConf)
	if err := renderers.Write(outPath, c); err != nil {
		return err
	}

	return nil
}
