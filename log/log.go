package log

import (
	"fmt"
	"os"
	"runtime/debug"
)

func Fatal(args ...any) {
	fmt.Print("\033[1;41m FATAL \033[0m ")
	fmt.Println(args...)
	trace := debug.Stack()
	fmt.Println(string(trace))
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	fmt.Print("\033[1;41m FATAL \033[0m ")
	fmt.Printf(format, args...)
	fmt.Println()
	trace := debug.Stack()
	fmt.Println(string(trace))
	os.Exit(1)
}

func Info(args ...any) {
	fmt.Print("\033[1;46m INFO \033[0m ")
	fmt.Println(args...)
}

func Warnf(format string, args ...any) {
	fmt.Print("\033[1;45m WARN \033[0m ")
	fmt.Printf(format, args...)
	fmt.Println()
}

func Warn(args ...any) {
	fmt.Print("\033[1;45m WARN \033[0m ")
	fmt.Println(args...)
}

func Debug(args ...any) {
	fmt.Print("\033[1;44m DEBUG \033[0m ")
	fmt.Println(args...)
}
