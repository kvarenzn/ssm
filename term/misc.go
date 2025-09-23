// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package term

import "fmt"

type TermSize struct {
	Col        int // terminal width (in cell)
	Row        int // terminal height (in cell)
	Xpixel     int // terminal width (in pixel)
	Ypixel     int // terminal height (in pixel)
	CellWidth  int // font width (in pixel)
	CellHeight int // font height (in pixel)
}

func SetWindowTitle(title string) {
	fmt.Printf("\x1b]2;%s\x07", title)
}
