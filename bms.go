package main

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Channel byte

const (
	ChannelBackgroundMusic Channel = iota + 1
	ChannelTimeSignature
	ChannelBPMChange
	CHannelBackgroundAnimation
	ChannelPoorBitmapChange Channel = iota + 2
	ChannelLayer
	ChannelExtendedBPM
	ChannelStop

	ChannelNoteTrack1 Channel = 16
	ChannelNoteTrack2 Channel = 11
	ChannelNoteTrack3 Channel = 12
	ChannelNoteTrack4 Channel = 13
	ChannelNoteTrack5 Channel = 14
	ChannelNoteTrack6 Channel = 15
	ChannelNoteTrack7 Channel = 18

	ChannelSpecialTrack1 Channel = 36
	ChannelSpecialTrack2 Channel = 31
	ChannelSpecialTrack3 Channel = 32
	ChannelSpecialTrack4 Channel = 33
	ChannelSpecialTrack5 Channel = 34
	ChannelSpecialTrack6 Channel = 35
	ChannelSpecialTrack7 Channel = 38

	ChannelHoldTrack1 Channel = 56
	ChannelHoldTrack2 Channel = 51
	ChannelHoldTrack3 Channel = 52
	ChannelHoldTrack4 Channel = 53
	ChannelHoldTrack5 Channel = 54
	ChannelHoldTrack6 Channel = 55
	ChannelHoldTrack7 Channel = 58
)

