// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"cmp"
	"maps"
	"slices"
)

type NodeType uint8

const (
	NDStart NodeType = iota
	NDEnd
)

type Clove[T cmp.Ordered] struct {
	ID   int
	Type NodeType
	Tick T
}

type Node[T cmp.Ordered] struct {
	Start T
	End   T
}

type Nodes[T cmp.Ordered] struct {
	nodes  []*Node[T]
	cloves []*Clove[T]
	maxID  int
}

func (nds *Nodes[T]) AddEvent(start, end T) {
	nds.cloves = append(nds.cloves, &Clove[T]{
		ID:   nds.maxID,
		Type: NDStart,
		Tick: start,
	}, &Clove[T]{
		ID:   nds.maxID,
		Type: NDEnd,
		Tick: end,
	})
	nds.nodes = append(nds.nodes, &Node[T]{
		Start: start,
		End:   end,
	})

	nds.maxID++
}

func (nds *Nodes[T]) Colorize() []int {
	// sort time nodes
	nodes := nds.cloves
	slices.SortFunc(nodes, func(a, b *Clove[T]) int {
		if a.Tick == b.Tick {
			return cmp.Compare(a.Type, b.Type)
		}

		return cmp.Compare(a.Tick, b.Tick)
	})
	nds.cloves = nodes

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
	for _, node := range nodes {
		if node.Type == NDStart {
			for id := range currentEvents {
				addConflict(id, node.ID)
			}

			currentEvents[node.ID] = struct{}{}
		} else {
			delete(currentEvents, node.ID)
		}
	}

	// color
	colored := map[int]int{}
	for i := range nds.maxID {
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
	result := []int{}
	for i := range nds.maxID {
		result = append(result, colored[i])
	}

	return result
}
