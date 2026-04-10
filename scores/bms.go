// Copyright (C) 2024, 2025, 2026 kvarenzn
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

type bmsChannelKind byte

const (
	bmsCKConfig  bmsChannelKind = '0'
	bmsCKNote    bmsChannelKind = '1'
	bmsCKSpecial bmsChannelKind = '3'
	bmsCKHold    bmsChannelKind = '5'
)

const (
	bmsCTBPMChange   = '3'
	bmsCTExtendedBPM = '8'
)

type bmsChannel struct {
	kind bmsChannelKind
	typ  byte
}

func (c *bmsChannel) String() string {
	return fmt.Sprintf("%c%c", c.kind, c.typ)
}

func (c *bmsChannel) trackID() int {
	if c.kind == bmsCKConfig {
		log.Fatalf("Config channel has no track id")
	}

	switch c.typ {
	case '6':
		return 0
	case '1':
		return 1
	case '2':
		return 2
	case '3':
		return 3
	case '4':
		return 4
	case '5':
		return 5
	case '8':
		return 6
	}

	log.Fatalf("Unknown channel data %c", c.typ)
	return 0
}

type bmsBasicNoteType byte

const (
	bmsNTTap bmsBasicNoteType = iota
	bmsNTFlick
	bmsNTSlideA
	bmsNTSlideB
	bmsNTSlideEndA
	bmsNTSlideEndFlickA
	NoteTypeSlideEndB
	NoteTypeSlideEndFlickB
	NoteTypeFlickLeft
	NoteTypeFlickRight
	NoteTypeAddLongDirFlick
	NoteTypeAddSlideDirFlick
	NoteTypeContBezierFrontA
	NoteTypeContBezierFrontB
	NoteTypeContBezierBackA
	NoteTypeContBezierBackB
	NoteTypeLongEndDirFlickLeft
	NoteTypeLongEndDirFlickRight
	NoteTypeSlideEndDirFlickLeftA
	NoteTypeSlideEndDirFlickLeftB
	NoteTypeSlideEndDirFlickRightA
	NoteTypeSlideEndDirFlickRightB
	NoteTypeLaneChange
)

var wavNoteTypeMap map[string]bmsBasicNoteType = map[string]bmsBasicNoteType{
	"":                            bmsNTTap,
	"bd.wav":                      bmsNTTap,
	"flick.wav":                   bmsNTFlick,
	"無音_flick.wav":                bmsNTFlick,
	"skill.wav":                   bmsNTTap,
	"slide_a.wav":                 bmsNTSlideA,
	"slide_a_skill.wav":           bmsNTSlideA,
	"slide_a_fever.wav":           bmsNTSlideA,
	"skill_slide_a.wav":           bmsNTSlideA,
	"slide_end_a.wav":             bmsNTSlideEndA,
	"slide_end_flick_a.wav":       bmsNTSlideEndFlickA,
	"slide_end_dir_flick_l_a.wav": NoteTypeSlideEndDirFlickLeftA,
	"slide_end_dir_flick_r_a.wav": NoteTypeSlideEndDirFlickRightA,
	"slide_b.wav":                 bmsNTSlideB,
	"slide_b_skill.wav":           bmsNTSlideB,
	"slide_b_fever.wav":           bmsNTSlideB,
	"skill_slide_b.wav":           bmsNTSlideB,
	"slide_end_b.wav":             NoteTypeSlideEndB,
	"slide_end_flick_b.wav":       NoteTypeSlideEndFlickB,
	"slide_end_dir_flick_l_b.wav": NoteTypeSlideEndDirFlickLeftB,
	"slide_end_dir_flick_r_b.wav": NoteTypeSlideEndDirFlickRightB,
	"fever_note.wav":              bmsNTTap,
	"fever_note_flick.wav":        bmsNTFlick,
	"fever_note_slide_a.wav":      bmsNTSlideA,
	"fever_note_slide_end_a.wav":  bmsNTSlideEndA,
	"fever_note_slide_b.wav":      bmsNTSlideB,
	"fever_note_slide_end_b.wav":  NoteTypeSlideEndB,
	"fever_slide_a.wav":           bmsNTSlideA,
	"fever_slide_end_a.wav":       bmsNTSlideEndA,
	"fever_slide_b.wav":           bmsNTSlideB,
	"fever_slide_end_b.wav":       NoteTypeSlideEndB,
	"directional_fl_l.wav":        NoteTypeFlickLeft,
	"directional_fl_r.wav":        NoteTypeFlickRight,
	"add_long_dir_flick.wav":      NoteTypeAddLongDirFlick,
	"add_slide_dir_flick.wav":     NoteTypeAddSlideDirFlick,
	"cont_bezier_front_a.wav":     NoteTypeContBezierFrontA,
	"cont_bezier_front_b.wav":     NoteTypeContBezierFrontB,
	"cont_bezier_back_a.wav":      NoteTypeContBezierBackA,
	"cont_bezier_back_b.wav":      NoteTypeContBezierBackB,
	"long_end_dir_flick_l.wav":    NoteTypeLongEndDirFlickLeft,
	"long_end_dir_flick_r.wav":    NoteTypeLongEndDirFlickRight,
	"lane_change.wav":             NoteTypeLaneChange,
}

