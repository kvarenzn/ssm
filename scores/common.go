// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/kvarenzn/ssm/common"
)

type BPMEvent struct {
	Tick float64
	BPM  float64
}

type SimpleRawEvent struct {
	Type      string
	Numerator int
	Common    *SimpleRawEventCommon
}

type SimpleRawEventCommon struct {
	Channel     string
	Measure     int
	Denominator int
}

func (e *SimpleRawEvent) Tick() float64 {
	return float64(e.Common.Measure) + float64(e.Numerator)/float64(e.Common.Denominator)
}

func splitString2(s string) []string {
	var result []string
	for i := 0; i < len(s); i += 2 {
		result = append(result, s[i:i+2])
	}
	return result
}

var errInvalidDataLineFormat = errors.New("Invalid data line format")

func parseDataLine(line string) ([]*SimpleRawEvent, *SimpleRawEventCommon, error) {
	line = strings.TrimSpace(line)
	hashIndex := strings.Index(line, "#")
	if hashIndex == -1 {
		return nil, nil, errInvalidDataLineFormat
	}

	line = line[hashIndex+1:]

	colonIndex := strings.Index(line, ":")
	if colonIndex == -1 || colonIndex < 5 {
		return nil, nil, errInvalidDataLineFormat
	}

	measure, err := strconv.ParseInt(line[:3], 10, 64)
	if err != nil {
		return nil, nil, err
	}

	str := strings.TrimSpace(line[colonIndex+1:])
	common := &SimpleRawEventCommon{
		Measure:     int(measure),
		Channel:     line[3:colonIndex],
		Denominator: len(str) / 2,
	}

	result := []*SimpleRawEvent{}

	for index, data := range splitString2(str) {
		if data == "00" {
			continue
		}

		result = append(result, &SimpleRawEvent{
			Type:      data,
			Numerator: index,
			Common:    common,
		})
	}

	return result, common, nil
}

type VTEGenerateConfig struct {
	TapDuration         int64
	FlickDuration       int64
	FlickReportInterval int64
	SlideReportInterval int64
}

// need TOUCH_DOWN
type TapEvent struct {
	Seconds float64
	Track   float64
	Width   float64
}

// need TOUCH_EXIST
type DragEvent struct {
	Seconds float64
	Track   float64
	Width   float64

	ignored  bool
	doNotTap bool
}

// need TOUCH_DOWN + TOUCH_SWIPE + TOUCH_UP
type FlickEvent struct {
	Seconds float64
	Track   float64
	Offset  common.Point2D
	Width   float64
}

// need TOUCH_EXIST + TOUCH_SWIPE + TOUCH_UP
type ThrowEvent struct {
	Seconds float64
	Track   float64
	Offset  common.Point2D
	Width   float64

	doNotTap bool
}

// need TOUCH_DOWN + TOUCH_UP
type HoldEvent struct {
	Seconds    float64
	EndSeconds float64
	Track      float64
	FlickEnd   bool
	Width      float64
}

type TraceItem struct {
	Seconds float64
	Track   float64
	Width   float64
}

// need TOUCH_DOWN + TOUCH_MOVE + TOUCH_UP
type SlideEvent struct {
	Seconds  float64
	Track    float64
	Trace    []*TraceItem
	FlickEnd bool
	Mark     uint8
	Width    float64
}

type NoteEvent interface {
	Start() float64
}

func (e *TapEvent) Start() float64 {
	return e.Seconds
}

func (e *DragEvent) Start() float64 {
	return e.Seconds
}

func (e *FlickEvent) Start() float64 {
	return e.Seconds
}

func (e *ThrowEvent) Start() float64 {
	return e.Seconds
}

func (e *HoldEvent) Start() float64 {
	return e.Seconds
}

func (e *SlideEvent) Start() float64 {
	return e.Seconds
}

func quantify(time float64) int64 {
	return int64(math.Round(time * 1000))
}
