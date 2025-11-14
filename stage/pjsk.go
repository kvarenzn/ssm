// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package stage

func PJSKJudgeLinePos(width, height float64) (float64, float64, float64) {
	var laneWidth, laneYOffset float64
	isWideScreen := 9*width >= 16*height
	if isWideScreen {
		laneWidth = height * 275 / 216
		laneYOffset = height * 7 / 24
	} else {
		laneWidth = width * 275 / 384
		laneYOffset = width * 21 / 128
	}

	half := laneWidth / 2
	middle := width / 2
	return middle - half, middle + half, height/2 + laneYOffset
}
