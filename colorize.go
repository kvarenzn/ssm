// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"cmp"
	"sort"

	"github.com/kvarenzn/ssm/dlx"
)

type Graph[T cmp.Ordered] struct {
	edges    map[T]map[T]struct{}
	vertices map[T]struct{}
	size     int
}

func NewGraph[T cmp.Ordered]() Graph[T] {
	return Graph[T]{
		edges:    map[T]map[T]struct{}{},
		vertices: map[T]struct{}{},
		size:     0,
	}
}

func (g *Graph[T]) Order() int {
	return len(g.vertices)
}

func (g *Graph[T]) Size() int {
	return g.size
}

func (g *Graph[T]) Connected(a, b T) bool {
	if a > b {
		a, b = b, a
	}

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
	if a > b {
		a, b = b, a
	}

	g.vertices[a] = struct{}{}
	g.vertices[b] = struct{}{}

	if _, ok := g.edges[a]; !ok {
		g.edges[a] = map[T]struct{}{
			b: {},
		}

		g.size++
		return true
	}

	if _, ok := g.edges[a][b]; !ok {
		g.edges[a][b] = struct{}{}
		g.size++
		return true
	}

	return false
}

func (g *Graph[T]) Disconnect(a, b T) bool {
	if a > b {
		a, b = b, a
	}

	m, ok := g.edges[a]
	if !ok {
		return false
	}

	if _, ok := m[b]; !ok {
		return false
	}

	delete(m, b)
	if len(m) == 0 {
		delete(g.edges, a)
	}
	g.size--
	return true
}

func (g *Graph[T]) Vertices() []T {
	result := []T{}
	for v := range g.vertices {
		result = append(result, v)
	}
	return result
}

func (g *Graph[T]) Edges() [][2]T {
	result := [][2]T{}
	for k, v := range g.edges {
		for k1 := range v {
			result = append(result, [2]T{k, k1})
		}
	}
	return result
}

func GrabFrom[T cmp.Ordered](g map[T]map[T]struct{}, start T) Graph[T] {
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

func AllNColoring[T cmp.Ordered](g Graph[T], n int) [][][]T {
	if n <= 0 || n > g.Order() {
		return nil
	}

	vs := g.Vertices()
	nv := g.Order()
	ne := g.Size()

	ones := []struct {
		First  int
		Second []int
	}{}
	vd := map[T]int{}
	colormap := map[int]struct {
		v T
		c int
	}{}
	k := 0

	for i, v := range vs {
		vd[v] = i
		for c := range n {
			ones = append(ones, struct {
				First  int
				Second []int
			}{k, []int{i}})
			colormap[k] = struct {
				v T
				c int
			}{v, c}
			k++
		}
	}

	kk := nv
	var v0, v1 int
	for _, e := range g.Edges() {
		v0 = n * vd[e[0]]
		v1 = n * vd[e[1]]
		for c := range n {
			ones[v0].Second = append(ones[v0].Second, kk+c)
			ones[v1].Second = append(ones[v1].Second, kk+c)
			v0++
			v1++
		}
		kk += n
	}

	if n > 2 {
		for i := 0; i < n*ne; n++ {
			ones = append(ones, struct {
				First  int
				Second []int
			}{k + i, []int{nv + i}})
		}
	}

	onesSecond := [][]int{}
	coloring := [][]T{}
	usedColors := map[int]struct{}{}
	for _, one := range ones {
		onesSecond = append(onesSecond, one.Second)
	}

	result := [][][]T{}

	d := dlx.NewDLX(onesSecond)
	for d.Search() {
		coloring = nil
		for range n {
			coloring = append(coloring, nil)
		}
		clear(usedColors)

		for _, x := range d.Solution {
			if s, ok := colormap[x]; ok {
				usedColors[s.c] = struct{}{}
				coloring[s.c] = append(coloring[s.c], s.v)
			}
		}

		if len(usedColors) == n {
			result = append(result, coloring)
		}
	}

	return result
}

func NColoring[T cmp.Ordered](g Graph[T], n int) [][]T {
	if n <= 0 || n > g.Order() {
		return nil
	}

	vs := g.Vertices()
	nv := g.Order()
	ne := g.Size()

	ones := []struct {
		First  int
		Second []int
	}{}
	vd := map[T]int{}
	colormap := map[int]struct {
		v T
		c int
	}{}
	k := 0

	for i, v := range vs {
		vd[v] = i
		for c := range n {
			ones = append(ones, struct {
				First  int
				Second []int
			}{k, []int{i}})
			colormap[k] = struct {
				v T
				c int
			}{v, c}
			k++
		}
	}

	kk := nv
	var v0, v1 int
	for _, e := range g.Edges() {
		v0 = n * vd[e[0]]
		v1 = n * vd[e[1]]
		for c := range n {
			ones[v0].Second = append(ones[v0].Second, kk+c)
			ones[v1].Second = append(ones[v1].Second, kk+c)
			v0++
			v1++
		}
		kk += n
	}

	if n > 2 {
		for i := range n * ne {
			ones = append(ones, struct {
				First  int
				Second []int
			}{k + i, []int{nv + i}})
		}
	}

	onesSecond := [][]int{}
	coloring := [][]T{}
	usedColors := map[int]struct{}{}
	for _, one := range ones {
		onesSecond = append(onesSecond, one.Second)
	}

	d := dlx.NewDLX(onesSecond)
	if d.Search() {
		coloring = nil
		for range n {
			coloring = append(coloring, nil)
		}
		clear(usedColors)

		for _, x := range d.Solution {
			if s, ok := colormap[x]; ok {
				usedColors[s.c] = struct{}{}
				coloring[s.c] = append(coloring[s.c], s.v)
			}
		}

		return coloring
	}

	return nil
}

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
		for j := 2; ; j++ {
			coloring := NColoring(g, j)
			if len(coloring) == 0 {
				continue
			}

			for c, ids := range coloring {
				for _, id := range ids {
					colored[id] = c
				}
			}

			break
		}

	}
	result := []int{}
	for i := range nds.maxID {
		result = append(result, colored[i])
	}

	return result
}
