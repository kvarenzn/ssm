package term

func ClearScreen() {
	print("\033[2J")
}

func ClearCurrentLine() {
	print("\033[2K")
}

func ClearToRight() {
	print("\033[K")
}

func UseAlternateScreenBuffer() {
	print("\033[?1049h")
}

func UseNormalScreenBuffer() {
	print("\033[?1049l")
}
