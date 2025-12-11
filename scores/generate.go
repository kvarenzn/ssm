// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"cmp"
	"encoding/json"
	"math"
	"os"
	"slices"

	"github.com/kvarenzn/ssm/common"
	"github.com/kvarenzn/ssm/log"
	"github.com/kvarenzn/ssm/utils"
)

func GenerateTouchEvent(config *VTEGenerateConfig, events []*star) common.RawVirtualEvents {
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
			case flickNote, throwNote:
				dx, _ := ev.delta(config.FlickFactor)
				s.AddTrace([]struct {
					T float64
					P float64
				}{
					{ev.seconds, ev.track},
					{ev.seconds + float64(config.FlickDuration+config.FlickReportInterval)/1000, ev.track + dx},
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
		log.Debugf("%d tap(s), %d drag(s), %d throw(s)", tapCount, dragCount, throwCount)

		type exportedNote struct {
			Kind    noteKind `json:"kind"`
			Seconds float64  `json:"seconds"`
			Track   float64  `json:"track"`
			Width   float64  `json:"width"`
		}

		type exportedEdge struct {
			From      int  `json:"a"`
			To        int  `json:"b"`
			Connected bool `json:"connected"`
		}

		type exportedData struct {
			Notes []*exportedNote `json:"notes"`
			Edges []*exportedEdge `json:"edges"`
		}

		xport := &exportedData{}

		for _, n := range noteNodes {
			xport.Notes = append(xport.Notes, &exportedNote{
				Kind:    n.kind(),
				Seconds: n.start(),
				Track:   n.track,
				Width:   n.width,
			})
		}

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
			const connectBonus = 1e8
			const dropCost = connectBonus * 0.51
			const maxDistance = 1 // second(s)
			const kNeighbors = 10
			source := 0
			sink := noteNodeCount*2 + 1
			nodeCount := noteNodeCount*2 + 1 + 1 // every note has two nodes (in & out); plus a super Source and a super Sink
			fg := newFlowGraph(nodeCount)
			log.Debugf("%d node(s) in flow graph", nodeCount)

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

				// only drags & throws can connect before
				if s.kind() != dragNote && s.kind() != throwNote {
					continue
				}

				potentialNeighbors := []*struct {
					dist float64
					from *star
				}{}

				far := s.start() - maxDistance
				for p := startIdxs[s.start()] - 1; p >= 0 && lines[p][0].start() > far; p-- {
					for _, from := range lines[p] {
						// only taps & drags can accept connection from later
						if from.kind() != tapNote && from.kind() != dragNote {
							continue
						}

						dist := math.Hypot(from.start()-s.start(), (from.x()-s.x())*1)
						if dist >= maxDistance {
							continue
						}

						potentialNeighbors = append(potentialNeighbors, &struct {
							dist float64
							from *star
						}{dist, from})
					}
				}

				slices.SortFunc(potentialNeighbors, func(a, b *struct {
					dist float64
					from *star
				},
				) int {
					return cmp.Compare(a.dist, b.dist)
				})

				isBlocked := func(_, _ *star) bool {
					// [TODO] check whether some notes are between connection
					return false
				}

				for _, n := range potentialNeighbors[:min(len(potentialNeighbors), kNeighbors)] {
					if isBlocked(s, n.from) {
						continue
					}

					xport.Edges = append(xport.Edges, &exportedEdge{
						From:      noteIDMap[n.from],
						To:        noteIDMap[s],
						Connected: false,
					})
					fg.addEdge(outIDOf(noteIDMap[n.from]), inIDOf(noteIDMap[s]), 1, n.dist)
				}
			}

			log.Debugf("%d edge(s) in flow graph", fg.edgeCount)

			connections, maxFlow := fg.mc(source, sink)
			connections = slices.DeleteFunc(connections, func(conn *struct{ from, to int }) bool {
				return conn.from == source || conn.to == sink || conn.to-conn.from == noteNodeCount
			})

			for _, conn := range connections {
				if conn.from > noteNodeCount {
					conn.from -= noteNodeCount
				}
				conn.from--

				if conn.to > noteNodeCount {
					conn.to -= noteNodeCount
				}
				conn.to--
			}
			log.Debugf("maximum flow is %d", maxFlow)
			log.Debugf("%d connection(s)", len(connections))

			for _, conn := range connections {
				xport.Edges = append(xport.Edges, &exportedEdge{
					From:      conn.from,
					To:        conn.to,
					Connected: true,
				})
			}

			outData, err := json.MarshalIndent(xport, "", "    ")
			if err != nil {
				log.Dief("Failed to marshal outData: %s", err)
			}
			if err := os.WriteFile("out.json", outData, 0o644); err != nil {
				log.Dief("Failed to write outData: %s", err)
			}

			slices.SortFunc(connections, func(a, b *struct{ from, to int }) int {
				return cmp.Compare(a.from, b.from)
			})

			toBeDeleted.Clear()
			for _, conn := range connections {
				from := noteNodes[conn.from]
				to := noteNodes[conn.to]
				if !from.isSlide() {
					from.markAsHead()
				}
				to.chainsAfter(from)
				toBeDeleted.Add(from)
			}
			log.Debugf("delete %d note(s)", toBeDeleted.Len())

			// delete chained notes
			events = slices.DeleteFunc(events, func(e *star) bool {
				return toBeDeleted.Contains(e)
			})

			// sort all events again
			slices.SortFunc(events, func(a, b *star) int {
				return cmp.Compare(a.start(), b.start())
			})
		}
	}

	// register events for allocation
	nodes := NewCloves[int64]()
	for id, event := range events {
		ms := quantify(event.start())
		switch event.kind() {
		case tapNote, dragNote:
			nodes.AddEvent(id, ms, ms+config.TapDuration)
		case flickNote, throwNote:
			nodes.AddEvent(id, ms, ms+config.FlickDuration+config.FlickReportInterval)
		case slideNote:
			endMs := quantify(event.seconds)
			if !event.isFlick() {
				nodes.AddEvent(id, ms, endMs+1)
			} else {
				nodes.AddEvent(id, ms, endMs+config.FlickDuration+config.FlickReportInterval)
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

	addFlickTail := func(event *star, pointerID int, ms int64, xs float64) {
		dx, dy := event.delta(config.FlickFactor)
		factor := 1.0 / math.Pow(float64(config.FlickDuration), config.FlickPow)
		for i := config.FlickReportInterval; i <= config.FlickDuration; i += config.FlickReportInterval {
			rate := factor * math.Pow(float64(i), config.FlickPow)
			addEvent(i+ms, &common.VirtualTouchEvent{
				X:         xs + dx*rate,
				Y:         dy * rate,
				Action:    common.TouchMove,
				PointerID: pointerID,
			})
		}
		addEvent(ms+config.FlickDuration+config.FlickReportInterval, &common.VirtualTouchEvent{
			X:         xs + dx,
			Y:         dy,
			Action:    common.TouchUp,
			PointerID: pointerID,
		})
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
			ms := quantify(event.seconds)
			addEvent(ms, &common.VirtualTouchEvent{
				X:         event.track,
				Y:         0,
				Action:    common.TouchDown,
				PointerID: pointerID,
			})
			addFlickTail(event, pointerID, ms, event.track)
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

			addFlickTail(event, pointerID, ms, xStart)
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
