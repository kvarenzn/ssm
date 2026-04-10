// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kvarenzn/ssm/utils"
)

const (
	susTap            = '1'
	susCritical       = '2'
	susFlick          = '3'
	susDamage         = '4'
	susTrend          = '5'
	susCriticalTrend  = '6'
	susCancel         = '7'
	susCriticalCancel = '8'

	susAirUp         = '1'
	susAirDown       = '2'
	susAirUpperLeft  = '3'
	susAirUpperRight = '4'
	susAirLowerLeft  = '5'
	susAirLowerRight = '6'

	susSlideBegin         = '1'
	susSlideEnd           = '2'
	susSlideStepVisible   = '3'
	susSlideStepInvisible = '5'
)

func susDegOf(t uint8) int {
	switch t {
	case susAirDown:
		return -90
	case susAirLowerLeft:
		return -135
	case susAirLowerRight:
		return -45
	case susAirUpperLeft:
		return 135
	case susAirUpperRight:
		return 45
	case susAirUp:
		fallthrough
	default:
		return 90
	}
}

func hexToInt(r byte) (int, error) {
	if r >= '0' && r <= '9' {
		return int(r - '0'), nil
	} else if r >= 'a' && r <= 'g' {
		return int(r-'a') + 10, nil
	} else if r >= 'A' && r <= 'G' {
		return int(r-'A') + 10, nil
	} else {
		return 0, fmt.Errorf("unknown char: %s", string(rune(r)))
	}
}

type susRawNoteEvent struct {
	kind       uint8
	width      int
	lane       int
	identifier uint8
	consumed   bool
}

const susLaneGaps = 11

func (n *susRawNoteEvent) track() float64 {
	return (float64(n.lane-2) + float64(n.width-1)/2) / susLaneGaps
}

type susEventsPack struct {
	barLength float64
	bpm       float64
	shorts    []*susRawNoteEvent
	airs      []*susRawNoteEvent
	slides    []*susRawNoteEvent
	trails    []*susRawNoteEvent
}

func polylineX(list []*vec2f, y float64) float64 {
	if y <= list[0].y {
		return list[0].x
	} else if y >= list[len(list)-1].y {
		return list[len(list)-1].x
	}

	for i, e := range list[1:] {
		s := list[i]
		if s.y <= y && y <= e.y {
			return s.x + (e.x-s.x)*(y-s.y)/(e.y-s.y)
		}
	}

	return list[len(list)-1].x
}

