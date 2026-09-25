// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package log

import (
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"

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

type FatalErr struct{}

var (
	beforeDieMu sync.Mutex
	beforeDie   func()
)

func SetBeforeDie(fn func()) {
	beforeDieMu.Lock()
	beforeDie = fn
	beforeDieMu.Unlock()
}

func exit() {
	beforeDieMu.Lock()
	fn := beforeDie
	beforeDieMu.Unlock()

	if fn != nil {
		fn()
	}
	panic(FatalErr{})
}

func Fatal(args ...any) {
	fmt.Println(sprint("[FATAL]"), sprint(args...))
	trace := debug.Stack()
	fmt.Println(string(trace))
	exit()
}

func Fatalln(args ...any) {
	for _, a := range args {
		fmt.Println(sprint("[FATAL]"), sprint(a))
	}
	trace := debug.Stack()
	fmt.Println(string(trace))
	exit()
}

func Fatalf(format string, args ...any) {
	fmt.Println(sprint("[FATAL]"), sprintf(format, args...))
	trace := debug.Stack()
	fmt.Println(string(trace))
	exit()
}

func Die(args ...any) {
	fmt.Println(sprint("[FATAL]"), sprint(args...))
	exit()
}

func Dieln(args ...any) {
	for _, a := range args {
		fmt.Println(sprint("[FATAL]"), sprint(a))
	}
	exit()
}

func Dief(format string, args ...any) {
	fmt.Println(sprint("[FATAL]"), sprintf(format, args...))
	exit()
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

var showDebug atomic.Bool

func ShowDebug(on bool) {
	showDebug.Store(on)
}

func Debugln(args ...any) {
	if !showDebug.Load() {
		return
	}

	fmt.Println(sprint("[DEBUG]"), sprint(args...))
}

func Debugf(format string, args ...any) {
	if !showDebug.Load() {
		return
	}

	fmt.Println(sprint("[DEBUG]"), sprintf(format, args...))
}
