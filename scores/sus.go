// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"fmt"
	"strconv"
	"strings"
)

type HighSpeedChangeEvent struct {
	BarIndex   int
	TickOffset int
	SpeedRatio float64
}

type SUSHeader struct {
	Title                 string
	Artist                string
	Designer              string
	Difficulty            string
	PlayLevel             string
	SongID                string
	Wave                  string
	WaveOffset            string
	Jacket                string
	Volume                string
	TicksPerBeat          int
	TimeSignatures        map[int]float64
	BPMChangeEvents       []*BPMEvent
	HighSpeedChangeEvents []*HighSpeedChangeEvent
}

const (
	SUSTap    = "1"
	SUSExTap  = "2"
	SUSFlick  = "3"
	SUSDamage = "4"

	SUSAirUp         = "1"
	SUSAirDown       = "2"
	SUSAirUpperLeft  = "3"
	SUSAirUpperRight = "4"
	SUSAirLowerLeft  = "5"
	SUSAirLowerRight = "6"

	SUSHoldBegin = "1"
	SUSHoldEnd   = "2"

	SUSSlideBegin         = "1"
	SUSSlideEnd           = "2"
	SUSSlideStepVisible   = "3"
	SUSSlideStepInvisible = "4"

	SUSAirActionBegin  = "1"
	SUSAirActionEnd    = "2"
	SUSAirActionAction = "3"
)

type SUSRawShortNoteEvent struct {
	Type  string
	Width int
}

type SUS struct {
	Header *SUSHeader
}

func ParseSUS(chartText string) (*SUS, error) {
	header := &SUSHeader{}
	timeSignatures := map[int]float64{}
	bpms := map[string]float64{}
	bpmEvents := []*BPMEvent{}
	var highSpeedEvents []*HighSpeedChangeEvent
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
			header.Title = value
		case "ARTIST":
			header.Artist = value
		case "DESIGNER":
			header.Designer = value
		case "DIFFICULTY":
			header.Difficulty = value
		case "PLAYLEVEL":
			header.PlayLevel = value
		case "SONGID":
			header.SongID = value
		case "WAVE":
			header.Wave = value
		case "WAVEOFFSET":
			header.WaveOffset = value
		case "JACKET":
			header.Jacket = value
		case "REQUEST":
			if strings.HasPrefix(value, "ticks_per_beat") {
				value = value[15:] // skip "ticks_per_beat "
				value = strings.TrimSpace(value)
				i, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse ticks_per_beat: %s", err)
				}
				header.TicksPerBeat = int(i)
			}
		case "TIL00":
			highSpeedEvents = nil
			items := strings.Split(value, ", ")
			for _, item := range items {
				s1 := strings.Split(item, "'")
				if len(s1) != 2 {
					return nil, fmt.Errorf("Failed to parse high speed change event list item `%s`", item)
				}

				barIndex, err := strconv.ParseInt(s1[0], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse high speed change event list item `%s`: %s", item, err)
				}

				s2 := strings.Split(s1[1], ":")
				if len(s2) != 2 {
					return nil, fmt.Errorf("Failed to parse high speed change event list item `%s`: %s", item, err)
				}

				tickOffset, err := strconv.ParseInt(s2[0], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse high speed change event list item `%s`: %s", item, err)
				}

				speedRatio, err := strconv.ParseFloat(s2[1], 64)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse high speed change event list item `%s`: %s", item, err)
				}

				highSpeedEvents = append(highSpeedEvents, &HighSpeedChangeEvent{
					BarIndex:   int(barIndex),
					TickOffset: int(tickOffset),
					SpeedRatio: speedRatio,
				})
			}
		case "VOLUME":
			header.Volume = value
		case "HISPEED":
			// do nothing
		case "MEASUREHS":
			// do nothing
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

				timeSignatures[int(index)] = sig
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
						bpmEvents = append(bpmEvents, &BPMEvent{
							Tick: ev.Tick(),
							BPM:  bpm,
						})
					}
				}
				continue
			}

			switch common.Channel[0] {
			case '1': // short notes
			case '5': // air
			case '2': // holds
			case '3': // slides
			}
		}
	}

	return &SUS{
		Header: header,
	}, nil
}
