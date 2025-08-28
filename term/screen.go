package term

import "fmt"

func ClearScreen() {
	fmt.Print("\033[2J")
}

func ClearCurrentLine() {
	fmt.Print("\033[2K")
}

func ClearToRight() {
	fmt.Print("\033[K")
}

func UseAlternateScreenBuffer() {
	fmt.Print("\033[?1049h")
}

func UseNormalScreenBuffer() {
	fmt.Print("\033[?1049l")
}
