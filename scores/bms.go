// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kvarenzn/ssm/common"
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

type BGMEvent struct {
	BGM string
}

type BPMRawEvent struct {
	BPM float64
}

type TapEvent struct {
	Seconds float64
	TrackID int
}

type FlickEvent struct {
	Seconds float64
	TrackID int
	Offset  common.Point2D
}

type HoldEvent struct {
	Seconds    float64
	EndSeconds float64
	TrackID    int
	FlickEnd   bool
}

type TraceItem struct {
	Tick  float64
	Track float64
}

type SlideEvent struct {
	Seconds  float64
	Track    float64
	Trace    []TraceItem
	FlickEnd bool
	Mark     string
}

type Event interface {
	_event()
}

func (BGMEvent) _event()    {}
func (BPMRawEvent) _event() {}
func (TapEvent) _event()    {}
func (FlickEvent) _event()  {}
func (HoldEvent) _event()   {}
func (SlideEvent) _event()  {}

type NoteEvent interface {
	Start() float64
}

func (e TapEvent) Start() float64 {
	return e.Seconds
}

func (e FlickEvent) Start() float64 {
	return e.Seconds
}

func (e HoldEvent) Start() float64 {
	return e.Seconds
}

func (e SlideEvent) Start() float64 {
	return e.Seconds
}

type ControlEvent interface {
	_controlEvent()
}

func (BGMEvent) _controlEvent()    {}
func (BPMRawEvent) _controlEvent() {}

type BMSHeader struct {
	Player       int
	Genre        string
	Title        string
	Artist       string
	BPM          float64
	PlayLevel    int
	Rank         int
	LongNoteType int
	StageFile    string
	Wavs         map[string]string
	ExtendedBPM  map[string]float64
	BGM          string
}

type BMS struct {
	Header     *BMSHeader
	BPMEvents  []*BPMEvent
	NoteEvents []NoteEvent
}

type RawEvent struct {
	Channel  string
	NoteType NoteType
	Extra    int
}

type EventsPack struct {
	BGMEvents []BGMEvent
	BPMEvents []BPMRawEvent
	RawEvents []RawEvent
}

