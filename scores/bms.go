// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"cmp"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/kvarenzn/ssm/log"
	"github.com/kvarenzn/ssm/utils"
)

const (
	ChannelBackgroundMusic     = "01"
	ChannelTimeSignature       = "02"
	ChannelBPMChange           = "03"
	ChannelBackgroundAnimation = "04"
	ChannelPoorBitmapChange    = "06"
	ChannelLayer               = "07"
	ChannelExtendedBPM         = "08"
	ChannelStop                = "09"

	ChannelNoteTrack1 = "16"
	ChannelNoteTrack2 = "11"
	ChannelNoteTrack3 = "12"
	ChannelNoteTrack4 = "13"
	ChannelNoteTrack5 = "14"
	ChannelNoteTrack6 = "15"
	ChannelNoteTrack7 = "18"

	ChannelSpecialTrack1 = "36"
	ChannelSpecialTrack2 = "31"
	ChannelSpecialTrack3 = "32"
	ChannelSpecialTrack4 = "33"
	ChannelSpecialTrack5 = "34"
	ChannelSpecialTrack6 = "35"
	ChannelSpecialTrack7 = "38"

	ChannelHoldTrack1 = "56"
	ChannelHoldTrack2 = "51"
	ChannelHoldTrack3 = "52"
	ChannelHoldTrack4 = "53"
	ChannelHoldTrack5 = "54"
	ChannelHoldTrack6 = "55"
	ChannelHoldTrack7 = "58"
)

var TRACKS_MAP = map[string]int{
	ChannelNoteTrack1:    0,
	ChannelNoteTrack2:    1,
	ChannelNoteTrack3:    2,
	ChannelNoteTrack4:    3,
	ChannelNoteTrack5:    4,
	ChannelNoteTrack6:    5,
	ChannelNoteTrack7:    6,
	ChannelSpecialTrack1: 0,
	ChannelSpecialTrack2: 1,
	ChannelSpecialTrack3: 2,
	ChannelSpecialTrack4: 3,
	ChannelSpecialTrack5: 4,
	ChannelSpecialTrack6: 5,
	ChannelSpecialTrack7: 6,
	ChannelHoldTrack1:    0,
	ChannelHoldTrack2:    1,
	ChannelHoldTrack3:    2,
	ChannelHoldTrack4:    3,
	ChannelHoldTrack5:    4,
	ChannelHoldTrack6:    5,
	ChannelHoldTrack7:    6,
}

var simpleTracks = []string{
	ChannelNoteTrack1,
	ChannelNoteTrack2,
	ChannelNoteTrack3,
	ChannelNoteTrack4,
	ChannelNoteTrack5,
	ChannelNoteTrack6,
	ChannelNoteTrack7,
}

type BasicNoteType byte

const (
	NoteTypeNote BasicNoteType = iota
	NoteTypeFlick
	NoteTypeSlideA
	NoteTypeSlideB
	NoteTypeSlideEndA
	NoteTypeSlideEndFlickA
	NoteTypeSlideEndB
	NoteTypeSlideEndFlickB
	NoteTypeFlickLeft
	NoteTypeFlickRight
)

var wavNoteTypeMap map[string]BasicNoteType = map[string]BasicNoteType{
	"bd.wav":                     NoteTypeNote,
	"flick.wav":                  NoteTypeFlick,
	"無音_flick.wav":               NoteTypeFlick,
	"skill.wav":                  NoteTypeNote,
	"slide_a.wav":                NoteTypeSlideA,
	"slide_a_skill.wav":          NoteTypeSlideA,
	"slide_a_fever.wav":          NoteTypeSlideA,
	"skill_slide_a.wav":          NoteTypeSlideA,
	"slide_end_a.wav":            NoteTypeSlideEndA,
	"slide_end_flick_a.wav":      NoteTypeSlideEndFlickA,
	"slide_b.wav":                NoteTypeSlideB,
	"slide_b_skill.wav":          NoteTypeSlideB,
	"slide_b_fever.wav":          NoteTypeSlideB,
	"skill_slide_b.wav":          NoteTypeSlideB,
	"slide_end_b.wav":            NoteTypeSlideEndB,
	"slide_end_flick_b.wav":      NoteTypeSlideEndFlickB,
	"fever_note.wav":             NoteTypeNote,
	"fever_note_flick.wav":       NoteTypeFlick,
	"fever_note_slide_a.wav":     NoteTypeSlideA,
	"fever_note_slide_end_a.wav": NoteTypeSlideEndA,
	"fever_note_slide_b.wav":     NoteTypeSlideB,
	"fever_note_slide_end_b.wav": NoteTypeSlideEndB,
	"fever_slide_a.wav":          NoteTypeSlideA,
	"fever_slide_end_a.wav":      NoteTypeSlideEndA,
	"fever_slide_b.wav":          NoteTypeSlideB,
	"fever_slide_end_b.wav":      NoteTypeSlideEndB,
	"directional_fl_l.wav":       NoteTypeFlickLeft,
	"directional_fl_r.wav":       NoteTypeFlickRight,
}

