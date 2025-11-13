// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kvarenzn/ssm/common"
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

func susOffsetOf(t uint8) common.Point2D {
	const dx = 1.0 / 5
	switch t {
	case susAirDown:
		return common.Point2D{X: 0, Y: -1}
	case susAirLowerLeft:
		return common.Point2D{X: -dx, Y: -1}
	case susAirLowerRight:
		return common.Point2D{X: dx, Y: -1}
	case susAirUpperLeft:
		return common.Point2D{X: -dx, Y: 1}
	case susAirUpperRight:
		return common.Point2D{X: dx, Y: 1}
	case susAirUp:
		fallthrough
	default:
		return common.Point2D{X: 0, Y: 1}
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
	BarLength float64
	BPM       float64
	Shorts    []*susRawNoteEvent
	Airs      []*susRawNoteEvent
	Slides    []*susRawNoteEvent
	Trails    []*susRawNoteEvent
}

func ParseSUS(chartText string) ([]NoteEvent, error) {
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

				collectedEvents[tick].BarLength = sig
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

						if collectedEvents[tick].BPM != 0.0 {
							return nil, fmt.Errorf("Duplicated BPM event at tick %f", tick)
						}

						collectedEvents[tick].BPM = bpm
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
					p.Shorts = append(p.Shorts, note)
				case '2': // holds
				case '3': // slides
					p.Slides = append(p.Slides, note)
				case '4': // air actions
				case '5': // air
					p.Airs = append(p.Airs, note)
				case '9': // decorated slides
					p.Trails = append(p.Trails, note)
				default:
					return nil, fmt.Errorf("Unknown type: %s", string(rune(common.Channel[0])))
				}
			}
		}
	}

	ticks := utils.SortedKeysOf(collectedEvents)
	secStart := 0.0
	tickStart := 0.0
	slides := map[uint8]*SlideEvent{}
	bpm := 120.0
	finalEvents := []NoteEvent{}
	barLength := 4.0
	for _, tick := range ticks {
		pack := collectedEvents[tick]

		if pack.BPM != 0.0 {
			secPerTick := barLength * 60 / bpm
			secStart += (tick - tickStart) * secPerTick
			tickStart = tick
			bpm = pack.BPM
		}

		if pack.BarLength != 0.0 {
			barLength = pack.BarLength
		}

		secPerTick := barLength * 60 / bpm
		secs := secStart + (tick-tickStart)*secPerTick

		for _, n := range pack.Slides {
			switch n.kind {
			case susSlideBegin:
				// + critical -> critical slide (ignored)
				// + tap + air -> slide with ease (ignored)
				for _, s := range pack.Shorts {
					if s.consumed || s.lane != n.lane || s.width != n.width {
						continue
					}

					s.consumed = true
				}

				for _, a := range pack.Airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					a.consumed = true
				}

				if _, ok := slides[n.identifier]; ok {
					return nil, fmt.Errorf("Duplicated slide begin with same identifier: %s", string(n.identifier))
				}

				slides[n.identifier] = &SlideEvent{
					Seconds: secs,
					Track:   n.track(),
					Mark:    n.identifier,
					Width:   float64(n.width) / susLaneGaps,
				}
			case susSlideEnd:
				// + critical -> critical slide end (ignored)
				// + air -> slide with flick end
				for _, s := range pack.Shorts {
					if s.consumed || s.lane != n.lane || s.width != n.width {
						continue
					}

					s.consumed = true
				}

				flickEnd := false
				for _, a := range pack.Airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					flickEnd = true
					a.consumed = true
				}

				if _, ok := slides[n.identifier]; !ok {
					return nil, fmt.Errorf("Slide begin with identifier %s not found", string(n.identifier))
				}

				s := slides[n.identifier]
				s.Trace = append(s.Trace, &TraceItem{
					Seconds: secs,
					Track:   n.track(),
					Width:   float64(n.width) / susLaneGaps,
				})
				s.FlickEnd = flickEnd
				finalEvents = append(finalEvents, s)
				delete(slides, n.identifier)
			case susSlideStepInvisible:
				// + critical -> critical slide (ignored)
				// + tap + air -> slide with ease (ignored)
				for _, s := range pack.Shorts {
					if s.consumed || s.lane != n.lane || s.width != n.width {
						continue
					}

					s.consumed = true
				}

				for _, a := range pack.Airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					a.consumed = true
				}

				if _, ok := slides[n.identifier]; !ok {
					return nil, fmt.Errorf("Slide begin with identifier %s not found", string(n.identifier))
				}

				s := slides[n.identifier]
				s.Trace = append(s.Trace, &TraceItem{
					Seconds: secs,
					Track:   n.track(),
					Width:   float64(n.width) / susLaneGaps,
				})
			case susSlideStepVisible:
				// + flick -> any position mid
				// + critical -> critical slide (ignored)
				// + tap + air -> slide with ease (ignored)
				ignorePosition := false
				for _, s := range pack.Shorts {
					if s.consumed || s.lane != n.lane || s.width != n.width {
						continue
					}

					s.consumed = true
					if s.kind == susFlick {
						ignorePosition = true
					}
				}

				for _, a := range pack.Airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					a.consumed = true
				}

				if _, ok := slides[n.identifier]; !ok {
					return nil, fmt.Errorf("Slide begin with identifier %s not found", string(n.identifier))
				}

				if !ignorePosition {
					s := slides[n.identifier]
					s.Trace = append(s.Trace, &TraceItem{
						Seconds: secs,
						Track:   n.track(),
						Width:   float64(n.width) / susLaneGaps,
					})
				}
			}
		}

		for _, n := range pack.Shorts {
			if n.consumed {
				continue
			}

			switch n.kind {
			case susTap, susCritical:
				// + air -> flick
				flick := false
				var flickType uint8
				for _, a := range pack.Airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					a.consumed = true
					flick = true
					flickType = a.kind
				}

				if flick {
					finalEvents = append(finalEvents, &FlickEvent{
						Seconds: secs,
						Track:   n.track(),
						Offset:  susOffsetOf(flickType),
						Width:   float64(n.width) / susLaneGaps,
					})
				} else {
					finalEvents = append(finalEvents, &TapEvent{
						Seconds: secs,
						Track:   n.track(),
						Width:   float64(n.width) / susLaneGaps,
					})
				}
			case susTrend, susCriticalTrend:
				// + air -> throw
				flick := false
				var flickType uint8
				for _, a := range pack.Airs {
					if a.consumed || a.lane != n.lane || a.width != n.width {
						continue
					}

					a.consumed = true
					flick = true
					flickType = a.kind
				}

				if flick {
					finalEvents = append(finalEvents, &FlickEvent{
						Seconds: secs,
						Track:   n.track(),
						Offset:  susOffsetOf(flickType),
						Width:   float64(n.width) / susLaneGaps,
					})
				} else {
					finalEvents = append(finalEvents, &TapEvent{
						Seconds: secs,
						Track:   n.track(),
						Width:   float64(n.width) / susLaneGaps,
					})
				}
			case susCancel, susCriticalCancel:
				// [TODO] WTF is this?
			}
		}
	}

	return finalEvents, nil
}
