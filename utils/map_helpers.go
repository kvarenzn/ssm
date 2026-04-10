// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package utils

import (
	"cmp"
	"slices"
)

func SortedKeysOf[K cmp.Ordered, V any](m map[K]V) []K {
	var result []K
	for k := range m {
		result = append(result, k)
	}
	slices.Sort(result)
	return result
}
