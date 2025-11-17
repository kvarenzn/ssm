// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package controllers

import (
	"cmp"
	"math"
)

func interp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func roundint(f float64) int {
	return int(math.Round(f))
}

func clamp[T cmp.Ordered](x, lo, hi T) T {
	if x < lo {
		return lo
	} else if x > hi {
		return hi
	} else {
		return x
	}
}

func crinterp(a, b, t, lo, hi float64) int {
	return clamp(roundint(interp(a, b, t)), int(lo), int(hi))
}
