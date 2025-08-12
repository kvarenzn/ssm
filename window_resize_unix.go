//go:build unix

package main

import (
	"os"
	"os/signal"
	"syscall"
)

func watchResize(sigCh chan<- os.Signal) {
	signal.Notify(sigCh, syscall.SIGWINCH)
}
