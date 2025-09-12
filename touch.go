// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"math"
	"slices"
	"sort"

	"github.com/kvarenzn/ssm/common"
	"github.com/kvarenzn/ssm/log"
)

type VTEGenerateConfig struct {
	TapDuration         int64
	FlickDuration       int64
	FlickReportInterval int64
	SlideReportInterval int64
}

func TrackIDToX(trackID float64) float64 {
	return trackID / 6
}

func GenerateTouchEvent(config VTEGenerateConfig, chart Chart) common.RawVirtualEvents {
	result := map[int64][]common.VirtualTouchEvent{}
	addEvent := func(tick int64, event common.VirtualTouchEvent) {
		_, ok := result[tick]
		if !ok {
			result[tick] = []common.VirtualTouchEvent{}
		}
		result[tick] = append(result[tick], event)
	}
	events := chart.NoteEvents
	sort.Slice(events, func(i, j int) bool {
		return events[i].Start() < events[j].Start()
	})

	nodes := Nodes[int64]{}
	for _, event := range events {
		switch ev := event.(type) {
		case TapEvent:
			ms := int64(math.Round(ev.Seconds * 1000))
			nodes.AddEvent(ms, ms+config.TapDuration)
		case FlickEvent:
			ms := int64(math.Round(ev.Seconds * 1000))
			nodes.AddEvent(ms, ms+config.FlickDuration)
		case HoldEvent:
			ms := int64(math.Round(ev.Seconds * 1000))
			endMs := int64(math.Round(ev.EndSeconds * 1000))
			if !ev.FlickEnd {
				nodes.AddEvent(ms, endMs)
			} else {
				nodes.AddEvent(ms, endMs+config.FlickDuration)
			}
		case SlideEvent:
			ms := int64(math.Round(ev.Seconds * 1000))
			endMs := int64(math.Round(ev.Trace[len(ev.Trace)-1].Tick * 1000))
			if !ev.FlickEnd {
				nodes.AddEvent(ms, endMs+1)
			} else {
				nodes.AddEvent(ms, endMs+config.FlickDuration)
			}
		}
	}
	pointers := nodes.Colorize()
	maxPtr := 0
	for _, ptr := range pointers {
		maxPtr = max(ptr, maxPtr)
	}
	log.Debugln(maxPtr+1, "pointers used.")

	for idx, event := range events {
		pointerID := pointers[idx]
		switch ev := event.(type) {
		case TapEvent:
			ms := int64(math.Round(ev.Seconds * 1000))
			offsetX := TrackIDToX(float64(ev.TrackID))
			addEvent(ms, common.VirtualTouchEvent{
				X:         offsetX,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})
			addEvent(ms+int64(config.TapDuration), common.VirtualTouchEvent{
				X:         offsetX,
				Y:         0,
				Action:    common.TouchUp,
				PointerID: pointerID,
			})
		case FlickEvent:
			offset := ev.Offset
			ms := int64(math.Round(ev.Seconds * 1000))
			offsetX := TrackIDToX(float64(ev.TrackID))
			addEvent(ms, common.VirtualTouchEvent{
				X:         offsetX,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})
			for i := ms + config.FlickReportInterval; i < ms+config.FlickDuration; i += config.FlickReportInterval {
				factor := float64(i-ms) / float64(config.FlickDuration)
				x := offsetX + offset.X*factor
				y := offset.Y * factor
				addEvent(i, common.VirtualTouchEvent{
					X:         x,
					Y:         y,
					Action:    common.TouchMove,
					PointerID: pointerID,
				})
			}
			addEvent(ms+config.FlickDuration, common.VirtualTouchEvent{
				X:         offsetX + offset.X,
				Y:         offset.Y,
				Action:    common.TouchUp,
				PointerID: pointerID,
			})
		case HoldEvent:
			ms := int64(math.Round(ev.Seconds * 1000))
			endMs := int64(math.Round(ev.EndSeconds * 1000))
			offsetX := TrackIDToX(float64(ev.TrackID))
			addEvent(ms, common.VirtualTouchEvent{
				X:         offsetX,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})

			if !ev.FlickEnd {
				addEvent(endMs, common.VirtualTouchEvent{
					X:         offsetX,
					Y:         0,
					Action:    common.TouchUp,
					PointerID: pointerID,
				})
				continue
			}

			for i := endMs + int64(config.FlickReportInterval); i < endMs+int64(config.FlickDuration); i += int64(config.FlickReportInterval) {
				offsetY := float64(i-endMs) / float64(config.FlickDuration)
				addEvent(i, common.VirtualTouchEvent{
					X:         offsetX,
					Y:         offsetY,
					Action:    common.TouchMove,
					PointerID: pointerID,
				})
			}
			addEvent(endMs+int64(config.FlickDuration), common.VirtualTouchEvent{
				X:         offsetX,
				Y:         1,
				Action:    common.TouchUp,
				PointerID: pointerID,
			})
		case SlideEvent:
			ms := int64(math.Round(ev.Seconds * 1000))
			trackID := ev.Track
			offsetX := TrackIDToX(trackID)
			addEvent(ms, common.VirtualTouchEvent{
				X:         offsetX,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})

			for _, step := range ev.Trace {
				nextMs := int64(math.Round(step.Tick * 1000))
				for i := ms + config.SlideReportInterval; i < nextMs; i += config.SlideReportInterval {
					currentTrack := trackID + (step.Track-trackID)*float64(i-ms)/float64(nextMs-ms)
					offsetX = TrackIDToX(currentTrack)
					addEvent(i, common.VirtualTouchEvent{
						X:         offsetX,
						Y:         0,
						Action:    common.TouchMove,
						PointerID: pointerID,
					})
				}
				ms = nextMs
				trackID = step.Track
				offsetX = TrackIDToX(trackID)
				addEvent(ms, common.VirtualTouchEvent{
					X:         offsetX,
					Y:         0,
					Action:    common.TouchMove,
					PointerID: pointerID,
				})
			}

			if !ev.FlickEnd {
				addEvent(ms+1, common.VirtualTouchEvent{
					X:         offsetX,
					Y:         0,
					Action:    common.TouchUp,
					PointerID: pointerID,
				})
				continue
			}

			for i := ms + config.FlickReportInterval; i < ms+config.FlickDuration; i += config.FlickReportInterval {
				offsetY := float64(i-ms) / float64(config.FlickDuration)
				addEvent(i, common.VirtualTouchEvent{
					X:         offsetX,
					Y:         offsetY,
					Action:    common.TouchMove,
					PointerID: pointerID,
				})
			}
			addEvent(ms+config.FlickDuration, common.VirtualTouchEvent{
				X:         offsetX,
				Y:         1,
				Action:    common.TouchUp,
				PointerID: pointerID,
			})
		}
	}

	ticks := getKeys(result)
	slices.Sort(ticks)

	res := []common.VirtualEventsItem{}
	for _, tick := range ticks {
		res = append(res, common.VirtualEventsItem{
			Timestamp: tick,
			Events:    result[tick],
		})
	}

	return res
}
