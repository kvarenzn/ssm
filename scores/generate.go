// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"cmp"
	"slices"

	"github.com/kvarenzn/ssm/common"
	"github.com/kvarenzn/ssm/log"
	"github.com/kvarenzn/ssm/utils"
)

func GenerateTouchEvent(config *VTEGenerateConfig, events []NoteEvent) common.RawVirtualEvents {
	// sort events by start time
	slices.SortFunc(events, func(a, b NoteEvent) int {
		return cmp.Compare(a.Start(), b.Start())
	})

	drags := []*DragEvent{}
	for _, ev := range events {
		if ev, ok := ev.(*DragEvent); ok {
			drags = append(drags, ev)
		}
	}
	if len(drags) > 0 {
		// ignore obscured drag events
		s := NewSLSF64()
		for _, ev := range events {
			switch ev := ev.(type) {
			case *TapEvent:
				s.AddTrace([]struct {
					T float64
					P float64
				}{
					{ev.Seconds, ev.Track},
					{ev.Seconds + float64(config.TapDuration)/1000, ev.Track},
				})
			case *DragEvent:
				s.AddQuery(ev.Seconds, &struct {
					Min float64
					Max float64
				}{ev.Track - ev.Width/2, ev.Track + ev.Width/2})
			case *FlickEvent:
				s.AddTrace([]struct {
					T float64
					P float64
				}{
					{ev.Seconds, ev.Track},
					{ev.Seconds + float64(config.FlickDuration)/1000, ev.Track + ev.Offset.X},
				})
			case *ThrowEvent:
				s.AddTrace([]struct {
					T float64
					P float64
				}{
					{ev.Seconds, ev.Track},
					{ev.Seconds + float64(config.FlickDuration)/1000, ev.Track + ev.Offset.X},
				})
			case *HoldEvent:
				s.AddTrace([]struct {
					T float64
					P float64
				}{
					{ev.Seconds, ev.Track},
					{ev.EndSeconds, ev.Track},
				})
			case *SlideEvent:
				trace := []struct{ T, P float64 }{
					{ev.Seconds, ev.Track},
				}
				for _, tr := range ev.Trace {
					trace = append(trace, struct {
						T float64
						P float64
					}{tr.Seconds, tr.Track})
				}
			}
		}
		obscured := s.Scan()
		for _, o := range obscured {
			drags[o.Query].ignored = true
		}

		// mark drags & throws that cannot be treated as tap or flick
		findNearByNote := func(idx int) bool {
			current := events[idx]
			var track float64
			switch current := current.(type) {
			case *DragEvent:
				track = current.Track
			case *ThrowEvent:
				track = current.Track
			}

			for i := idx + 1; i < len(events); i++ {
				ev := events[i]
				if ev.Start()-current.Start() > 0.125 {
					break
				}

				switch ev := ev.(type) {
				case *TapEvent:
					half := ev.Width / 2
					if ev.Track-half <= track && track <= ev.Track+half {
						return true
					}
				case *FlickEvent:
					half := ev.Width / 2
					if ev.Track-half <= track && track <= ev.Track+half {
						return true
					}
				case *HoldEvent:
					half := ev.Width / 2
					if ev.Track-half <= track && track <= ev.Track+half {
						return true
					}
				case *SlideEvent:
					half := ev.Width / 2
					if ev.Track-half <= track && track <= ev.Track+half {
						return true
					}
				}
			}

			return false
		}
		for i, ev := range events {
			switch ev := ev.(type) {
			case *DragEvent:
				if ev.ignored {
					continue
				}
				ev.doNotTap = findNearByNote(i)
			case *ThrowEvent:
				ev.doNotTap = findNearByNote(i)
			}
		}
	}

	// register events for allocation
	nodes := Nodes[int64]{}
	for _, event := range events {
		ms := quantify(event.Start())
		switch ev := event.(type) {
		case *TapEvent:
			nodes.AddEvent(ms, ms+config.TapDuration)
		case *FlickEvent:
			nodes.AddEvent(ms, ms+config.FlickDuration)
		case *HoldEvent:
			endMs := quantify(ev.EndSeconds)
			if !ev.FlickEnd {
				nodes.AddEvent(ms, endMs)
			} else {
				nodes.AddEvent(ms, endMs+config.FlickDuration)
			}
		case *SlideEvent:
			endMs := quantify(ev.Trace[len(ev.Trace)-1].Seconds)
			if !ev.FlickEnd {
				nodes.AddEvent(ms, endMs+1)
			} else {
				nodes.AddEvent(ms, endMs+config.FlickDuration)
			}
		}
	}

	// allocate!
	pointers := nodes.Colorize()

	// count how many pointers are used
	maxPtr := 0
	for _, ptr := range pointers {
		maxPtr = max(ptr, maxPtr)
	}
	log.Debugf("%d pointers used.", maxPtr+1)

	result := map[int64][]*common.VirtualTouchEvent{}
	addEvent := func(tick int64, event *common.VirtualTouchEvent) {
		_, ok := result[tick]
		if !ok {
			result[tick] = nil
		}
		result[tick] = append(result[tick], event)
	}
	for idx, event := range events {
		pointerID := pointers[idx]
		switch ev := event.(type) {
		case *TapEvent:
			ms := quantify(ev.Seconds)
			addEvent(ms, &common.VirtualTouchEvent{
				X:         ev.Track,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})
			addEvent(ms+int64(config.TapDuration), &common.VirtualTouchEvent{
				X:         ev.Track,
				Y:         0,
				Action:    common.TouchUp,
				PointerID: pointerID,
			})
		case *FlickEvent:
			offset := ev.Offset
			ms := quantify(ev.Seconds)
			addEvent(ms, &common.VirtualTouchEvent{
				X:         ev.Track,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})
			for i := ms + config.FlickReportInterval; i < ms+config.FlickDuration; i += config.FlickReportInterval {
				factor := float64(i-ms) / float64(config.FlickDuration)
				x := ev.Track + offset.X*factor
				y := offset.Y * factor
				addEvent(i, &common.VirtualTouchEvent{
					X:         x,
					Y:         y,
					Action:    common.TouchMove,
					PointerID: pointerID,
				})
			}
			addEvent(ms+config.FlickDuration, &common.VirtualTouchEvent{
				X:         ev.Track + offset.X,
				Y:         offset.Y,
				Action:    common.TouchUp,
				PointerID: pointerID,
			})
		case *HoldEvent:
			ms := quantify(ev.Seconds)
			endMs := quantify(ev.EndSeconds)
			addEvent(ms, &common.VirtualTouchEvent{
				X:         ev.Track,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})

			if !ev.FlickEnd {
				addEvent(endMs, &common.VirtualTouchEvent{
					X:         ev.Track,
					Y:         0,
					Action:    common.TouchUp,
					PointerID: pointerID,
				})
				continue
			}

			for i := endMs + int64(config.FlickReportInterval); i < endMs+int64(config.FlickDuration); i += int64(config.FlickReportInterval) {
				offsetY := float64(i-endMs) / float64(config.FlickDuration)
				addEvent(i, &common.VirtualTouchEvent{
					X:         ev.Track,
					Y:         offsetY,
					Action:    common.TouchMove,
					PointerID: pointerID,
				})
			}
			addEvent(endMs+int64(config.FlickDuration), &common.VirtualTouchEvent{
				X:         ev.Track,
				Y:         1,
				Action:    common.TouchUp,
				PointerID: pointerID,
			})
		case *SlideEvent:
			ms := quantify(ev.Seconds)
			xStart := ev.Track
			addEvent(ms, &common.VirtualTouchEvent{
				X:         ev.Track,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})

			for _, step := range ev.Trace {
				nextMs := quantify(step.Seconds)
				for i := ms + config.SlideReportInterval; i < nextMs; i += config.SlideReportInterval {
					currentX := xStart + (step.Track-xStart)*float64(i-ms)/float64(nextMs-ms)
					addEvent(i, &common.VirtualTouchEvent{
						X:         currentX,
						Y:         0,
						Action:    common.TouchMove,
						PointerID: pointerID,
					})
				}
				ms = nextMs
				xStart = step.Track
				addEvent(ms, &common.VirtualTouchEvent{
					X:         step.Track,
					Y:         0,
					Action:    common.TouchMove,
					PointerID: pointerID,
				})
			}

			if !ev.FlickEnd {
				addEvent(ms+1, &common.VirtualTouchEvent{
					X:         xStart,
					Y:         0,
					Action:    common.TouchUp,
					PointerID: pointerID,
				})
				continue
			}

			for i := ms + config.FlickReportInterval; i < ms+config.FlickDuration; i += config.FlickReportInterval {
				offsetY := float64(i-ms) / float64(config.FlickDuration)
				addEvent(i, &common.VirtualTouchEvent{
					X:         xStart,
					Y:         offsetY,
					Action:    common.TouchMove,
					PointerID: pointerID,
				})
			}
			addEvent(ms+config.FlickDuration, &common.VirtualTouchEvent{
				X:         xStart,
				Y:         1,
				Action:    common.TouchUp,
				PointerID: pointerID,
			})
		}
	}

	ticks := utils.SortedKeysOf(result)

	res := []*common.VirtualEventsItem{}
	for _, tick := range ticks {
		res = append(res, &common.VirtualEventsItem{
			Timestamp: tick,
			Events:    result[tick],
		})
	}

	return res
}
