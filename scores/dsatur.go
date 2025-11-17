// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

// DSATUR graph-coloring algorithm in go

package scores

func SaturationLargestFirstGreedyColoring[T comparable](g *graph[T]) map[T]int {
	if len(g.vertices) == 0 {
		return nil
	}

	N := len(g.vertices)

	forbiddenColors := map[T]map[int]struct{}{}

	result := map[T]int{}
	nodesToBeColored := map[T]struct{}{}

	// fill `forbiddenColors`, `nodesToBeColored` & find node with max degree
	maxDegree := 0
	var selected T
	for k, v := range g.edges {
		forbiddenColors[k] = map[int]struct{}{}
		nodesToBeColored[k] = struct{}{}

		degree := len(v)
		if degree > maxDegree {
			maxDegree = degree
			selected = k
		}
	}

	// color selected node using color 0 & update forbidden-colors for neighbors
	delete(nodesToBeColored, selected)
	result[selected] = 0
	for neighbor := range g.edges[selected] {
		forbiddenColors[neighbor][0] = struct{}{}
	}

	for len(result) < N {
		// find next node to be colored
		maxDegree = 0
		maxSaturation := 0

		for n := range nodesToBeColored {
			saturation := len(forbiddenColors[n])
			degree := len(g.edges[n])
			if saturation > maxSaturation || saturation == maxSaturation && degree > maxDegree {
				maxSaturation = saturation
				maxDegree = degree
				selected = n
			}
		}

		// color that node with a usable color
		for c := 0; ; c++ {
			if _, ok := forbiddenColors[selected][c]; !ok {
				// use the color to color that node

				delete(nodesToBeColored, selected)
				result[selected] = c

				for neighbor := range g.edges[selected] {
					forbiddenColors[neighbor][c] = struct{}{}
				}

				break
			}
		}
	}

	return result
}