type bmsNoteType interface {
	String() string
	noteType() bmsBasicNoteType
}

func (n bmsBasicNoteType) String() string {
	switch n {
	case bmsNTTap:
		return "Tap"
	case bmsNTFlick:
		return "Flick"
	case NoteTypeFlickLeft:
		return "Flick Left"
	case NoteTypeFlickRight:
		return "Flick Right"
	case bmsNTSlideA:
		return "Slide A"
	case bmsNTSlideEndA:
		return "Slide End A"
	case bmsNTSlideEndFlickA:
		return "Slide End Flick A"
	case NoteTypeSlideEndDirFlickLeftA:
		return "Slide End Dir Flick Left A"
	case NoteTypeSlideEndDirFlickRightA:
		return "Slide End Dir Flick Right A"
	case bmsNTSlideB:
		return "Slide B"
	case NoteTypeSlideEndB:
		return "Slide End B"
	case NoteTypeSlideEndFlickB:
		return "Slide End Flick B"
	case NoteTypeSlideEndDirFlickLeftB:
		return "Slide End Dir Flick Left B"
	case NoteTypeSlideEndDirFlickRightB:
		return "Slide End Dir Flick Right B"
	case NoteTypeLongEndDirFlickLeft:
		return "Long End Dir Flick Left"
	case NoteTypeLongEndDirFlickRight:
		return "Long End Dir Flick Right"
	default:
		return "Unknown"
	}
}

func (n bmsBasicNoteType) noteType() bmsBasicNoteType {
	return n
}

type bmsSpecialSlideNoteType struct {
	mark   string
	offset float64
}

