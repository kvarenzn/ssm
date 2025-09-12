// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package term

import (
	"fmt"
	"io"
)

func ClearScreen() {
	fmt.Print("\033[2J")
}

func ClearCurrentLine() {
	fmt.Print("\033[2K")
}

func ClearToRight() {
	fmt.Print("\033[K")
}

func FClearToRight(w io.Writer) {
	fmt.Fprint(w, "\033[K")
}

func UseAlternateScreenBuffer() {
	fmt.Print("\033[?1049h")
}

func UseNormalScreenBuffer() {
	fmt.Print("\033[?1049l")
}
