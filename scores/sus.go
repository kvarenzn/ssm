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
	HighSpeedChangeEvents map[int]*HighSpeedChangeEvent
}

type SUS struct {
	Header *SUSHeader
}

func ParseSUS(chartText string) (*SUS, error) {
	header := &SUSHeader{}
	timeSignatures := map[int]float64{}
	bpms := map[int]float64{}
	for line := range strings.Lines(chartText) {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}

		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			spaceIndex := strings.Index(line, " ")
			key := ""
			value := ""
			if spaceIndex != -1 {
				key = line[1:spaceIndex] // skip '#'
				value = line[spaceIndex+1:]
			} else {
				key = line[1:]
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
			case "VOLUME":
				header.Volume = value
			case "HISPEED":
				// do nothing
			case "MEASUREHS":
				// do nothing
			default:
				if strings.HasPrefix(key, "BPM") {
					index, err := strconv.ParseInt(key[3:], 36, 64)
					if err != nil {
						return nil, fmt.Errorf("Failed to parse BPM list item key `%s`: %s", key, err)
					}

					bpm, err := strconv.ParseFloat(value, 64)
				}
			}
		}
	}
}
