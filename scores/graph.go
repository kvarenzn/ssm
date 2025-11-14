// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package scores

type Graph[T comparable] struct {
	edges    map[T]map[T]struct{}
	vertices map[T]struct{}
	size     int
}

func NewGraph[T comparable]() *Graph[T] {
	return &Graph[T]{
		edges:    map[T]map[T]struct{}{},
		vertices: map[T]struct{}{},
	}
}

func (g *Graph[T]) addNode(node T) {
	if _, ok := g.vertices[node]; ok {
		return
	}

	g.vertices[node] = struct{}{}
	if _, ok := g.edges[node]; !ok {
		g.edges[node] = map[T]struct{}{}
	}
}

func (g *Graph[T]) Order() int {
	return len(g.vertices)
}

func (g *Graph[T]) Size() int {
	return g.size
}

func (g *Graph[T]) Connected(a, b T) bool {
	if _, ok := g.vertices[a]; !ok {
		return false
	}

	if _, ok := g.vertices[b]; !ok {
		return false
	}

	if _, ok := g.edges[a]; !ok {
		return false
	}

	_, ok := g.edges[a][b]
	return ok
}

func (g *Graph[T]) Connect(a, b T) bool {
	g.addNode(a)
	g.addNode(b)

	newEdge := false
	if _, ok := g.edges[a][b]; !ok {
		g.edges[a][b] = struct{}{}
		newEdge = true
	}

	if _, ok := g.edges[b][a]; !ok {
		g.edges[b][a] = struct{}{}
		newEdge = true
	}

	if newEdge {
		g.size++
	}

	return newEdge
}

func (g *Graph[T]) Disconnect(a, b T) bool {
	if !g.Connected(a, b) {
		return false
	}

	delete(g.edges[a], b)
	if len(g.edges[a]) == 0 {
		delete(g.edges, a)
		delete(g.vertices, a)
	}

	delete(g.edges[b], a)
	if len(g.edges[b]) == 0 {
		delete(g.edges, b)
		delete(g.vertices, b)
	}

	g.size--

	return true
}

func GrabFrom[T comparable](g map[T]map[T]struct{}, start T) *Graph[T] {
	result := NewGraph[T]()

	visited := map[T]struct{}{}
	var visit func(node T)
	visit = func(node T) {
		if _, ok := visited[node]; ok {
			return
		}

		visited[node] = struct{}{}
		for k := range g[node] {
			result.Connect(node, k)
			visit(k)
		}
	}

	visit(start)

	return result
}
