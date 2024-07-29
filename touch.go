package main

import (
	"math"
	"sort"
)

type VTEGenerateConfig struct {
	TapDuration         int64
	FlickDuration       int64
	FlickReportInterval int64
	SlideReportInterval int64
}

func NewVTEGenerateConfig() VTEGenerateConfig {
	return VTEGenerateConfig{
		TapDuration:         5,
		FlickDuration:       60,
		FlickReportInterval: 5,
		SlideReportInterval: 10,
	}
}

func TrackIDToX(trackID float64) float64 {
	return trackID / 6
}

func GenerateTouchEvent(config VTEGenerateConfig, chart Chart) RawVirtualEvents {
	result := map[int64][]VirtualTouchEvent{}
	addEvent := func(tick int64, event VirtualTouchEvent) {
		_, ok := result[tick]
		if !ok {
			result[tick] = []VirtualTouchEvent{}
		}
		result[tick] = append(result[tick], event)
	}
	events := chart.Events
	sort.Slice(events, func(i, j int) bool {
		return events[i].Start() < events[j].Start()
	})

	for _, event := range events {
		switch ev := event.(type) {
		case TapEvent:
			ms := int64(math.Round(ev.Seconds * 1000))
			offsetX := TrackIDToX(float64(ev.TrackID))
			pointerID := ev.TrackID
			addEvent(ms, VirtualTouchEvent{
				X:         offsetX,
				Y:         0,
				Action:    TouchDown,
				PointerID: pointerID,
			})
			addEvent(ms+int64(config.TapDuration), VirtualTouchEvent{
				X:         offsetX,
				Y:         0,
				Action:    TouchUp,
				PointerID: pointerID,
			})
		case FlickEvent:
			offset := ev.Offset
			ms := int64(math.Round(ev.Seconds * 1000))
			offsetX := TrackIDToX(float64(ev.TrackID))
			pointerID := ev.TrackID
			addEvent(ms, VirtualTouchEvent{
				X:         offsetX,
				Y:         0,
				Action:    TouchDown,
				PointerID: pointerID,
			})
			for i := ms + config.FlickReportInterval; i < ms+config.FlickDuration; i += config.FlickReportInterval {
				factor := float64(i-ms) / float64(config.FlickDuration)
				x := offsetX + offset[0]*factor
				y := offset[1] * factor
				addEvent(i, VirtualTouchEvent{
					X:         x,
					Y:         y,
					Action:    TouchMove,
					PointerID: pointerID,
				})
			}
			addEvent(ms+config.FlickDuration, VirtualTouchEvent{
				X:         offsetX + offset[0],
				Y:         offset[1],
				Action:    TouchUp,
				PointerID: pointerID,
			})
		case HoldEvent:
			ms := int64(math.Round(ev.Seconds * 1000))
			endMs := int64(math.Round(ev.EndSeconds * 1000))
			pointerID := ev.TrackID
			offsetX := TrackIDToX(float64(ev.TrackID))
			addEvent(ms, VirtualTouchEvent{
				X:         offsetX,
				Y:         0,
				Action:    TouchDown,
				PointerID: pointerID,
			})

			if !ev.FlickEnd {
				addEvent(endMs, VirtualTouchEvent{
					X:         offsetX,
					Y:         0,
					Action:    TouchUp,
					PointerID: pointerID,
				})
				continue
			}

			for i := endMs + int64(config.FlickReportInterval); i < endMs+int64(config.FlickDuration); i += int64(config.FlickReportInterval) {
				offsetY := float64(i-endMs) / float64(config.FlickDuration)
				addEvent(i, VirtualTouchEvent{
					X:         offsetX,
					Y:         offsetY,
					Action:    TouchMove,
					PointerID: pointerID,
				})
			}
			addEvent(endMs+int64(config.FlickDuration), VirtualTouchEvent{
				X:         offsetX,
				Y:         1,
				Action:    TouchUp,
				PointerID: pointerID,
			})
		case SlideEvent:
			pointerID := 8
			if ev.Mark == "b" {
				pointerID = 9
			}
			ms := int64(math.Round(ev.Seconds * 1000))
			trackID := ev.Track
			offsetX := TrackIDToX(trackID)
			addEvent(ms, VirtualTouchEvent{
				X:         offsetX,
				Y:         0,
				Action:    TouchDown,
				PointerID: pointerID,
			})

			for _, step := range ev.Trace {
				nextMs := int64(math.Round(step[0] * 1000))
				for i := ms + config.SlideReportInterval; i < nextMs; i += config.SlideReportInterval {
					currentTrack := trackID + (step[1]-trackID)*float64(i-ms)/float64(nextMs-ms)
					offsetX = TrackIDToX(currentTrack)
					addEvent(i, VirtualTouchEvent{
						X:         offsetX,
						Y:         0,
						Action:    TouchMove,
						PointerID: pointerID,
					})
				}
				ms = nextMs
				trackID = step[1]
				offsetX = TrackIDToX(trackID)
				addEvent(ms, VirtualTouchEvent{
					X:         offsetX,
					Y:         0,
					Action:    TouchMove,
					PointerID: pointerID,
				})
			}

			if !ev.FlickEnd {
				addEvent(ms+1, VirtualTouchEvent{
					X:         offsetX,
					Y:         0,
					Action:    TouchUp,
					PointerID: pointerID,
				})
				continue
			}

			for i := ms + config.FlickReportInterval; i < ms+config.FlickDuration; i += config.FlickReportInterval {
				offsetY := float64(i-ms) / float64(config.FlickDuration)
				addEvent(i, VirtualTouchEvent{
					X:         offsetX,
					Y:         offsetY,
					Action:    TouchMove,
					PointerID: pointerID,
				})
			}
			addEvent(ms+config.FlickDuration, VirtualTouchEvent{
				X:         offsetX,
				Y:         1,
				Action:    TouchUp,
				PointerID: pointerID,
			})
		}
	}

	ticks := getKeys(result)
	sort.Slice(ticks, func(i, j int) bool {
		return ticks[i] < ticks[j]
	})

	res := []VirtualEventsItem{}
	for _, tick := range ticks {
		res = append(res, VirtualEventsItem{
			Timestamp: tick,
			Events:    result[tick],
		})
	}

	return res
}
