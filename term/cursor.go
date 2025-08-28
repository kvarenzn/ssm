package term

import "fmt"

func HideCursor() {
	fmt.Print("\033[?25l")
}

func ShowCursor() {
	fmt.Print("\033[?25h")
}

func MoveRight(count int) {
	fmt.Printf("\033[%dC", count)
}

func MoveDown(count int) {
	fmt.Printf("\033[%dB", count)
}

func MoveDownAndReset(count int) {
	fmt.Printf("\033[%dE", count)
}

func MoveUp(count int) {
	fmt.Printf("\033[%dA", count)
}

func MoveUpAndReset(count int) {
	fmt.Printf("\033[%dF", count)
}

func MoveTo(row, column int) {
	fmt.Printf("\033[%d;%dH", row+1, column+1)
}

func MoveToColumn(column int) {
	fmt.Printf("\033[%dG", column+1)
}

func MoveHome() {
	MoveToColumn(0)
}

func ResetCursor() {
	fmt.Print("\033[H")
}