func newBMSSpecialSlideNoteType(name string) (*bmsSpecialSlideNoteType, error) {
	re := regexp.MustCompile(`slide_(.)_(L|R)S(\d\d)\.wav`)
	subs := re.FindStringSubmatch(name)
	if len(subs) < 3 {
		return &bmsSpecialSlideNoteType{}, fmt.Errorf("not a special slide note type")
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

	return &bmsSpecialSlideNoteType{
		mark:   mark,
		offset: offset,
	}, nil
}

func (n *bmsSpecialSlideNoteType) String() string {
	return fmt.Sprintf("Slide Special %s", n.mark)
}

func (n *bmsSpecialSlideNoteType) noteType() bmsBasicNoteType {
	switch n.mark {
	case "a":
		return bmsNTSlideA
	case "b":
		return bmsNTSlideB
	default:
		return bmsNTTap
	}
}

func noteTypeOf(wav string) (bmsNoteType, error) {
	basicType, ok := wavNoteTypeMap[wav]
	if ok {
		return basicType, nil
	}

	note, err := newBMSSpecialSlideNoteType(wav)
	if err == nil {
		return note, nil
	}

	return bmsNTTap, fmt.Errorf("unknown wav: %s", wav)
}

var simpleNoteChannels = []*bmsChannel{
	{bmsCKNote, '6'},
	{bmsCKNote, '1'},
	{bmsCKNote, '2'},
	{bmsCKNote, '3'},
	{bmsCKNote, '4'},
	{bmsCKNote, '5'},
	{bmsCKNote, '8'},
}

var channelsMap = map[string]*bmsChannel{
	"16": simpleNoteChannels[0],
	"11": simpleNoteChannels[1],
	"12": simpleNoteChannels[2],
	"13": simpleNoteChannels[3],
	"14": simpleNoteChannels[4],
	"15": simpleNoteChannels[5],
	"18": simpleNoteChannels[6],
}

func channelOf(rawChannel string) (*bmsChannel, error) {
	bytes := []byte(rawChannel)
	if len(bytes) != 2 {
		return nil, fmt.Errorf("unknown raw channel: %s", rawChannel)
	}

	var kind bmsChannelKind
	switch bytes[0] {
	case '0':
		kind = bmsCKConfig
	case '1':
		kind = bmsCKNote
	case '3':
		kind = bmsCKSpecial
	case '5':
		kind = bmsCKHold
	default:
		return nil, fmt.Errorf("unknown channel kind: %c", bytes[0])
	}

	if c, ok := channelsMap[rawChannel]; ok {
		return c, nil
	}

	c := &bmsChannel{
		kind: kind,
		typ:  bytes[1],
	}

	channelsMap[rawChannel] = c
	return c, nil
}

type parsedNoteEvent struct {
	channel  *bmsChannel
	noteType bmsNoteType
	aux      int
}

func (ev *parsedNoteEvent) String() string {
	return fmt.Sprintf("%s(@ %s)", ev.noteType, ev.channel)
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

	rawBpmEvents := map[float64]float64{}

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
			bpm, err := strconv.ParseFloat(value, 64)
			if err != nil {
				log.Fatalf("failed to parse value of #BPM(%s), err: %+v", value, err)
			}
			rawBpmEvents[0] = bpm // tick = 0时，bpm为初始bpm
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

	type rawNoteEvent struct {
		channel *bmsChannel
		wav     string
	}
	rawNoteEvents := map[float64][]*rawNoteEvent{}

	// 第一步：统计所有BPM事件，同时收集所有音符的wav数据，但不进行任何处理
	for lineNumber := 0; len(lines) != 0; lineNumber++ {
		line := lines[0]
		lines = lines[1:]

		events, _, err := parseDataLine(line)
		if err == errInvalidDataLineFormat {
			continue
		} else if err != nil {
			log.Fatalf("Failed to parse line #%d %s: %s", lineNumber, line, err)
		}

		for _, ev := range events {
			tick := ev.Tick()
			channel, err := channelOf(ev.Common.Channel)
			if err != nil {
				log.Fatalf("Failed to parse channel %s: %s", ev.Common.Channel, err)
			}

			if channel.kind == bmsCKConfig {
				switch channel.typ {
				case bmsCTBPMChange:
					value, err := strconv.ParseInt(ev.Type, 16, 64)
					if err != nil {
						log.Fatalf("Failed to parse value of line #%d bpm(%s), err: %+v", lineNumber, ev.Type, err)
					}

					rawBpmEvents[tick] = float64(value)
				case bmsCTExtendedBPM:
					rawBpmEvents[tick] = extendedBPM[ev.Type]
				}
			} else {
				if _, ok := rawNoteEvents[tick]; !ok {
					rawNoteEvents[tick] = nil
				}

				wav, ok := wavs[ev.Type]
				if !ok {
					rawNoteEvents[tick] = append(rawNoteEvents[tick], &rawNoteEvent{
						channel: channel,
					})
					continue
				}

				rawNoteEvents[tick] = append(rawNoteEvents[tick], &rawNoteEvent{
					channel: channel,
					wav:     wav,
				})
			}
		}
	}

	// 第二步：统计所有bpm事件，建立tick -> seconds转换表
	bpmTicks := utils.SortedKeysOf(rawBpmEvents)
	type bpmEvent struct {
		tick    float64
		bpm     float64
		seconds float64
	}
	bpmTable := []*bpmEvent{}

	lastTick := 0.0
	secStart := 0.0
	lastBpm := rawBpmEvents[0]
	for _, tick := range bpmTicks {
		secStart += barLength * 60 / lastBpm * (tick - lastTick)
		bpm := rawBpmEvents[tick]
		bpmTable = append(bpmTable, &bpmEvent{
			tick:    tick,
			bpm:     bpm,
			seconds: secStart,
		})

		lastTick = tick
		lastBpm = bpm
	}

	secondsOf := func(tick float64) float64 {
		idx, found := slices.BinarySearchFunc(bpmTable, tick, func(e *bpmEvent, t float64) int {
			return cmp.Compare(e.tick, t)
		})
		if !found {
			idx--
		}
		bpmInfo := bpmTable[idx]
		return bpmInfo.seconds + barLength*60/bpmInfo.bpm*(tick-bpmInfo.tick)
	}

	// 第三步：将rawNoteEvents初步转换为parsedNoteEvents
	// 主要是为了将wav解析到音符类型，以及将tick转换为seconds
	// 在这一步后，时间单位将统一为秒
	parsedNoteEvents := map[float64][]*parsedNoteEvent{}
	directionalFlickSeconds := map[float64][]byte{}
	noteTicks := utils.SortedKeysOf(rawNoteEvents)
	noteSeconds := []float64{}
	for _, tick := range noteTicks {
		evs := rawNoteEvents[tick]
		seconds := secondsOf(tick)
		noteSeconds = append(noteSeconds, seconds)
		parsedEvents := []*parsedNoteEvent{}
		for _, ev := range evs {
			noteType, err := noteTypeOf(ev.wav)
			if err != nil {
				log.Warnf("Unknown wav at channel %s, time: %s: %+v", ev.channel, utils.FormatSeconds(seconds), err)
				noteType = bmsNTTap
			}
			parsedEvents = append(parsedEvents, &parsedNoteEvent{
				channel:  ev.channel,
				noteType: noteType,
			})

			// 收集带方向的滑动音符信息，以便在下一步中将其合并
			if noteType == NoteTypeFlickLeft || noteType == NoteTypeFlickRight {
				if _, ok := directionalFlickSeconds[seconds]; !ok {
					directionalFlickSeconds[seconds] = make([]byte, 7)
				}
				v := directionalFlickSeconds[seconds]
				if noteType == NoteTypeFlickLeft {
					v[ev.channel.trackID()] = '<'
				} else {
					v[ev.channel.trackID()] = '>'
				}
			}
		}
		parsedNoteEvents[seconds] = parsedEvents
	}

	// 第四步：合并相邻的同一方向的滑动按键为一个，比如>>>可以视作一个滑动长度为3的滑键
	for seconds, v := range directionalFlickSeconds {
		start := -1
		length := 0
		newParsedEvents := []*parsedNoteEvent{}
		for i, c := range append(v, 0) {
			if c == '>' {
				if start == -1 {
					start = i
					length = 1
				} else {
					length++
				}
			} else {
				if start != -1 {
					newParsedEvents = append(newParsedEvents, &parsedNoteEvent{
						channel:  simpleNoteChannels[start],
						noteType: NoteTypeFlickRight,
						aux:      length,
					})
					start = -1
					length = 0
				}
			}
		}

		rev := append([]byte{0}, v...)
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
					newParsedEvents = append(newParsedEvents, &parsedNoteEvent{
						channel:  simpleNoteChannels[start],
						noteType: NoteTypeFlickLeft,
						aux:      length,
					})
					start = -1
					length = 0
				}
			}
		}

		for _, ev := range parsedNoteEvents[seconds] {
			if ev.noteType != NoteTypeFlickLeft && ev.noteType != NoteTypeFlickRight {
				newParsedEvents = append(newParsedEvents, ev)
			}
		}

		parsedNoteEvents[seconds] = newParsedEvents
	}

	// 第五步：将每个音符事件转换为手法规划器支持的结构（star）
	finalEvents := []*star{}
	holdTracks := [7]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}
	var slideA, slideB *star

	for _, sec := range noteSeconds {
		log.Debugln("TIME:", utils.FormatSeconds(sec))
		events := parsedNoteEvents[sec]
		slices.SortFunc(events, func(a, b *parsedNoteEvent) int {
			return -cmp.Compare(a.noteType.noteType(), b.noteType.noteType())
		})

		for _, ev := range events {
			switch ev.channel.kind {
			case bmsCKNote:
				trackID := float64(ev.channel.trackID()) / 6
				switch ev.noteType {
				// normal note
				case bmsNTTap:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							markAsTap())
				// flick note
				case bmsNTFlick:
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
				case bmsNTSlideA:
					if slideA == nil {
						slideA = newStar(sec, trackID, 1.0/6).
							markAsTap().
							markAsHead()
					} else {
						slideA = newStar(sec, trackID, 1.0/6).
							chainsAfter(slideA)
					}
				case bmsNTSlideEndA:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideA).
							markAsEnd())
					slideA = nil
				case bmsNTSlideEndFlickA:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideA).
							flickToIfOk(true, 90).
							markAsEnd())
					slideA = nil
				case NoteTypeSlideEndDirFlickLeftA:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideA).
							flickToIfOk(true, 180).
							markAsEnd())
					slideA = nil
				case NoteTypeSlideEndDirFlickRightA:
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideA).
							flickToIfOk(true, 0).
							markAsEnd())
					slideA = nil
				// slide b
				case bmsNTSlideB:
					if slideB == nil {
						slideB = newStar(sec, trackID, 1.0/6).
							markAsTap().
							markAsHead()
					} else {
						slideB = newStar(sec, trackID, 1.0/6).
							chainsAfter(slideB)
					}
				case NoteTypeSlideEndB:
					if slideB == nil {
						log.Fatalf("%s at channel %s time %s has no slide head", ev.noteType, ev.channel, utils.FormatSeconds(sec))
					}
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideB).
							markAsEnd())
					slideB = nil
				case NoteTypeSlideEndFlickB:
					if slideB == nil {
						log.Fatalf("%s at channel %s time %s has no slide head", ev.noteType, ev.channel, utils.FormatSeconds(sec))
					}
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideB).
							flickToIfOk(true, 90).
							markAsEnd())
					slideB = nil
				case NoteTypeSlideEndDirFlickLeftB:
					if slideB == nil {
						log.Fatalf("%s at channel %s time %s has no slide head", ev.noteType, ev.channel, utils.FormatSeconds(sec))
					}
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideB).
							flickToIfOk(true, 180).
							markAsEnd())
					slideB = nil
				case NoteTypeSlideEndDirFlickRightB:
					if slideB == nil {
						log.Fatalf("%s at channel %s time %s has no slide head", ev.noteType, ev.channel, utils.FormatSeconds(sec))
					}
					finalEvents = append(
						finalEvents,
						newStar(sec, trackID, 1.0/6).
							chainsAfter(slideB).
							flickToIfOk(true, 0).
							markAsEnd())
					slideB = nil

				case NoteTypeAddLongDirFlick:
					// do nothing
				case NoteTypeAddSlideDirFlick:
					// do nothing

				case NoteTypeContBezierFrontA:
				case NoteTypeContBezierFrontB:
				case NoteTypeContBezierBackA:
				case NoteTypeContBezierBackB:

				case NoteTypeLaneChange:
					// [TODO] 解析大小键
					// update: 算了，等下次愚人节再说吧

				// unknown
				default:
					log.Warnf("unknown note type %s at track %f time %s\n", ev.noteType, trackID, utils.FormatSeconds(sec))
				}
			case bmsCKHold:
				trackID := ev.channel.trackID()
				trackX := float64(trackID) / 6
				switch ev.noteType {
				case bmsNTTap:
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
				case bmsNTFlick:
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
				case NoteTypeLongEndDirFlickLeft:
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
							flickToIfOk(true, 180).
							markAsEnd())
					holdTracks[trackID] = math.NaN()
				case NoteTypeLongEndDirFlickRight:
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
							flickToIfOk(true, 0).
							markAsEnd())
					holdTracks[trackID] = math.NaN()
				default:
					log.Warnf("unknown note type %s at track %d, time %f s\n", ev.noteType, trackID, sec)
				}
			case bmsCKSpecial:
				trackID := float64(ev.channel.trackID()) / 6
				switch nt := ev.noteType.(type) {
				case *bmsSpecialSlideNoteType:
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
				case bmsBasicNoteType:
					switch nt {
					case bmsNTSlideA:
						if slideA == nil {
							slideA = newStar(sec, trackID, 1.0/6).
								markAsTap().
								markAsHead()
						} else {
							slideA = newStar(sec, trackID, 1.0/6).
								chainsAfter(slideA)
						}
					case bmsNTSlideB:
						if slideB == nil {
							slideB = newStar(sec, trackID, 1.0/6).
								markAsTap().
								markAsHead()
						} else {
							slideB = newStar(sec, trackID, 1.0/6).
								chainsAfter(slideB)
						}
					default:
						log.Warnf("%s should not appear at channel %d, time %f s", ev.noteType, ev.channel, sec)
					}
				default:
					log.Warnf("%s should not appear at channel %d, time %f s", ev.noteType, ev.channel, sec)
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
