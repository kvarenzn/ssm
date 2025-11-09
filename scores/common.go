// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"errors"
	"strconv"
	"strings"
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

	str := line[colonIndex+1:]
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
