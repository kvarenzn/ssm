// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package stage

// 不要问我为什么这里的数字都这么怪
// 这都是我拿不同尺寸的截图一张张测量完硬凑的

func BanGJudgeLinePos(width, height float64) (float64, float64, float64) {
	var laneWidth, laneYOffset float64
	isWideScreen := 9*width > 16*height
	limitedRatio := min(width/height, 2)
	if isWideScreen {
		laneWidth = height * (limitedRatio/4 + 2.0/3)
		laneYOffset = height * (115 - 9*limitedRatio) / 110
	} else {
		laneWidth = width * 9 / 13
		laneYOffset = width * 9 / 16
	}

	half := laneWidth / 2
	middle := width / 2
	return middle - half, middle + half, height/2 + laneYOffset*26/81
}
