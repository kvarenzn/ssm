// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"cmp"
	"maps"
	"slices"
)

type cloveKind uint8

const (
	cloveStart cloveKind = iota
	cloveEnd
)

type clove[T cmp.Ordered] struct {
	id   int
	kind cloveKind
	tick T
}

type cloves[T cmp.Ordered] struct {
	ids    map[int]struct{}
	cloves []*clove[T]
}

func NewCloves[T cmp.Ordered]() *cloves[T] {
	return &cloves[T]{
		ids: map[int]struct{}{},
	}
}

func (nds *cloves[T]) AddEvent(id int, start, end T) {
	if _, ok := nds.ids[id]; ok {
		panic("Duplicated id")
	}

	nds.ids[id] = struct{}{}
	nds.cloves = append(nds.cloves, &clove[T]{
		id:   id,
		kind: cloveStart,
		tick: start,
	}, &clove[T]{
		id:   id,
		kind: cloveEnd,
		tick: end,
	})
}

func (nds *cloves[T]) Colorize() map[int]int {
	// sort time cloves
	slices.SortFunc(nds.cloves, func(a, b *clove[T]) int {
		if a.tick == b.tick {
			return cmp.Compare(a.kind, b.kind)
		}

		return cmp.Compare(a.tick, b.tick)
	})

	// find conflicts
	conflicts := map[int]map[int]struct{}{}
	addConflict := func(a, b int) {
		if _, ok := conflicts[a]; !ok {
			conflicts[a] = map[int]struct{}{}
		}

		if _, ok := conflicts[b]; !ok {
			conflicts[b] = map[int]struct{}{}
		}

		conflicts[a][b] = struct{}{}
		conflicts[b][a] = struct{}{}
	}
	currentEvents := map[int]struct{}{}
	for _, node := range nds.cloves {
		if node.kind == cloveStart {
			for id := range currentEvents {
				addConflict(id, node.id)
			}

			currentEvents[node.id] = struct{}{}
		} else {
			delete(currentEvents, node.id)
		}
	}

	// color
	colored := map[int]int{}
	for i := range nds.ids {
		if _, ok := colored[i]; ok {
			continue
		}

		// find connected graph
		if _, ok := conflicts[i]; !ok {
			colored[i] = 0
			continue
		}

		g := GrabFrom(conflicts, i)
		colors := SaturationLargestFirstGreedyColoring(g)

		maps.Copy(colored, colors)
	}

	return colored
}