func ParseSUS(chartText string) (Chart, error) {
	bpms := map[string]float64{}

	collectedEvents := map[float64]*susEventsPack{}

	for line := range strings.Lines(chartText) {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}

		var key, value string

		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			spaceIndex := strings.Index(line, " ")
			if spaceIndex != -1 {
				key = line[1:spaceIndex] // skip '#'
				value = line[spaceIndex+1:]
			} else {
				key = line[1:]
			}
		} else {
			key = line[1:colonIndex]
			value = line[colonIndex+1:]
		}

		value = strings.TrimSpace(value)
		if strings.HasPrefix(value, "\"") {
			value = value[1 : len(value)-1]
		}

		switch key {
		case "TITLE":
		case "ARTIST":
		case "DESIGNER":
		case "DIFFICULTY":
		case "PLAYLEVEL":
		case "SONGID":
		case "WAVE":
		case "WAVEOFFSET":
		case "JACKET":
		case "REQUEST":
			if strings.HasPrefix(value, "ticks_per_beat") {
				value = value[15:] // skip "ticks_per_beat "
				value = strings.TrimSpace(value)
				i, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse ticks_per_beat: %s", err)
				}

				_ = int(i)
			}
		case "VOLUME":
		case "HISPEED":
		case "MEASUREHS":
		case "TIL00":
		default:
			if strings.HasPrefix(key, "BPM") {
				index := key[3:]

				bpm, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse BPM list item value `%s`: %s", value, err)
				}

				bpms[index] = bpm
				continue
			} else if len(key) == 5 && strings.HasSuffix(key, "02") {
				index, err := strconv.ParseInt(key[:3], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse time signature list item key `%s`: %s", key, err)
				}

				sig, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse time signature list item value `%s`: %s", value, err)
				}

				tick := float64(index)
				if _, ok := collectedEvents[tick]; !ok {
					collectedEvents[tick] = &susEventsPack{}
				}

				collectedEvents[tick].barLength = sig
				continue
			}

			events, common, err := parseDataLine(line)
			if err != nil {
				return nil, err
			}

			if common.Channel == "08" {
				for _, ev := range events {
					if bpm, ok := bpms[ev.Type]; !ok {
						return nil, fmt.Errorf("Invalid BPM index `%s`", ev.Type)
					} else {
						tick := ev.Tick()
						if _, ok := collectedEvents[tick]; !ok {
							collectedEvents[tick] = &susEventsPack{}
						}

						if collectedEvents[tick].bpm != 0.0 {
							return nil, fmt.Errorf("Duplicated BPM event at tick %f", tick)
						}

						collectedEvents[tick].bpm = bpm
					}
				}
				continue
			}

			laneID, err := hexToInt(common.Channel[1])
			if err != nil {
				return nil, err
			}

			if laneID < 2 || laneID > 13 {
				// skip special / control lanes
				continue
			}

			var identifier uint8
			if len(common.Channel) == 3 {
				identifier = common.Channel[2]
			}

			for _, event := range events {
				tick := event.Tick()
				if _, ok := collectedEvents[tick]; !ok {
					collectedEvents[tick] = &susEventsPack{}
				}

				width, err := hexToInt(event.Type[1])
				if err != nil {
					return nil, fmt.Errorf("Unknown width char: %s", string(event.Type[1]))
				}

				note := &susRawNoteEvent{
					kind:       event.Type[0],
					width:      width,
					lane:       laneID,
					identifier: identifier,
				}
				p := collectedEvents[tick]
				noteType := common.Channel[0]
				switch noteType {
				case '1': // taps
					p.shorts = append(p.shorts, note)
				case '2': // holds
				case '3': // slides
					p.slides = append(p.slides, note)
				case '4': // air actions
				case '5': // air
					p.airs = append(p.airs, note)
				case '9': // decorated slides
					p.trails = append(p.trails, note)
				default:
					return nil, fmt.Errorf("Unknown type: %s", string(rune(common.Channel[0])))
				}
			}
		}
	}

	ticks := utils.SortedKeysOf(collectedEvents)
	secStart := 0.0
	tickStart := 0.0
	slides := map[uint8]*star{}
	slideDirections := map[uint8]uint8{}

	chainSlide := func(id uint8, secs, track, width float64) {
		const epsilon = 0.0007

		prev := slides[id]
		easeIn, easeOut := false, false
		switch slideDirections[id] {
		case susAirDown:
			easeIn = true
		case susAirLowerLeft, susAirLowerRight:
			easeOut = true
		default: // no ease
			slides[id] = newStar(secs, track, width).
				chainsAfter(prev)
			return
		}

		var l0, l1, l2, l3, r0, r1, r2, r3 *vec2f
		l0 = newVec2f(prev.track-prev.width/2, prev.seconds)
		if easeIn {
			l1 = newVec2f(prev.track-prev.width/2, (prev.seconds+secs)/2)
		} else {
			l1 = newVec2f(prev.track-prev.width/2, prev.seconds)
		}
		if easeOut {
			l2 = newVec2f(track-width/2, (prev.seconds+secs)/2)
		} else {
			l2 = newVec2f(track-width/2, secs)
		}
		l3 = newVec2f(track-width/2, secs)

		r0 = newVec2f(prev.track+prev.width/2, prev.seconds)
		if easeIn {
			r1 = newVec2f(prev.track+prev.width/2, (prev.seconds+secs)/2)
		} else {
			r1 = newVec2f(prev.track+prev.width/2, prev.seconds)
		}
		if easeOut {
			r2 = newVec2f(track+width/2, (prev.seconds+secs)/2)
		} else {
			r2 = newVec2f(track+width/2, secs)
		}
		r3 = newVec2f(track+width/2, secs)

		left := bezierToPolyline(l0, l1, l2, l3, epsilon)
		right := bezierToPolyline(r0, r1, r2, r3, epsilon)
		ys := map[float64]struct{}{}
		for _, p := range left {
			ys[p.y] = struct{}{}
		}
		for _, p := range right {
			ys[p.y] = struct{}{}
		}
		for _, y := range utils.SortedKeysOf(ys)[1:] {
			xl := polylineX(left, y)
			xr := polylineX(right, y)
			slides[id] = newStar(y, (xl+xr)/2, xr-xl).
				chainsAfter(slides[id])
		}
	}

	bpm := 120.0
	finalEvents := []*star{}
	barLength := 4.0
	for _, tick := range ticks {
		pack := collectedEvents[tick]

		if pack.bpm != 0.0 {
			secPerTick := barLength * 60 / bpm
			secStart += (tick - tickStart) * secPerTick
			tickStart = tick
			bpm = pack.bpm
		}

		if pack.barLength != 0.0 {
			barLength = pack.barLength
		}

		secPerTick := barLength * 60 / bpm
		secs := secStart + (tick-tickStart)*secPerTick

		for _, n := range pack.slides {
			switch n.kind {
			case susSlideBegin:
				// + tap + air -> slide with ease
				// + critical -> critical slide (ignored)
				for _, s := range pack.shorts {
					if s.consumed || s.lane != n.lane || s.width != n.width {
						continue
					}

					s.consumed = true
				}

				direction := uint8(0)
				for _, a := range pack.airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					a.consumed = true
					direction = a.kind
				}

				if _, ok := slides[n.identifier]; ok {
					return nil, fmt.Errorf("Duplicated slide begin with same identifier: %s", string(n.identifier))
				}

				slides[n.identifier] = newStar(
					secs,
					n.track(),
					float64(n.width)/susLaneGaps,
				).
					markAsHead().
					markAsTap()
				slideDirections[n.identifier] = direction
			case susSlideEnd:
				// + air -> slide with flick end
				// + critical -> critical slide end (ignored)
				for _, s := range pack.shorts {
					if s.consumed || s.lane != n.lane || s.width != n.width {
						continue
					}

					s.consumed = true
				}

				flickEnd := false
				for _, a := range pack.airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					flickEnd = true
					a.consumed = true
				}

				if _, ok := slides[n.identifier]; !ok {
					return nil, fmt.Errorf("Slide begin with identifier %s not found", string(n.identifier))
				}

				chainSlide(
					n.identifier,
					secs,
					n.track(),
					float64(n.width)/susLaneGaps)
				finalEvents = append(finalEvents,
					slides[n.identifier].
						flickToIfOk(flickEnd, 90).
						markAsEnd())
				delete(slides, n.identifier)
				delete(slideDirections, n.identifier)
			case susSlideStepInvisible:
				// + tap + air -> slide with ease
				// + critical -> critical slide (ignored)
				for _, s := range pack.shorts {
					if s.consumed || s.lane != n.lane || s.width != n.width {
						continue
					}

					s.consumed = true
				}

				direction := uint8(0)
				for _, a := range pack.airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					a.consumed = true
					direction = a.kind
				}

				if _, ok := slides[n.identifier]; !ok {
					return nil, fmt.Errorf("Slide begin with identifier %s not found", string(n.identifier))
				}

				chainSlide(
					n.identifier,
					secs,
					n.track(),
					float64(n.width)/susLaneGaps)
				slideDirections[n.identifier] = direction
			case susSlideStepVisible:
				// + flick -> any position mid
				// + tap + air -> slide with ease
				// + critical -> critical slide (ignored)
				ignorePosition := false
				for _, s := range pack.shorts {
					if s.consumed || s.lane != n.lane || s.width != n.width {
						continue
					}

					s.consumed = true
					if s.kind == susFlick {
						ignorePosition = true
					}
				}

				direction := uint8(0)
				for _, a := range pack.airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					a.consumed = true
					direction = a.kind
				}

				if _, ok := slides[n.identifier]; !ok {
					return nil, fmt.Errorf("Slide begin with identifier %s not found", string(n.identifier))
				}

				if !ignorePosition {
					chainSlide(
						n.identifier,
						secs,
						n.track(),
						float64(n.width)/susLaneGaps)
					slideDirections[n.identifier] = direction
				}
			}
		}

		for _, n := range pack.shorts {
			if n.consumed {
				continue
			}

			switch n.kind {
			case susTap, susCritical:
				// + air -> flick
				flick := false
				var flickType uint8
				for _, a := range pack.airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					a.consumed = true
					flick = true
					flickType = a.kind
				}

				finalEvents = append(finalEvents,
					newStar(
						secs,
						n.track(),
						float64(n.width)/susLaneGaps,
					).
						markAsTap().
						flickToIfOk(flick, susDegOf(flickType)))
			case susTrend, susCriticalTrend:
				// + air -> throw
				flick := false
				var flickType uint8
				for _, a := range pack.airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					a.consumed = true
					flick = true
					flickType = a.kind
				}

				finalEvents = append(finalEvents,
					newStar(
						secs,
						n.track(),
						float64(n.width)/susLaneGaps,
					).
						flickToIfOk(flick, susDegOf(flickType)))
			case susCancel, susCriticalCancel:
				// [TODO] WTF is this?
			case susDamage:
			}
		}
	}

	return finalEvents, nil
}