type NoteType interface {
	String() string
	NoteType() BasicNoteType
	Mark() string
	Offset() float64
}

func (n BasicNoteType) String() string {
	switch n {
	case NoteTypeNote:
		return "Tap"
	case NoteTypeFlick:
		return "Flick"
	case NoteTypeSlideA:
		return "Slide A"
	case NoteTypeSlideEndA:
		return "Slide End A"
	case NoteTypeSlideEndFlickA:
		return "Slide End Flick A"
	case NoteTypeSlideB:
		return "Slide B"
	case NoteTypeSlideEndB:
		return "Slide End B"
	case NoteTypeSlideEndFlickB:
		return "Slide End Flick B"
	default:
		return "Unknown"
	}
}

func (n BasicNoteType) NoteType() BasicNoteType {
	return n
}

func (n BasicNoteType) Mark() string {
	switch n {
	case NoteTypeSlideA, NoteTypeSlideEndA, NoteTypeSlideEndFlickA:
		return "a"
	case NoteTypeSlideB, NoteTypeSlideEndB, NoteTypeSlideEndFlickB:
		return "b"
	default:
		return ""
	}
}

func (n BasicNoteType) Offset() float64 {
	return 0.0
}

type SpecialSlideNoteType struct {
	mark   string
	offset float64
}

func NewSpecialSlideNoteType(name string) (SpecialSlideNoteType, error) {
	re := regexp.MustCompile(`slide_(.)_(L|R)S(\d\d)\.wav`)
	subs := re.FindStringSubmatch(name)
	if len(subs) < 3 {
		return SpecialSlideNoteType{}, fmt.Errorf("not a special slide note type")
	}
	mark := subs[1]
	direction := subs[2]
	rawOffset := subs[3]

	offInt, err := strconv.ParseInt(rawOffset, 10, 64)
	if err != nil {
		log.Fatalf("parse rawOffset(%s) failed: %s", rawOffset, err)
	}
	offset := float64(offInt) / 100.0

	if direction == "L" {
		offset = -offset
	}

	return SpecialSlideNoteType{
		mark:   mark,
		offset: offset,
	}, nil
}

func (n SpecialSlideNoteType) String() string {
	return fmt.Sprintf("Slide Special %s", n.mark)
}

func (n SpecialSlideNoteType) NoteType() BasicNoteType {
	switch n.mark {
	case "a":
		return NoteTypeSlideA
	case "b":
		return NoteTypeSlideB
	default:
		return NoteTypeNote
	}
}

func (n SpecialSlideNoteType) Mark() string {
	return n.mark
}

func (n SpecialSlideNoteType) Offset() float64 {
	return n.offset
}

func NoteTypeOf(wav string) (NoteType, error) {
	basicType, ok := wavNoteTypeMap[wav]
	if ok {
		return basicType, nil
	}

	note, err := NewSpecialSlideNoteType(wav)
	if err == nil {
		return note, nil
	}

	return NoteTypeNote, fmt.Errorf("unknown wav: %s", wav)
}

type bmsBPMEvent struct {
	Tick float64
	BPM  float64
}

type bmsRawEvent struct {
	Channel  string
	NoteType NoteType
	Extra    int
}

type bmsEventsPack struct {
	bpmEvents []float64
	RawEvents []*bmsRawEvent
}

