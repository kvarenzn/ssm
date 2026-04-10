// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"cmp"
	"slices"
)

type slsEventKind uint8

const (
	slsChange slsEventKind = iota
	slsQuery
	slsEnd
)

type slsEvent[T cmp.Ordered, P, Q any] interface {
	kind() slsEventKind
	start() T
}

type slsChangeEvent[T cmp.Ordered, P, Q any] struct {
	id             int
	tick, nextTick T
	pos, nextPos   P
}

func (e *slsChangeEvent[T, P, Q]) kind() slsEventKind { return slsChange }
func (e *slsChangeEvent[T, P, Q]) start() T           { return e.tick }

type slsEndEvent[T cmp.Ordered, P, Q any] struct {
	id   int
	tick T
}

func (e *slsEndEvent[T, P, Q]) kind() slsEventKind { return slsEnd }
func (e *slsEndEvent[T, P, Q]) start() T           { return e.tick }

type slsQueryEvent[T cmp.Ordered, P, Q any] struct {
	id    int
	tick  T
	query Q
}

func (e *slsQueryEvent[T, P, Q]) kind() slsEventKind { return slsQuery }
func (e *slsQueryEvent[T, P, Q]) start() T           { return e.tick }

type sls[T cmp.Ordered, P, Q any] struct {
	events       []slsEvent[T, P, Q]
	traceCounter int
	queryCounter int

	interpolate func(tStart, tEnd, tCurrent T, start, end P) P
	satisfy     func(current P, query Q) bool
}

func (s *sls[T, P, Q]) AddTrace(trace []struct {
	T T
	P P
},
) {
	id := s.traceCounter
	if len(trace) < 2 {
		panic("?")
	}

	for i := range len(trace) - 1 {
		s.events = append(s.events, &slsChangeEvent[T, P, Q]{
			id:       id,
			tick:     trace[i].T,
			nextTick: trace[i+1].T,
			pos:      trace[i].P,
			nextPos:  trace[i+1].P,
		})
	}

	s.events = append(s.events, &slsEndEvent[T, P, Q]{
		id:   id,
		tick: trace[len(trace)-1].T,
	})

	s.traceCounter++
}

func (s *sls[T, P, Q]) AddQuery(tick T, query Q) {
	s.events = append(s.events, &slsQueryEvent[T, P, Q]{
		id:    s.queryCounter,
		tick:  tick,
		query: query,
	})

	s.queryCounter++
}

func (s *sls[T, P, Q]) Scan() []struct{ Query, Trace int } {
	slices.SortFunc(s.events, func(a, b slsEvent[T, P, Q]) int {
		ta, tb := a.start(), b.start()
		if ta != tb {
			return cmp.Compare(ta, tb)
		}

		return cmp.Compare(a.kind(), b.kind())
	})

	type status struct {
		tick, nextTick T
		pos, nextPos   P
	}

	traces := map[int]*status{}
	result := []struct{ Query, Trace int }{}

	for _, event := range s.events {
		switch event := event.(type) {
		case *slsChangeEvent[T, P, Q]:
			traces[event.id] = &status{
				tick:     event.tick,
				nextTick: event.nextTick,
				pos:      event.pos,
				nextPos:  event.nextPos,
			}
		case *slsEndEvent[T, P, Q]:
			delete(traces, event.id)
		case *slsQueryEvent[T, P, Q]:
			for id, trace := range traces {
				currentPos := s.interpolate(trace.tick, trace.nextTick, event.tick, trace.pos, trace.nextPos)
				if s.satisfy(currentPos, event.query) {
					result = append(result, struct {
						Query int
						Trace int
					}{event.id, id})
					break
				}
			}
		}
	}

	return result
}

func NewSLS[T cmp.Ordered, P, Q any](interpolate func(tStart, tEnd, tCurrent T, start, end P) P, satisfy func(current P, query Q) bool) *sls[T, P, Q] {
	return &sls[T, P, Q]{
		interpolate: interpolate,
		satisfy:     satisfy,
	}
}

func interpF64(tStart, tEnd, tCurrent, start, end float64) float64 {
	return start + (end-start)*(tCurrent-tStart)/(tEnd-tStart)
}

func satisfyF64(current float64, query *struct{ Min, Max float64 }) bool {
	return current >= query.Min && current <= query.Max
}

func NewSLSF64() *sls[float64, float64, *struct{ Min, Max float64 }] {
	return NewSLS(interpF64, satisfyF64)
}
