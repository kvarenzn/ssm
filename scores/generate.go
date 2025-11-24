// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"cmp"
	"math"
	"slices"

	"github.com/kvarenzn/ssm/common"
	"github.com/kvarenzn/ssm/log"
	"github.com/kvarenzn/ssm/utils"
)

func GenerateTouchEvent(config *VTEGenerateConfig, events []*star) common.RawVirtualEvents {
	const flickFactor = 1.0 / 3
	// sort events by start time
	slices.SortFunc(events, func(a, b *star) int {
		return cmp.Compare(a.start(), b.start())
	})

	drags := []*star{}
	for _, ev := range events {
		if ev.kind() == dragNote {
			drags = append(drags, ev)
		}
	}
	if len(drags) > 0 {
		// ignore obscured drag events
		s := NewSLSF64()
		for _, ev := range events {
			switch ev.kind() {
			case tapNote:
				s.AddTrace([]struct {
					T float64
					P float64
				}{
					{ev.seconds, ev.track},
					{ev.seconds + float64(config.TapDuration)/1000, ev.track},
				})
			case dragNote:
				s.AddQuery(ev.seconds, &struct {
					Min float64
					Max float64
				}{ev.track - ev.width/2, ev.track + ev.width/2})
			case flickNote:
				dx, _ := ev.delta(flickFactor)
				s.AddTrace([]struct {
					T float64
					P float64
				}{
					{ev.seconds, ev.track},
					{ev.seconds + float64(config.FlickDuration)/1000, ev.track + dx},
				})
			case throwNote:
				dx, _ := ev.delta(flickFactor)
				s.AddTrace([]struct {
					T float64
					P float64
				}{
					{ev.seconds, ev.track},
					{ev.seconds + float64(config.FlickDuration)/1000, ev.track + dx},
				})
			case slideNote:
				trace := []struct{ T, P float64 }{}
				for t := range ev.iterSlide() {
					trace = append(trace, struct {
						T float64
						P float64
					}{t.seconds, t.track})
				}

				s.AddTrace(trace)
			}
		}

		toBeDeleted := utils.NewSet[*star]()
		obscured := s.Scan()
		for _, o := range obscured {
			toBeDeleted.Add(drags[o.Query])
		}

		log.Debugf("%d drag(s) obscured", len(obscured))

		// delete obscured drags from events
		events = slices.DeleteFunc(events, func(e *star) bool {
			return toBeDeleted.Contains(e)
		})

		// mark drags & throws that cannot be treated as tap or flick
		isThisCannotTap := func(idx int) bool {
			current := events[idx]
			var track float64
			switch current.kind() {
			case dragNote:
				track = current.track
			case throwNote:
				track = current.track
			}

			for i := idx + 1; i < len(events); i++ {
				ev := events[i]
				if ev.start()-current.start() > 0.125 {
					break
				}

				switch ev.kind() {
				case tapNote:
					half := ev.width / 2
					if ev.track-half <= track && track <= ev.track+half {
						return true
					}
				case flickNote:
					half := ev.width / 2
					if ev.track-half <= track && track <= ev.track+half {
						return true
					}
				case slideNote:
					head := ev.head
					half := head.width / 2
					if head.track-half <= track && track <= head.track+half {
						return true
					}
				}
			}

			return false
		}

		noteNodeCount := 0
		noteNodes := []*star{}
		noteIDMap := map[*star]int{}
		doNotTap := utils.NewSet[*star]()
		noteMap := map[float64][]*star{}
		lines := [][]*star{}
		var tapCount, dragCount, throwCount int
		for i, s := range events {
			start := s.start()
			switch s.kind() {
			case tapNote:
				noteNodes = append(noteNodes, s)
				noteIDMap[s] = noteNodeCount
				noteNodeCount++
				tapCount++

				if _, ok := noteMap[start]; !ok {
					noteMap[start] = nil
				}
				noteMap[start] = append(noteMap[start], s)
			case dragNote:
				noteNodes = append(noteNodes, s)
				noteIDMap[s] = noteNodeCount
				noteNodeCount++
				dragCount++

				if isThisCannotTap(i) {
					doNotTap.Add(s)
				}

				if _, ok := noteMap[start]; !ok {
					noteMap[start] = nil
				}
				noteMap[start] = append(noteMap[start], s)
			case throwNote:
				noteNodes = append(noteNodes, s)
				noteIDMap[s] = noteNodeCount
				noteNodeCount++
				throwCount++
				if isThisCannotTap(i) {
					doNotTap.Add(s)
				}

				if _, ok := noteMap[start]; !ok {
					noteMap[start] = nil
				}
				noteMap[start] = append(noteMap[start], s)
			}
		}
		log.Debugf("%d taps, %d drags, %d throws", tapCount, dragCount, throwCount)

		startIdxs := map[float64]int{}
		for i, k := range utils.SortedKeysOf(noteMap) {
			startIdxs[k] = i
			notes := noteMap[k]
			slices.SortFunc(notes, func(a, b *star) int {
				return cmp.Compare(a.x(), b.x())
			})
			lines = append(lines, notes)
		}

		{
			const connectBonus = 1e9
			const dropCost = connectBonus * 0.51
			const maxDistance = 0.5 // second(s)
			const kNeighbors = 10
			source := 0
			sink := noteNodeCount*2 + 1
			nodeCount := noteNodeCount*2 + 1 + 1 // every note has two nodes (in & out); plus a super Source and a super Sink
			fg := newFlowGraph(nodeCount)
			log.Debugf("%d nodes in flow graph", nodeCount)

			inIDOf := func(i int) int {
				return i + 1
			}

			outIDOf := func(i int) int {
				return noteNodeCount + inIDOf(i)
			}

			for i, s := range noteNodes {
				switch s.kind() {
				case tapNote:
					fg.addEdge(source, inIDOf(i), 1, 0)
					fg.addEdge(inIDOf(i), outIDOf(i), 1, 0)
				case dragNote:
					fg.addEdge(source, inIDOf(i), 1, dropCost)
					fg.addEdge(inIDOf(i), outIDOf(i), 1, -connectBonus)
					fg.addEdge(outIDOf(i), sink, 1, dropCost)
				case throwNote:
					fg.addEdge(inIDOf(i), outIDOf(i), 1, -connectBonus)
					fg.addEdge(outIDOf(i), sink, 1, dropCost)
				}

				if s.kind() != tapNote && s.kind() != dragNote {
					continue
				}

				potentialNeighbors := []*struct {
					dist float64
					to   *star
				}{}

				far := s.start() + maxDistance
				for p := startIdxs[s.start()] + 1; p < len(lines) && lines[p][0].start() < far; p++ {
					for _, n := range lines[p] {
						if n.kind() != dragNote && n.kind() != throwNote {
							continue
						}

						dd := math.Hypot(n.start()-s.start(), n.x()-s.x())
						if dd >= maxDistance*maxDistance {
							potentialNeighbors = append(potentialNeighbors, &struct {
								dist float64
								to   *star
							}{dd, n})
						}
					}
				}

				slices.SortFunc(potentialNeighbors, func(a, b *struct {
					dist float64
					to   *star
				},
				) int {
					return cmp.Compare(a.dist, b.dist)
				})

				isBlocked := func(_, _ *star) bool {
					// [TODO] check whether some notes are between connection
					return false
				}

				for _, n := range potentialNeighbors[:min(len(potentialNeighbors), kNeighbors)] {
					if isBlocked(s, n.to) {
						continue
					}

					fg.addEdge(outIDOf(noteIDMap[s]), inIDOf(noteIDMap[n.to]), 1, n.dist)
				}
			}

			log.Debugf("%d edges in flow graph", fg.edgeCount)

			connections, maxFlow := fg.mcmf(source, sink)
			log.Debugf("maxFlow = %d", maxFlow)
			log.Debugf("%d connections", len(connections))
		}
	}

	// register events for allocation
	nodes := NewCloves[int64]()
	for id, event := range events {
		ms := quantify(event.start())
		switch event.kind() {
		case tapNote:
			nodes.AddEvent(id, ms, ms+config.TapDuration)
		case flickNote:
			nodes.AddEvent(id, ms, ms+config.FlickDuration)
		case dragNote:
			nodes.AddEvent(id, ms, ms+config.TapDuration)
		case throwNote:
			nodes.AddEvent(id, ms, ms+config.FlickDuration)
		case slideNote:
			endMs := quantify(event.seconds)
			if !event.isFlick() {
				nodes.AddEvent(id, ms, endMs+1)
			} else {
				nodes.AddEvent(id, ms, endMs+config.FlickDuration)
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
		switch event.kind() {
		case tapNote:
			ms := quantify(event.seconds)
			addEvent(ms, &common.VirtualTouchEvent{
				X:         event.track,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})
			addEvent(ms+int64(config.TapDuration), &common.VirtualTouchEvent{
				X:         event.track,
				Y:         0,
				Action:    common.TouchUp,
				PointerID: pointerID,
			})
		case dragNote:
			ms := quantify(event.seconds)
			addEvent(ms, &common.VirtualTouchEvent{
				X:         event.track,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})
			addEvent(ms+int64(config.TapDuration), &common.VirtualTouchEvent{
				X:         event.track,
				Y:         0,
				Action:    common.TouchUp,
				PointerID: pointerID,
			})
		case throwNote, flickNote:
			dx, dy := event.delta(flickFactor)
			ms := quantify(event.seconds)
			addEvent(ms, &common.VirtualTouchEvent{
				X:         event.track,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})
			for i := ms + config.FlickReportInterval; i < ms+config.FlickDuration; i += config.FlickReportInterval {
				factor := float64(i-ms) / float64(config.FlickDuration)
				x := event.track + dx*factor
				y := dy * factor
				addEvent(i, &common.VirtualTouchEvent{
					X:         x,
					Y:         y,
					Action:    common.TouchMove,
					PointerID: pointerID,
				})
			}
			addEvent(ms+config.FlickDuration, &common.VirtualTouchEvent{
				X:         event.track + dx,
				Y:         dy,
				Action:    common.TouchUp,
				PointerID: pointerID,
			})
		case slideNote:
			var ms int64
			var xStart float64

			first := true
			for step := range event.iterSlide() {
				if first {
					ms = quantify(step.seconds)
					xStart = step.track
					addEvent(ms, &common.VirtualTouchEvent{
						X:         step.track,
						Y:         0,
						Action:    common.TouchDown,
						PointerID: pointerID,
					})
					first = false
					continue
				}

				nextMs := quantify(step.seconds)
				for i := ms + config.SlideReportInterval; i < nextMs; i += config.SlideReportInterval {
					factor := float64(i-ms) / float64(nextMs-ms)
					currentX := xStart + (step.track-xStart)*factor
					addEvent(i, &common.VirtualTouchEvent{
						X:         currentX,
						Y:         0,
						Action:    common.TouchMove,
						PointerID: pointerID,
					})
				}
				ms = nextMs
				xStart = step.track
				addEvent(ms, &common.VirtualTouchEvent{
					X:         step.track,
					Y:         0,
					Action:    common.TouchMove,
					PointerID: pointerID,
				})
			}

			if !event.isFlick() {
				addEvent(ms+1, &common.VirtualTouchEvent{
					X:         xStart,
					Y:         0,
					Action:    common.TouchUp,
					PointerID: pointerID,
				})
				continue
			}

			dx, dy := event.delta(flickFactor)
			for i := ms + config.FlickReportInterval; i < ms+config.FlickDuration; i += config.FlickReportInterval {
				factor := float64(i-ms) / float64(config.FlickDuration)
				addEvent(i, &common.VirtualTouchEvent{
					X:         xStart + dx*factor,
					Y:         dy * factor,
					Action:    common.TouchMove,
					PointerID: pointerID,
				})
			}
			addEvent(ms+config.FlickDuration, &common.VirtualTouchEvent{
				X:         xStart + dx,
				Y:         dy,
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
