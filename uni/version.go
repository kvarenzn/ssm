// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

type Version []int

func (v Version) GreaterEqual(target ...int) bool {
	if len(target) == 1 {
		target = append(target, 0, 0, 0)
	} else if len(target) == 2 {
		target = append(target, 0, 0)
	} else if len(target) == 3 {
		target = append(target, 0)
	}

	return v[0] > target[0] || v[0] == target[0] && v[1] > target[1] || v[0] == target[0] && v[1] == target[1] && v[2] > target[2] || v[0] == target[0] && v[1] == target[1] && v[2] == target[2] && v[3] >= target[3]
}

func (v Version) GreaterThan(target ...int) bool {
	if len(target) == 1 {
		target = append(target, 0, 0, 0)
	} else if len(target) == 2 {
		target = append(target, 0, 0)
	} else if len(target) == 3 {
		target = append(target, 0)
	}

	return v[0] > target[0] || v[0] == target[0] && v[1] > target[1] || v[0] == target[0] && v[1] == target[1] && v[2] > target[2] || v[0] == target[0] && v[1] == target[1] && v[2] == target[2] && v[3] > target[3]
}

func (v Version) LessEqual(target ...int) bool {
	if len(target) == 1 {
		target = append(target, 0, 0, 0)
	} else if len(target) == 2 {
		target = append(target, 0, 0)
	} else if len(target) == 3 {
		target = append(target, 0)
	}

	return v[0] < target[0] || v[0] == target[0] && v[1] < target[1] || v[0] == target[0] && v[1] == target[1] && v[2] < target[2] || v[0] == target[0] && v[1] == target[1] && v[2] == target[2] && v[3] <= target[3]
}

func (v Version) LessThan(target ...int) bool {
	if len(target) == 1 {
		target = append(target, 0, 0, 0)
	} else if len(target) == 2 {
		target = append(target, 0, 0)
	} else if len(target) == 3 {
		target = append(target, 0)
	}

	return v[0] < target[0] || v[0] == target[0] && v[1] < target[1] || v[0] == target[0] && v[1] == target[1] && v[2] < target[2] || v[0] == target[0] && v[1] == target[1] && v[2] == target[2] && v[3] < target[3]
}
