package main

import (
	"cmp"
	"sort"
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
	nodes  []Node[T]
	cloves []Clove[T]
	maxID  int
}

func (nds *Nodes[T]) AddEvent(start, end T) {
	nds.cloves = append(nds.cloves, Clove[T]{
		ID:   nds.maxID,
		Type: NDStart,
		Tick: start,
	}, Clove[T]{
		ID:   nds.maxID,
		Type: NDEnd,
		Tick: end,
	})
	nds.nodes = append(nds.nodes, Node[T]{
		Start: start,
		End:   end,
	})

	nds.maxID++
}

func (nds *Nodes[T]) Colorize() []int {
	// sort time nodes
	nodes := nds.cloves
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Tick == nodes[j].Tick {
			return nodes[i].Type < nodes[j].Type
		}

		return nodes[i].Tick < nodes[j].Tick
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

	colored := map[int]int{}
	for i := 0; i < nds.maxID; i++ {
		if _, ok := colored[i]; ok {
			continue
		}

		// find connected graph
		if _, ok := conflicts[i]; !ok {
			colored[i] = 0
			continue
		}

		visited := map[int]struct{}{} // id -> index
		degrees := []struct {
			ID     int
			Degree int
		}{}
		var visit func(id int)
		visit = func(id int) {
			if _, ok := visited[id]; ok {
				return
			}

			visited[id] = struct{}{}
			degrees = append(degrees, struct {
				ID     int
				Degree int
			}{id, len(conflicts[id])})

			for k := range conflicts[id] {
				visit(k)
			}
		}
		visit(i)

		// welsh-powell algorithm
		sort.Slice(degrees, func(i, j int) bool {
			if degrees[i].Degree == degrees[j].Degree {
				return nds.nodes[degrees[i].ID].Start < nds.nodes[degrees[j].ID].Start
			}

			return degrees[i].Degree > degrees[j].Degree
		})

		uncolored := make([]int, len(degrees))
		for i, d := range degrees {
			uncolored[i] = d.ID
		}

		for c := 0; len(uncolored) > 0; c++ {
			colored[uncolored[0]] = c // color the first node as c
			uncolored = uncolored[1:]
			next := []int{}
			for _, id := range uncolored {
				ok := true
				for k := range conflicts[id] {
					if cc, yes := colored[k]; yes && cc == c {
						ok = false
						break
					}
				}
				if ok {
					colored[id] = c
				} else {
					next = append(next, id)
				}
			}

			uncolored = next
		}
	}
	result := []int{}
	for i := 0; i < nds.maxID; i++ {
		result = append(result, colored[i])
	}

	return result
}
