// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package log

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/kvarenzn/ssm/locale"
)

func sprintf(format string, args ...any) string {
	return locale.P.Sprintf(format, args...)
}

func sprints(args ...any) []string {
	s := []string{}
	for _, a := range args {
		s = append(s, locale.P.Sprintf(fmt.Sprint(a)))
	}
	return s
}

func sprint(args ...any) string {
	return strings.Join(sprints(args...), " ")
}

var beforeDie func()

func SetBeforeDie(fn func()) {
	beforeDie = fn
}

func Fatal(args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	fmt.Println(sprint("[FATAL]"), sprint(args...))
	trace := debug.Stack()
	fmt.Println(string(trace))
	os.Exit(1)
}

func Fatalln(args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	for _, a := range args {
		fmt.Println(sprint("[FATAL]"), sprint(a))
	}
	trace := debug.Stack()
	fmt.Println(string(trace))
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	fmt.Println(sprint("[FATAL]"), sprintf(format, args...))
	trace := debug.Stack()
	fmt.Println(string(trace))
	os.Exit(1)
}

func Die(args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	fmt.Println(sprint("[FATAL]"), sprint(args...))
	os.Exit(1)
}

func Dieln(args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	for _, a := range args {
		fmt.Println(sprint("[FATAL]"), sprint(a))
	}
	os.Exit(1)
}

func Dief(format string, args ...any) {
	if beforeDie != nil {
		beforeDie()
	}
	fmt.Println(sprint("[FATAL]"), sprintf(format, args...))
	os.Exit(1)
}

func Info(args ...any) {
	fmt.Print(sprint("[INFO]"), " ", sprint(args...))
}

func Infof(format string, args ...any) {
	fmt.Println(sprint("[INFO]"), sprintf(format, args...))
}

func Infoln(args ...any) {
	fmt.Println(sprint("[INFO]"), sprint(args...))
}

func Warnf(format string, args ...any) {
	fmt.Println(sprint("[WARN]"), sprintf(format, args...))
}

func Warn(args ...any) {
	fmt.Println(sprint("[WARN]"), sprint(args...))
}

var showDebug = false

func ShowDebug(on bool) {
	showDebug = on
}

func Debugln(args ...any) {
	if !showDebug {
		return
	}

	fmt.Println(sprint("[DEBUG]"), sprint(args...))
}

func Debugf(format string, args ...any) {
	if !showDebug {
		return
	}

	fmt.Println(sprint("[DEBUG]"), sprintf(format, args...))
}
