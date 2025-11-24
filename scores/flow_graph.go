package scores

import (
	"math"

	"github.com/kvarenzn/ssm/utils"
)

type flowGraph struct {
	heads      []int
	tos        []int
	nexts      []int
	capacities []int
	costs      []float64
	edgeCount  int
	nodeCount  int

	parentEdge   []int
	dists        []float64
	queue        *utils.Queue[int]
	inQueue      []bool
	updateCounts []int
}

func newFlowGraph(nodeCount int) *flowGraph {
	heads := make([]int, nodeCount)
	for i := range nodeCount {
		heads[i] = -1
	}
	return &flowGraph{
		heads:        heads,
		nodeCount:    nodeCount,
		parentEdge:   make([]int, nodeCount),
		dists:        make([]float64, nodeCount),
		queue:        utils.NewQueue[int](nodeCount),
		inQueue:      make([]bool, nodeCount),
		updateCounts: make([]int, nodeCount),
	}
}

func (g *flowGraph) addEdge(u, v int, capacity int, cost float64) {
	g.tos = append(g.tos, v)
	g.capacities = append(g.capacities, capacity)
	g.costs = append(g.costs, cost)
	g.nexts = append(g.nexts, g.heads[u])
	g.heads[u] = g.edgeCount
	g.edgeCount++

	g.tos = append(g.tos, u)
	g.capacities = append(g.capacities, 0)
	g.costs = append(g.costs, -cost)
	g.nexts = append(g.nexts, g.heads[v])
	g.heads[v] = g.edgeCount
	g.edgeCount++
}

func (g *flowGraph) spfa(source, sink int) bool {
	const epsilon = 1e-9
	for i := range g.nodeCount {
		g.dists[i] = math.Inf(0)
		g.parentEdge[i] = -1
		g.inQueue[i] = false
		g.updateCounts[i] = 0
	}

	g.queue.Push(source)
	g.dists[source] = 0

	for !g.queue.Empty() {
		u, _ := g.queue.Pop()
		g.inQueue[u] = false

		for edge := g.heads[u]; edge != -1; edge = g.nexts[edge] {
			v := g.tos[edge]
			cost := g.costs[edge]
			if g.capacities[edge] <= 0 || g.dists[v] <= g.dists[u]+cost+epsilon {
				continue
			}

			g.dists[v] = g.dists[u] + cost
			g.parentEdge[v] = edge
			if !g.inQueue[v] {
				g.queue.Push(v)
				g.inQueue[v] = true

				g.updateCounts[v]++
				if g.updateCounts[v] > g.nodeCount {
					panic("?")
				}
			}
		}
	}

	return !math.IsInf(g.dists[sink], 0)
}

func (g *flowGraph) mcmf(source, sink int) ([]*struct{ from, to int }, int) {
	const flow = 1

	maxFlow := 0
	for g.spfa(source, sink) {
		current := sink
		maxFlow += flow
		for current != source {
			idx := g.parentEdge[current]
			g.capacities[idx] -= flow
			g.capacities[idx^1] += flow

			current = g.tos[idx^1]
		}
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
