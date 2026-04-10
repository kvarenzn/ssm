// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

import (
	"math"

	"github.com/kvarenzn/ssm/log"
	"github.com/kvarenzn/ssm/utils"
)

type flowGraph struct {
	heads      []int
	tos        []int
	nexts      []int
	capacities []int
	costs      []int64
	potential  []int64
	parents    []struct {
		from int
		edge int
	}
	edgeCount int
	nodeCount int

	dists []int64
}

const inf64 = math.MaxInt64

func newFlowGraph(nodeCount int) *flowGraph {
	heads := make([]int, nodeCount)
	for i := range nodeCount {
		heads[i] = -1
	}
	return &flowGraph{
		heads:     heads,
		potential: make([]int64, nodeCount),
		parents: make([]struct {
			from int
			edge int
		}, nodeCount),
		nodeCount: nodeCount,
		dists:     make([]int64, nodeCount),
	}
}

func (g *flowGraph) addEdge(from, to int, capacity int, cost float64) {
	const scaleFactor = 1e7
	g.tos = append(g.tos, to)
	g.capacities = append(g.capacities, capacity)
	g.costs = append(g.costs, int64(math.Round(cost*scaleFactor)))
	g.nexts = append(g.nexts, g.heads[from])
	g.heads[from] = g.edgeCount
	g.edgeCount++

	g.tos = append(g.tos, from)
	g.capacities = append(g.capacities, 0)
	g.costs = append(g.costs, -int64(math.Round(cost*scaleFactor)))
	g.nexts = append(g.nexts, g.heads[to])
	g.heads[to] = g.edgeCount
	g.edgeCount++
}

func (g *flowGraph) spfa(source int) {
	inQueue := make([]bool, g.nodeCount)
	for i := range g.nodeCount {
		g.potential[i] = inf64
	}

	q := utils.NewQueue[int](g.nodeCount)
	q.Push(source)
	inQueue[source] = true
	g.potential[source] = 0

	for !q.Empty() {
		from, err := q.Pop()
		if err != nil {
			panic(err)
		}
		inQueue[from] = false

		for edge := g.heads[from]; edge != -1; edge = g.nexts[edge] {
			to := g.tos[edge]
			cost := g.costs[edge]
			if g.capacities[edge] > 0 && g.potential[from]+cost < g.potential[to] {
				g.potential[to] = g.potential[from] + cost
				if !inQueue[to] {
					q.Push(to)
					inQueue[to] = true
				}
			}
		}
	}
}

func (g *flowGraph) dijkstra(source, sink int) bool {
	pq := utils.NewPriorityQueue[int64, int](nil)
	for i := range g.nodeCount {
		g.dists[i] = inf64
	}
	g.dists[source] = 0
	pq.Push(0, source)

	for !pq.Empty() {
		from, dist := pq.Pop()
		if dist != g.dists[from] {
			continue
		}

		for edge := g.heads[from]; edge != -1; edge = g.nexts[edge] {
			if g.capacities[edge] <= 0 {
				continue
			}

			to := g.tos[edge]
			newCost := g.costs[edge] + g.potential[from] - g.potential[to]
			if newCost < 0 {
				log.Dief("cost = %d, from.potential = %d, to.potential = %d, newCost = %d", g.costs[edge], g.potential[from], g.potential[to], newCost)
			}
			if g.dists[to] > g.dists[from]+newCost {
				g.dists[to] = g.dists[from] + newCost
				g.parents[to].from = from
				g.parents[to].edge = edge
				pq.Push(g.dists[to], to)
			}
		}
	}

	if g.dists[sink] == inf64 {
		return false
	}

	for i := range g.nodeCount {
		if g.dists[i] != inf64 {
			g.potential[i] += g.dists[i]
		}
	}

	return true
}

func (g *flowGraph) mc(source, sink int) ([]*struct{ from, to int }, int) {
	maxFlow := 0
	g.spfa(source)
	for g.dijkstra(source, sink) {
		if g.potential[sink] >= 0 {
			break
		}

		flow := math.MaxInt32
		for i := sink; i != source; i = g.parents[i].from {
			flow = min(flow, g.capacities[g.parents[i].edge])
		}

		for i := sink; i != source; i = g.parents[i].from {
			g.capacities[g.parents[i].edge] -= flow
			g.capacities[g.parents[i].edge^1] += flow
		}

		maxFlow += flow
	}

	connections := []*struct{ from, to int }{}
	for i := 0; i < len(g.tos); i += 2 {
		if g.capacities[i^1] > 0 {
			connections = append(connections, &struct {
				from int
				to   int
			}{g.tos[i^1], g.tos[i]})
		}
	}

	return connections, maxFlow
}
