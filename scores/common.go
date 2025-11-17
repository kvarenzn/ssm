// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"errors"
	"iter"
	"math"
	"strconv"
	"strings"
)

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

type noteKind uint8

const (
	tapNote noteKind = iota
	dragNote
	flickNote
	throwNote
	slideNote
)

type star struct {
	seconds      float64
	track, width float64
	direction    float64 // flick direction, NaN if this is not a flick

	head, prev, next *star
}

type Chart []*star

func (s *star) kind() noteKind {
	if s.isSlide() {
		return slideNote
	} else if s.isTap() && s.isFlick() {
		return flickNote
	} else if s.isTap() {
		return tapNote
	} else if s.isFlick() {
		return throwNote
	} else {
		return dragNote
	}
}

func (s *star) start() float64 {
	if s.isSlide() {
		return s.head.seconds
	}

	return s.seconds
}

func (s *star) isSlide() bool {
	return s.head != nil
}

func (s *star) isFlick() bool {
	return !math.IsNaN(s.direction)
}

func (s *star) isTap() bool {
	return s.prev == s
}

func (s *star) isEnd() bool {
	return s.next == s
}

func (s *star) isLast() bool {
	return s.next == nil || s.isEnd()
}

func (s *star) iterSlide() iter.Seq[*star] {
	return func(yield func(*star) bool) {
		cur := s.head
		if !yield(cur) {
			return
		}

		if cur.isLast() {
			return
		}

		for {
			cur = cur.next
			if !yield(cur) {
				return
			}

			if cur.isLast() {
				return
			}
		}
	}
}

func (s *star) chainsAfter(prev *star) *star {
	s.prev = prev
	s.head = prev.head
	prev.next = s
	return s
}

func (s *star) flickToIfOk(ok bool, deg int) *star {
	if ok {
		s.direction = float64(deg) * math.Pi / 180
		s.markAsEnd()
	}
	return s
}

func (s *star) markAsHead() *star {
	s.head = s
	return s
}

func (s *star) markAsTap() *star {
	s.prev = s
	return s
}

func (s *star) markAsEnd() *star {
	s.next = s
	return s
}

func (s *star) delta(factor float64) (float64, float64) {
	return math.Cos(s.direction) * factor, math.Sin(s.direction) * factor
}

func newStar(seconds, track, width float64) *star {
	return &star{
		seconds:   seconds,
		track:     track,
		width:     width,
		direction: math.NaN(),
	}
}

func quantify(time float64) int64 {
	return int64(math.Round(time * 1000))
}

type vec2f struct {
	x float64
	y float64
}

func (v *vec2f) distanceTo(o *vec2f) float64 {
	return math.Hypot(o.x-v.x, o.y-v.y)
}

func (v *vec2f) equals(o *vec2f) bool {
	return v.x == o.x && v.y == o.y
}

func (v *vec2f) mul(f float64) *vec2f {
	return &vec2f{
		x: v.x * f,
		y: v.y * f,
	}
}

func (v *vec2f) add(o *vec2f) *vec2f {
	return &vec2f{
		x: v.x + o.x,
		y: v.y + o.y,
	}
}

func newVec2f(x, y float64) *vec2f {
	return &vec2f{
		x: x,
		y: y,
	}
}

func pointToLine(p, a, b *vec2f) float64 {
	if a.equals(b) {
		return p.distanceTo(a)
	}

	area := math.Abs((b.x-a.x)*(a.y-p.y) - (a.x-p.x)*(b.y-a.y))
	length := a.distanceTo(b)
	return area / length
}

func isFlat(p0, p1, p2, p3 *vec2f, epsilon float64) bool {
	d1 := pointToLine(p1, p0, p3)
	d2 := pointToLine(p2, p0, p3)
	return max(d1, d2) <= epsilon
}

func interp(a, b *vec2f, t float64) *vec2f {
	return (a.mul(1 - t)).add(b.mul(t))
}

func deCasteljauSplit(p0, p1, p2, p3 *vec2f, t float64) ([]*vec2f, []*vec2f) {
	p01 := interp(p0, p1, t)
	p12 := interp(p1, p2, t)
	p23 := interp(p2, p3, t)

	p012 := interp(p01, p12, t)
	p123 := interp(p12, p23, t)

	p0123 := interp(p012, p123, t)

	return []*vec2f{p0, p01, p012, p0123}, []*vec2f{p0123, p123, p23, p3}
}

func adaptiveBezierSubdivision(p0, p1, p2, p3 *vec2f, epsilon float64, out *[]*vec2f) {
	if isFlat(p0, p1, p2, p3, epsilon) {
		*out = append(*out, p3)
	} else {
		left, right := deCasteljauSplit(p0, p1, p2, p3, 0.5)

		adaptiveBezierSubdivision(left[0], left[1], left[2], left[3], epsilon, out)
		adaptiveBezierSubdivision(right[0], right[1], right[2], right[3], epsilon, out)
	}
}

func bezierToPolyline(p0, p1, p2, p3 *vec2f, epsilon float64) []*vec2f {
	out := []*vec2f{p0}
	adaptiveBezierSubdivision(p0, p1, p2, p3, epsilon, &out)
	return out
}