func ParseBMS(chartText string) *BMS {
	const FIELD_BEGIN = "*----------------------"
	const HEADER_BEGIN = "*---------------------- HEADER FIELD"
	const EXPANSION_BEGIN = "*---------------------- EXPANSION FIELD"
	const MAIN_DATA_BEGIN = "*---------------------- MAIN DATA FIELD"
	headerTag := regexp.MustCompile(`^#([0-9A-Z]+) (.*)$`)
	extendedHeaderTag := regexp.MustCompile(`^#([0-9A-Z]+) (.*)$`)
	newline := regexp.MustCompile(`\r?\n`)

	wavs := map[string]string{}
	extendedBPM := map[string]float64{}

	lines := newline.Split(chartText, -1)

	// drop anything before header
	for !strings.Contains(lines[0], HEADER_BEGIN) {
		lines = lines[1:]
	}

	lines = lines[1:]

	header := &BMSHeader{
		Player:       1,
		Genre:        "",
		Title:        "",
		Artist:       "",
		BPM:          130,
		PlayLevel:    0,
		StageFile:    "",
		Rank:         3,
		LongNoteType: 1,
		Wavs:         wavs,
		ExtendedBPM:  extendedBPM,
		BGM:          "",
	}

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
			player, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				log.Fatalf("failed to parse value of #PLAYER %q, err: %+v", value, err.Error())
			}
			header.Player = int(player)
		case "GENRE":
			header.Genre = value
		case "TITLE":
			header.Title = value
		case "ARTIST":
			header.Artist = value
		case "PLAYLEVEL":
			playlevel, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				log.Fatalf("failed to parse value of #PLAYLEVEL(%s), err: %+v", value, err)
			}
			header.PlayLevel = int(playlevel)
		case "STAGEFILE":
			header.StageFile = value
		case "RANK":
			rank, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				log.Fatalf("failed to parse value of #RANK(%s), err: %+v", value, err)
			}
			header.Rank = int(rank)
		case "LNTYPE":
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				log.Fatalf("failed to parse value of #LNTYPE(%s), err: %+v", value, err)
			}
			header.LongNoteType = int(t)
		case "BPM":
			bpm, err := strconv.ParseFloat(value, 64)
			if err != nil {
				log.Fatalf("failed to parse value of #BPM(%s), err: %+v", value, err)
			}
			header.BPM = bpm
		case "BGM":
			header.BGM = value
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
				header.BGM = value
			default:
				log.Warnf("unknown command in EXPANSION FIELD: %s: %s", key, value)
			}
		}
	}

	// MAIN DATA FILED
	lines = lines[1:]

	finalEvents := []NoteEvent{}
	rawEvents := map[float64]*EventsPack{}

	directionalFlickTicks := map[float64][7]byte{}
	for len(lines) != 0 {
		line := lines[0]
		lines = lines[1:]

		events, err := parseDataLine(line)
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
				if _, ok := rawEvents[tick]; !ok {
					rawEvents[tick] = &EventsPack{}
				}

				rawEvents[tick].BGMEvents = append(rawEvents[tick].BGMEvents, BGMEvent{
					BGM: header.Wavs[ev.Type],
				})
			case ChannelBPMChange:
				if _, ok := rawEvents[tick]; !ok {
					rawEvents[tick] = &EventsPack{}
				}

				value, err := strconv.ParseInt(ev.Type, 16, 64)
				if err != nil {
					log.Fatalf("failed to parse value of bpm(%s), err: %+v", ev.Type, err)
				}

				rawEvents[tick].BPMEvents = append(rawEvents[tick].BPMEvents, BPMRawEvent{
					BPM: float64(value),
				})
			case ChannelExtendedBPM:
				if _, ok := rawEvents[tick]; !ok {
					rawEvents[tick] = &EventsPack{}
				}

				rawEvents[tick].BPMEvents = append(rawEvents[tick].BPMEvents, BPMRawEvent{
					BPM: extendedBPM[ev.Type],
				})
			default:
				if _, ok := rawEvents[tick]; !ok {
					rawEvents[tick] = &EventsPack{}
				}

				wav, ok := header.Wavs[ev.Type]
				if !ok {
					rawEvents[tick].RawEvents = append(rawEvents[tick].RawEvents, RawEvent{
						Channel:  channel,
						NoteType: NoteTypeNote,
					})
					continue
				}

				noteType, err := NoteTypeOf(wav)
				if err == nil {
					rawEvents[tick].RawEvents = append(rawEvents[tick].RawEvents, RawEvent{
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
					log.Warnf("failed to get note type: %+v, skipped", err)
				}
			}
		}
	}

	for tick, v := range directionalFlickTicks {
		start := -1
		length := 0
		newEvents := []RawEvent{}
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
					newEvents = append(newEvents, RawEvent{
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
					newEvents = append(newEvents, RawEvent{
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

	// for _, tick := range ticks {
	// 	fmt.Printf("%f: %+v\n", tick, rawEvents[tick])
	// }

	holdTracks := [7]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}
	slideA := []TraceItem{}
	slideB := []TraceItem{}
	secStart := 0.0
	tickStart := 0.0
	bpm := header.BPM

	bpmEvents := []*BPMEvent{}

	for _, tick := range ticks {
		pack := rawEvents[tick]
		for _, bpmEvent := range pack.BPMEvents {
			bpmEvents = append(bpmEvents, &BPMEvent{
				Tick: tick,
				BPM:  bpmEvent.BPM,
			})
			secStart += 240.0 / bpm * (tick - tickStart)
			tickStart = tick
			bpm = float64(bpmEvent.BPM)
		}

		rawNoteEvents := pack.RawEvents
		sort.Slice(rawNoteEvents, func(i, j int) bool {
			return rawNoteEvents[i].NoteType.NoteType() >= rawNoteEvents[j].NoteType.NoteType()
		})

		for _, ev := range rawNoteEvents {
			switch ev.Channel {
			case ChannelNoteTrack1, ChannelNoteTrack2, ChannelNoteTrack3, ChannelNoteTrack4, ChannelNoteTrack5, ChannelNoteTrack6, ChannelNoteTrack7:
				trackID := TRACKS_MAP[ev.Channel]
				sec := secStart + 240.0/bpm*(tick-tickStart)
				switch ev.NoteType {
				// normal note
				case NoteTypeNote:
					finalEvents = append(finalEvents, TapEvent{
						Seconds: sec,
						TrackID: trackID,
					})
				// flick note
				case NoteTypeFlick:
					finalEvents = append(finalEvents, FlickEvent{
						Seconds: sec,
						TrackID: trackID,
						Offset:  common.Point2D{X: 0, Y: 1},
					})
				case NoteTypeFlickLeft:
					finalEvents = append(finalEvents, FlickEvent{
						Seconds: sec,
						TrackID: trackID,
						Offset:  common.Point2D{X: -float64(ev.Extra) / 6.0, Y: 0},
					})
				case NoteTypeFlickRight:
					finalEvents = append(finalEvents, FlickEvent{
						Seconds: sec,
						TrackID: trackID,
						Offset:  common.Point2D{X: float64(ev.Extra) / 6.0, Y: 0},
					})
				// slide a
				case NoteTypeSlideA:
					slideA = append(slideA, TraceItem{sec, float64(trackID)})
				case NoteTypeSlideEndA:
					slideA = append(slideA, TraceItem{sec, float64(trackID)})
					start := slideA[0]
					slideA = slideA[1:]
					finalEvents = append(finalEvents, SlideEvent{
						Seconds:  start.Tick,
						Track:    start.Track,
						Trace:    slideA,
						FlickEnd: false,
						Mark:     "a",
					})
					slideA = []TraceItem{}
				case NoteTypeSlideEndFlickA:
					slideA = append(slideA, TraceItem{sec, float64(trackID)})
					start := slideA[0]
					slideA = slideA[1:]
					finalEvents = append(finalEvents, SlideEvent{
						Seconds:  start.Tick,
						Track:    start.Track,
						Trace:    slideA,
						FlickEnd: true,
						Mark:     "a",
					})
					slideA = []TraceItem{}
				// slide b
				case NoteTypeSlideB:
					slideB = append(slideB, TraceItem{sec, float64(trackID)})
				case NoteTypeSlideEndB:
					slideB = append(slideB, TraceItem{sec, float64(trackID)})
					start := slideB[0]
					slideB = slideB[1:]
					finalEvents = append(finalEvents, SlideEvent{
						Seconds:  start.Tick,
						Track:    start.Track,
						Trace:    slideB,
						FlickEnd: false,
						Mark:     "b",
					})
					slideB = []TraceItem{}
				case NoteTypeSlideEndFlickB:
					slideB = append(slideB, TraceItem{sec, float64(trackID)})
					start := slideB[0]
					slideB = slideB[1:]
					finalEvents = append(finalEvents, SlideEvent{
						Seconds:  start.Tick,
						Track:    start.Track,
						Trace:    slideB,
						FlickEnd: true,
						Mark:     "b",
					})
					slideB = []TraceItem{}
				// unknown
				default:
					log.Warnf("unknown note type %s on note track %d\n", ev.NoteType, trackID)
				}
			case ChannelHoldTrack1, ChannelHoldTrack2, ChannelHoldTrack3, ChannelHoldTrack4, ChannelHoldTrack5, ChannelHoldTrack6, ChannelHoldTrack7:
				trackID := TRACKS_MAP[ev.Channel]
				sec := secStart + 240.0/bpm*(tick-tickStart)
				switch ev.NoteType {
				case NoteTypeNote:
					startTick := holdTracks[trackID]
					if math.IsNaN(startTick) {
						holdTracks[trackID] = sec
					} else {
						finalEvents = append(finalEvents, HoldEvent{
							Seconds:    startTick,
							EndSeconds: sec,
							TrackID:    trackID,
							FlickEnd:   false,
						})
						holdTracks[trackID] = math.NaN()
					}
				case NoteTypeFlick:
					startTick := holdTracks[trackID]
					if math.IsNaN(startTick) {
						log.Fatalf("no hold start data on track %d", trackID)
					}
					finalEvents = append(finalEvents, HoldEvent{
						Seconds:    startTick,
						EndSeconds: sec,
						TrackID:    trackID,
						FlickEnd:   true,
					})
					holdTracks[trackID] = math.NaN()
				default:
					log.Warnf("unknown note type %s on note track %d\n", ev.NoteType, trackID)
				}
			case ChannelSpecialTrack1, ChannelSpecialTrack2, ChannelSpecialTrack3, ChannelSpecialTrack4, ChannelSpecialTrack5, ChannelSpecialTrack6, ChannelSpecialTrack7:
				trackID := TRACKS_MAP[ev.Channel]
				sec := secStart + 240.0/bpm*(tick-tickStart)
				switch nt := ev.NoteType.(type) {
				case SpecialSlideNoteType:
					switch nt.mark {
					case "a":
						slideA = append(slideA, TraceItem{sec, float64(trackID) + nt.offset})
					case "b":
						slideB = append(slideB, TraceItem{sec, float64(trackID) + nt.offset})
					default:
						log.Warnf("unknown mark %s\n", nt.mark)
					}
				case BasicNoteType:
					switch nt {
					case NoteTypeSlideA:
						slideA = append(slideA, TraceItem{sec, float64(trackID)})
					case NoteTypeSlideB:
						slideB = append(slideB, TraceItem{sec, float64(trackID)})
					default:
						log.Warnf("%s should not appear on channel %d (tick = %f)", ev.NoteType, ev.Channel, tick)
					}
				default:
					log.Warnf("%s should not appear on channel %d (tick = %f)", ev.NoteType, ev.Channel, tick)
				}
			}
		}
	}

	if len(slideA) > 0 {
		start := slideA[0]
		slideA = slideA[1:]
		finalEvents = append(finalEvents, SlideEvent{
			Seconds:  start.Tick,
			Track:    start.Track,
			Trace:    slideA,
			FlickEnd: false,
			Mark:     "a",
		})
		slideA = []TraceItem{}
	}

	if len(slideB) > 0 {
		start := slideB[0]
		slideB = slideB[1:]
		finalEvents = append(finalEvents, SlideEvent{
			Seconds:  start.Tick,
			Track:    start.Track,
			Trace:    slideB,
			FlickEnd: false,
			Mark:     "b",
		})
		slideB = []TraceItem{}
	}

	return &BMS{
		Header:     header,
		BPMEvents:  bpmEvents,
		NoteEvents: finalEvents,
	}
}
