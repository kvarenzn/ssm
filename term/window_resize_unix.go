//go:build unix

package term

import (
	"os"
	"os/signal"
	"syscall"
)

func WatchResize(sigCh chan<- os.Signal) {
	signal.Notify(sigCh, syscall.SIGWINCH)
}