var TRACKS_MAP map[Channel]int = map[Channel]int{
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

type BasicNoteType byte

const (
	NoteTypeNote BasicNoteType = iota
	NoteTypeFlick
	NoteTypeSlideA
	NoteTypeSlideEndA
	NoteTypeSlideEndFlickA
	NoteTypeSlideB
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
	"skill_slide_a.wav":          NoteTypeSlideA,
	"slide_end_a.wav":            NoteTypeSlideEndA,
	"slide_end_flick_a.wav":      NoteTypeSlideEndFlickA,
	"slide_b.wav":                NoteTypeSlideB,
	"slide_b_skill.wav":          NoteTypeSlideB,
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
		fmt.Println(err)
		panic("parse failed")
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
	if n.mark == "a" {
		return NoteTypeSlideA
	} else if n.mark == "b" {
		return NoteTypeSlideB
	} else {
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

type BPMEvent struct {
	BPM int
}

type TapEvent struct {
	Seconds float64
	TrackID int
}

type FlickEvent struct {
	Seconds float64
	TrackID int
	Offset  [2]float64
}

type HoldEvent struct {
	Seconds    float64
	EndSeconds float64
	TrackID    int
	FlickEnd   bool
}

type SlideEvent struct {
	Seconds  float64
	Track    float64
	Trace    [][2]float64
	FlickEnd bool
	Mark     string
}

type Event interface {
	_event()
}

func (BGMEvent) _event()   {}
func (BPMEvent) _event()   {}
func (TapEvent) _event()   {}
func (FlickEvent) _event() {}
func (HoldEvent) _event()  {}
func (SlideEvent) _event() {}

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

func (BGMEvent) _controlEvent() {}
func (BPMEvent) _controlEvent() {}

type ChartHeader struct {
	Player    int
	Genre     string
	Title     string
	Artist    string
	BPM       float64
	PlayLevel int
	Rank      int
	StageFile string
	Wavs      map[string]string
}

type Chart struct {
	Header ChartHeader
	Events []NoteEvent
}

type RawEvent struct {
	Channel  Channel
	NoteType NoteType
}

type RawEvents interface {
	_rawEvent()
}

func (BPMEvent) _rawEvent() {}
func (BGMEvent) _rawEvent() {}
func (RawEvent) _rawEvent() {}

func splitString2(s string) []string {
	var result []string
	for i := 0; i < len(s); i += 2 {
		result = append(result, s[i:i+2])
	}
	return result
}

func getKeys[K comparable, V any](m map[K]V) []K {
	var result []K
	for k := range m {
		result = append(result, k)
	}
	return result
}

func Parse(chartText string) Chart {
	const HEADER_BEGIN = "*---------------------- HEADER FIELD"
	const MAIN_DATA_BEGIN = "*---------------------- MAIN DATA FIELD"
	headerTag := regexp.MustCompile(`^#([0-9A-Z]+) (.*)$`)
	gameData := regexp.MustCompile(`#([0-9][0-9][0-9])([0-9][0-9]):([0-9A-Z]+)`)
	newline := regexp.MustCompile(`\r?\n`)

	wavs := map[string]string{}

	lines := newline.Split(chartText, -1)

	for !strings.Contains(lines[0], HEADER_BEGIN) {
		lines = lines[1:]
	}

	lines = lines[1:]

	header := ChartHeader{
		Player:    1,
		Genre:     "",
		Title:     "",
		Artist:    "",
		BPM:       130,
		PlayLevel: 0,
		StageFile: "",
		Rank:      3,
		Wavs:      wavs,
	}

	for ; !strings.Contains(lines[0], MAIN_DATA_BEGIN); lines = lines[1:] {
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
				log.Fatalf("failed to parse value of #PLAYER %q, err: %s", value, err.Error())
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
				panic("failed to parse value of #PLAYLEVEL")
			}
			header.PlayLevel = int(playlevel)
		case "STAGEFILE":
			header.StageFile = value
		case "RANK":
			rank, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				panic("failed to parse value of #RANk")
			}
			header.Rank = int(rank)
		case "BPM":
			bpm, err := strconv.ParseFloat(value, 64)
			if err != nil {
				panic("failed to parse value of #BPM")
			}
			header.BPM = bpm
		default:
			if strings.HasPrefix(key, "WAV") {
				point := key[3:]
				wavs[point] = value
			}
		}
	}

	lines = lines[1:]

	finalEvents := []NoteEvent{}
	rawEvents := map[float64][]RawEvents{}

	for len(lines) != 0 {
		line := lines[0]
		lines = lines[1:]

		m := gameData.FindStringSubmatch(line)
		if len(m) == 0 {
			continue
		}

		message := m[3]

		measure, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			panic("failed to parse measure value")
		}

		channelValue, err := strconv.ParseInt(m[2], 10, 64)
		if err != nil {
			panic("failed to parse channel value")
		}
		channel := Channel(channelValue)

		denominator := float64(len(message) / 2)
		for index, data := range splitString2(message) {
			tick := float64(measure) + float64(index)/denominator
			if data == "00" {
				continue
			}

			switch channel {
			case ChannelBackgroundMusic:
				_, ok := rawEvents[tick]
				if !ok {
					rawEvents[tick] = []RawEvents{}
				}
				rawEvents[tick] = append(rawEvents[tick], BGMEvent{
					BGM: header.Wavs[data],
				})
			case ChannelBPMChange:
				_, ok := rawEvents[tick]
				if !ok {
					rawEvents[tick] = []RawEvents{}
				}
				value, err := strconv.ParseInt(data, 16, 64)
				if err != nil {
					panic("failed to parse bpm value")
				}
				rawEvents[tick] = append(rawEvents[tick], BPMEvent{
					BPM: int(value),
				})
			default:
				_, ok := rawEvents[tick]
				if !ok {
					rawEvents[tick] = []RawEvents{}
				}

				wav, ok := header.Wavs[data]
				if !ok {
					fmt.Printf("unknown data %s, treated as tap note\n", data)
					rawEvents[tick] = append(rawEvents[tick], RawEvent{
						Channel:  channel,
						NoteType: NoteTypeNote,
					})
				}

				noteType, err := NoteTypeOf(wav)
				if err == nil {
					rawEvents[tick] = append(rawEvents[tick], RawEvent{
						Channel:  channel,
						NoteType: noteType,
					})
				}
			}
		}
	}

	ticks := getKeys(rawEvents)
	sort.Float64s(ticks)
	holdTracks := [7]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}
	slideA := [][2]float64{}
	slideB := [][2]float64{}
	secStart := 0.0
	tickStart := 0.0
	bpm := header.BPM

	for _, tick := range ticks {
		events := rawEvents[tick]
		for _, event := range events {
			switch ev := event.(type) {
			case BGMEvent:
				continue
			case BPMEvent:
				secStart += 240.0 / bpm * (tick - tickStart)
				tickStart = tick
				bpm = float64(ev.BPM)
				continue
			case RawEvent:
				switch ev.Channel {
				case ChannelNoteTrack1, ChannelNoteTrack2, ChannelNoteTrack3, ChannelNoteTrack4, ChannelNoteTrack5, ChannelNoteTrack6, ChannelNoteTrack7:
					trackID := TRACKS_MAP[ev.Channel]
					sec := secStart + 240.0/bpm*(tick-tickStart)
					switch ev.NoteType {
					case NoteTypeNote:
						finalEvents = append(finalEvents, TapEvent{
							Seconds: sec,
							TrackID: trackID,
						})

					case NoteTypeFlick:
						finalEvents = append(finalEvents, FlickEvent{
							Seconds: sec,
							TrackID: trackID,
							Offset:  [2]float64{0, 1},
						})
					case NoteTypeFlickLeft:
						finalEvents = append(finalEvents, FlickEvent{
							Seconds: sec,
							TrackID: trackID,
							Offset:  [2]float64{-1.0 / 6.0, 0},
						})
					case NoteTypeFlickRight:
						finalEvents = append(finalEvents, FlickEvent{
							Seconds: sec,
							TrackID: trackID,
							Offset:  [2]float64{1.0 / 6.0, 0},
						})

					case NoteTypeSlideA:
						slideA = append(slideA, [2]float64{sec, float64(trackID)})
					case NoteTypeSlideEndA:
						slideA = append(slideA, [2]float64{sec, float64(trackID)})
						start := slideA[0]
						slideA = slideA[1:]
						finalEvents = append(finalEvents, SlideEvent{
							Seconds:  start[0],
							Track:    start[1],
							Trace:    slideA,
							FlickEnd: false,
							Mark:     "a",
						})
						slideA = [][2]float64{}
					case NoteTypeSlideEndFlickA:
						slideA = append(slideA, [2]float64{sec, float64(trackID)})
						start := slideA[0]
						slideA = slideA[1:]
						finalEvents = append(finalEvents, SlideEvent{
							Seconds:  start[0],
							Track:    start[1],
							Trace:    slideA,
							FlickEnd: true,
							Mark:     "a",
						})
						slideA = [][2]float64{}

					case NoteTypeSlideB:
						slideB = append(slideB, [2]float64{sec, float64(trackID)})
					case NoteTypeSlideEndB:
						slideB = append(slideB, [2]float64{sec, float64(trackID)})
						start := slideB[0]
						slideB = slideB[1:]
						finalEvents = append(finalEvents, SlideEvent{
							Seconds:  start[0],
							Track:    start[1],
							Trace:    slideB,
							FlickEnd: false,
							Mark:     "b",
						})
						slideB = [][2]float64{}
					case NoteTypeSlideEndFlickB:
						slideB = append(slideB, [2]float64{sec, float64(trackID)})
						start := slideB[0]
						slideB = slideB[1:]
						finalEvents = append(finalEvents, SlideEvent{
							Seconds:  start[0],
							Track:    start[1],
							Trace:    slideB,
							FlickEnd: true,
							Mark:     "b",
						})
						slideB = [][2]float64{}
					default:
						fmt.Printf("unknown note type %s on note track %d\n", ev.NoteType, trackID)
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
						fmt.Printf("unknown note type %s on note track %d\n", ev.NoteType, trackID)
					}
				case ChannelSpecialTrack1, ChannelSpecialTrack2, ChannelSpecialTrack3, ChannelSpecialTrack4, ChannelSpecialTrack5, ChannelSpecialTrack6, ChannelSpecialTrack7:
					trackID := TRACKS_MAP[ev.Channel]
					sec := secStart + 240.0/bpm*(tick-tickStart)
					switch nt := ev.NoteType.(type) {
					case SpecialSlideNoteType:
						if nt.mark == "a" {
							slideA = append(slideA, [2]float64{sec, float64(trackID) + nt.offset})
						} else if nt.mark == "b" {
							slideB = append(slideB, [2]float64{sec, float64(trackID) + nt.offset})
						} else {
							fmt.Printf("unknown mark %s\n", nt.mark)
						}
					case BasicNoteType:
						switch nt {
						case NoteTypeSlideA:
							slideA = append(slideA, [2]float64{sec, float64(trackID)})
						case NoteTypeSlideB:
							slideB = append(slideB, [2]float64{sec, float64(trackID)})
						default:
							fmt.Printf("%s should not appear on channel %d (tick = %f)", ev.NoteType, ev.Channel, tick)
						}
					default:
						fmt.Printf("%s should not appear on channel %d (tick = %f)", ev.NoteType, ev.Channel, tick)
					}
				}
			}
		}
	}

	return Chart{
		Header: header,
		Events: finalEvents,
	}
}
