// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package log

import (
	"fmt"
	"os"
	"runtime/debug"
)

var beforeDie func()

func SetBeforeDie(fn func()) {
	beforeDie = fn
}

func Fatal(args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	fmt.Print("\033[1;41m FATAL \033[0m ")
	fmt.Println(args...)
	trace := debug.Stack()
	fmt.Println(string(trace))
	os.Exit(1)
}

func Fatalln(args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	for _, a := range args {
		fmt.Print("\033[1;41m FATAL \033[0m ")
		fmt.Println(a)
	}
	trace := debug.Stack()
	fmt.Println(string(trace))
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	fmt.Print("\033[1;41m FATAL \033[0m ")
	fmt.Printf(format, args...)
	fmt.Println()
	trace := debug.Stack()
	fmt.Println(string(trace))
	os.Exit(1)
}

func Die(args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	fmt.Print("\033[1;41m FATAL \033[0m ")
	fmt.Println(args...)
	os.Exit(1)
}

func Dieln(args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	for _, a := range args {
		fmt.Print("\033[1;41m FATAL \033[0m ")
		fmt.Println(a)
	}
	os.Exit(1)
}

func Dief(format string, args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	fmt.Print("\033[1;41m FATAL \033[0m ")
	fmt.Printf(format, args...)
	fmt.Println()
	os.Exit(1)
}

func Info(args ...any) {
	fmt.Print("\033[1;46m INFO \033[0m ")
	fmt.Print(args...)
}

func Infof(format string, args ...any) {
	fmt.Print("\033[1;46m INFO \033[0m ")
	fmt.Printf(format, args...)
	fmt.Println()
}

func Infoln(args ...any) {
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

var showDebug = false

func ShowDebug(on bool) {
	showDebug = on
}

func Debugln(args ...any) {
	if !showDebug {
		return
	}

	fmt.Print("\033[1;44m DEBUG \033[0m ")
	fmt.Println(args...)
}

func Debugf(format string, args ...any) {
	if !showDebug {
		return
	}

	fmt.Print("\033[1;44m DEBUG \033[0m ")
	fmt.Printf(format, args...)
	fmt.Println()
}