func ParseBMS(chartText string) Chart {
	const barLength = 4
	const FIELD_BEGIN = "*----------------------"
	const HEADER_BEGIN = "*---------------------- HEADER FIELD"
	const EXPANSION_BEGIN = "*---------------------- EXPANSION FIELD"
	const MAIN_DATA_BEGIN = "*---------------------- MAIN DATA FIELD"
	headerTag := regexp.MustCompile(`^#([0-9A-Z]+) (.*)$`)
	extendedHeaderTag := regexp.MustCompile(`^#([0-9A-Z]+) (.*)$`)
	newline := regexp.MustCompile(`\r?\n`)

	bpm := 130.0
	wavs := map[string]string{}
	extendedBPM := map[string]float64{}

	lines := newline.Split(chartText, -1)

	// drop anything before header
	for !strings.Contains(lines[0], HEADER_BEGIN) {
		lines = lines[1:]
	}

	lines = lines[1:]

	// HEADER FIELD
	for ; !strings.Contains(lines[0], FIELD_BEGIN); lines = lines[1:] {
		subs := headerTag.FindStringSubmatch(lines[0])
		if len(subs) == 0 {
			continue
		}

		key := subs[1]
		value := subs[2]

		switch key {
		case "PLAYER":
		case "GENRE":
		case "TITLE":
		case "ARTIST":
		case "PLAYLEVEL":
		case "STAGEFILE":
		case "RANK":
		case "LNTYPE":
		case "BPM":
			var err error
			bpm, err = strconv.ParseFloat(value, 64)
			if err != nil {
				log.Fatalf("failed to parse value of #BPM(%s), err: %+v", value, err)
			}
		case "BGM":
		default:
			if strings.HasPrefix(key, "WAV") {
				point := key[3:]
				wavs[point] = value
			} else if strings.HasPrefix(key, "BPM") {
				point := key[3:]
				bpm, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.Fatalf("failed to parse value of #BPM%s(%s), err: %+v", point, value, err)
				}
				extendedBPM[point] = bpm
			} else {
				log.Warnf("unknown command in HEADER FIELD: %s: %s", key, value)
			}
		}
	}

	// EXPANSION FIELD
	if strings.Contains(lines[0], EXPANSION_BEGIN) {
		for ; !strings.Contains(lines[0], FIELD_BEGIN); lines = lines[1:] {
			subs := extendedHeaderTag.FindStringSubmatch(lines[0])
			if len(subs) == 0 {
				continue
			}

			key := subs[1]
			value := subs[2]
			switch key {
			case "BGM":
			default:
				log.Warnf("unknown command in EXPANSION FIELD: %s: %s", key, value)
			}
		}
	}

	// MAIN DATA FILED
	lines = lines[1:]

	finalEvents := []*star{}
	rawEvents := map[float64]*bmsEventsPack{}

	directionalFlickTicks := map[float64][7]byte{}
	for len(lines) != 0 {
		line := lines[0]
		lines = lines[1:]

		events, _, err := parseDataLine(line)
		if err == errInvalidDataLineFormat {
			continue
		} else if err != nil {
			log.Fatalf("Failed to parse line %s: %s", line, err)
		}

		for _, ev := range events {
			tick := ev.Tick()
			channel := ev.Common.Channel

			switch channel {
			case ChannelBackgroundMusic:
				// do nothing
			case ChannelBPMChange:
				if _, ok := rawEvents[tick]; !ok {
					rawEvents[tick] = &bmsEventsPack{}
				}

				value, err := strconv.ParseInt(ev.Type, 16, 64)
				if err != nil {
					log.Fatalf("failed to parse value of bpm(%s), err: %+v", ev.Type, err)
				}

				rawEvents[tick].bpmEvents = append(rawEvents[tick].bpmEvents, float64(value))
			case ChannelExtendedBPM:
				if _, ok := rawEvents[tick]; !ok {
					rawEvents[tick] = &bmsEventsPack{}
				}

				rawEvents[tick].bpmEvents = append(rawEvents[tick].bpmEvents, extendedBPM[ev.Type])
			default:
				if _, ok := rawEvents[tick]; !ok {
					rawEvents[tick] = &bmsEventsPack{}
				}

				wav, ok := wavs[ev.Type]
				if !ok {
					rawEvents[tick].RawEvents = append(rawEvents[tick].RawEvents, &bmsRawEvent{
						Channel:  channel,
						NoteType: NoteTypeNote,
					})
					continue
				}

				noteType, err := NoteTypeOf(wav)
				if err == nil {
					rawEvents[tick].RawEvents = append(rawEvents[tick].RawEvents, &bmsRawEvent{
						Channel:  channel,
						NoteType: noteType,
					})

					// record directional flicks
					if noteType == NoteTypeFlickLeft || noteType == NoteTypeFlickRight {
						if _, ok := directionalFlickTicks[tick]; !ok {
							directionalFlickTicks[tick] = [7]byte{}
						}
						v := directionalFlickTicks[tick]
						if noteType == NoteTypeFlickLeft {
							v[TRACKS_MAP[channel]] = '<'
						} else {
							v[TRACKS_MAP[channel]] = '>'
						}
						directionalFlickTicks[tick] = v
					}
				} else {
					log.Warnf("failed to get note type: %+v, treated as normal tap", err)
					rawEvents[tick].RawEvents = append(rawEvents[tick].RawEvents, &bmsRawEvent{
						Channel:  channel,
						NoteType: NoteTypeNote,
					})
				}
			}
		}
	}

	for tick, v := range directionalFlickTicks {
		start := -1
		length := 0
		newEvents := []*bmsRawEvent{}
		for i, c := range append(v[:], 0) {
			if c == '>' {
				if start == -1 {
					start = i
					length = 1
				} else {
					length++
				}
			} else {
				if start != -1 {
					newEvents = append(newEvents, &bmsRawEvent{
						Channel:  simpleTracks[start],
						NoteType: NoteTypeFlickRight,
						Extra:    length,
					})
					start = -1
					length = 0
				}
			}
		}

		rev := append([]byte{0}, v[:]...)
		for i := 6; i >= -1; i-- {
			c := rev[i+1]
			if c == '<' {
				if start == -1 {
					start = i
					length = 1
				} else {
					length++
				}
			} else {
				if start != -1 {
					newEvents = append(newEvents, &bmsRawEvent{
						Channel:  simpleTracks[start],
						NoteType: NoteTypeFlickLeft,
						Extra:    length,
					})
					start = -1
					length = 0
				}
			}
		}

		for _, ev := range rawEvents[tick].RawEvents {
			if ev.NoteType != NoteTypeFlickLeft && ev.NoteType != NoteTypeFlickRight {
				newEvents = append(newEvents, ev)
			}
		}

		rawEvents[tick].RawEvents = newEvents
	}

	ticks := utils.SortedKeysOf(rawEvents)

	holdTracks := [7]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}
	var slideA, slideB *star
	secStart := 0.0
	tickStart := 0.0

	bpmEvents := []*bmsBPMEvent{}

	for _, tick := range ticks {
		pack := rawEvents[tick]
		for _, bpmValue := range pack.bpmEvents {
			bpmEvents = append(bpmEvents, &bmsBPMEvent{
				Tick: tick,
				BPM:  bpmValue,
			})
			secStart += barLength * 60 / bpm * (tick - tickStart)
			tickStart = tick
			bpm = bpmValue
		}

		slices.SortFunc(pack.RawEvents, func(a, b *bmsRawEvent) int {
			return -cmp.Compare(a.NoteType.NoteType(), b.NoteType.NoteType())
		})

		for _, ev := range pack.RawEvents {
			switch ev.Channel {
			case ChannelNoteTrack1, ChannelNoteTrack2, ChannelNoteTrack3, ChannelNoteTrack4, ChannelNoteTrack5, ChannelNoteTrack6, ChannelNoteTrack7:
				trackID := float64(TRACKS_MAP[ev.Channel]) / 6
				sec := secStart + barLength*60/bpm*(tick-tickStart)
				switch ev.NoteType {
				// normal note
				case NoteTypeNote:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							markAsTap())
				// flick note
				case NoteTypeFlick:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							markAsTap().
							flickToIfOk(true, 90))
				case NoteTypeFlickLeft:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							markAsTap().
							flickToIfOk(true, 180))
				case NoteTypeFlickRight:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							markAsTap().
							flickToIfOk(true, 0))
				// slide a
				case NoteTypeSlideA:
					if slideA == nil {
						slideA = newStar(sec, trackID, 1.0/6).
							markAsTap().
							markAsHead()
					} else {
						slideA = newStar(sec, trackID, 1.0/6).
							chainsAfter(slideA)
					}
				case NoteTypeSlideEndA:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideA).
							markAsEnd())
					slideA = nil
				case NoteTypeSlideEndFlickA:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideA).
							flickToIfOk(true, 90).
							markAsEnd())
					slideA = nil
				// slide b
				case NoteTypeSlideB:
					if slideB == nil {
						slideB = newStar(sec, trackID, 1.0/6).
							markAsTap().
							markAsHead()
					} else {
						slideB = newStar(sec, trackID, 1.0/6).
							chainsAfter(slideB)
					}
				case NoteTypeSlideEndB:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideB).
							markAsEnd())
					slideB = nil
				case NoteTypeSlideEndFlickB:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideB).
							flickToIfOk(true, 90).
							markAsEnd())
					slideB = nil
				// unknown
				default:
					log.Warnf("unknown note type %s on note track %d\n", ev.NoteType, trackID)
				}
			case ChannelHoldTrack1, ChannelHoldTrack2, ChannelHoldTrack3, ChannelHoldTrack4, ChannelHoldTrack5, ChannelHoldTrack6, ChannelHoldTrack7:
				trackID := TRACKS_MAP[ev.Channel]
				trackX := float64(trackID) / 6
				sec := secStart + 240.0/bpm*(tick-tickStart)
				switch ev.NoteType {
				case NoteTypeNote:
					startTick := holdTracks[trackID]
					if math.IsNaN(startTick) {
						holdTracks[trackID] = sec
					} else {
						finalEvents = append(
							finalEvents,
							newStar(sec, trackX, 1.0/6).
								chainsAfter(
									newStar(startTick, trackX, 1.0/6).
										markAsTap().
										markAsHead(),
								).
								markAsEnd())
						holdTracks[trackID] = math.NaN()
					}
				case NoteTypeFlick:
					startTick := holdTracks[trackID]
					if math.IsNaN(startTick) {
						log.Fatalf("no hold start data on track %d", trackID)
					}
					finalEvents = append(
						finalEvents,
						newStar(sec, trackX, 1.0/6).
							chainsAfter(
								newStar(startTick, trackX, 1.0/6).
									markAsTap().
									markAsHead(),
							).
							flickToIfOk(true, 90).
							markAsEnd())
					holdTracks[trackID] = math.NaN()
				default:
					log.Warnf("unknown note type %s on note track %d\n", ev.NoteType, trackID)
				}
			case ChannelSpecialTrack1, ChannelSpecialTrack2, ChannelSpecialTrack3, ChannelSpecialTrack4, ChannelSpecialTrack5, ChannelSpecialTrack6, ChannelSpecialTrack7:
				trackID := float64(TRACKS_MAP[ev.Channel]) / 6
				sec := secStart + 240.0/bpm*(tick-tickStart)
				switch nt := ev.NoteType.(type) {
				case SpecialSlideNoteType:
					switch nt.mark {
					case "a":
						if slideA == nil {
							slideA = newStar(sec, trackID+nt.offset/6, 1.0/6).
								markAsTap().
								markAsHead()
						} else {
							slideA = newStar(sec, trackID+nt.offset/6, 1.0/6).
								chainsAfter(slideA)
						}
					case "b":
						if slideB == nil {
							slideB = newStar(sec, trackID+nt.offset/6, 1.0/6).
								markAsTap().
								markAsHead()
						} else {
							slideB = newStar(sec, trackID+nt.offset/6, 1.0/6).
								chainsAfter(slideB)
						}
					default:
						log.Warnf("unknown mark %s\n", nt.mark)
					}
				case BasicNoteType:
					switch nt {
					case NoteTypeSlideA:
						if slideA == nil {
							slideA = newStar(sec, trackID, 1.0/6).
								markAsTap().
								markAsHead()
						} else {
							slideA = newStar(sec, trackID, 1.0/6).
								chainsAfter(slideA)
						}
					case NoteTypeSlideB:
						if slideB == nil {
							slideB = newStar(sec, trackID, 1.0/6).
								markAsTap().
								markAsHead()
						} else {
							slideB = newStar(sec, trackID, 1.0/6).
								chainsAfter(slideB)
						}
					default:
						log.Warnf("%s should not appear on channel %d (tick = %f)", ev.NoteType, ev.Channel, tick)
					}
				default:
					log.Warnf("%s should not appear on channel %d (tick = %f)", ev.NoteType, ev.Channel, tick)
				}
			}
		}
	}

	if slideA != nil {
		finalEvents = append(finalEvents, slideA.markAsEnd())
		slideA = nil
	}

	if slideB != nil {
		finalEvents = append(finalEvents, slideB.markAsEnd())
		slideB = nil
	}

	return finalEvents
}
