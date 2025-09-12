// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build unix

package term

import (
	"os"
	"os/signal"
	"syscall"
)

func StartWatchResize(sigCh chan<- os.Signal) {
	signal.Notify(sigCh, syscall.SIGWINCH)
}

func StopWatchResize(sigCh chan<- os.Signal) {
	signal.Stop(sigCh)
}
