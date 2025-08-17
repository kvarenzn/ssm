//go:build windows

package term

import (
	"os"
	"syscall"

	"github.com/kvarenzn/ssm/log"
)

var listenerRunning = map[chan<- os.Signal]*bool{}

func watchResizeInner(sigCh chan<- os.Signal) {
	h := syscall.Handle(os.Stdin.Fd())
	running := new(bool)
	*running = true
	listenerRunning[sigCh] = running

	var inputRec [1]InputRecord
	var eventsRead uint32

	for *running {
		var count uint32
		err := peekConsoleInput(h, &inputRec[0], 1, &count)
		if err != nil {
			log.Die("PeekConsoleInput() error:", err)
			return
		}
		if count == 0 {
			continue
		}

		err = readConsoleInput(h, &inputRec[0], 1, &eventsRead)
		if err != nil {
			log.Die("ReadConsoleInput() error:", err)
			return
		}

		ev := inputRec[0]
		if ev.EventType == WINDOW_BUFFER_SIZE_EVENT {
			select {
			case sigCh <- syscall.Signal(114514): // send something
			default:
			}
		} else {
			_ = writeConsoleInput(h, &ev, 1, &eventsRead) // write event back
		}
	}
}

func StartWatchResize(sigCh chan<- os.Signal) {
	go watchResizeInner(sigCh)
}

func StopWatchResize(sigCh chan<- os.Signal) {
	running, ok := listenerRunning[sigCh]
	if !ok {
		return
	}

	*running = false
}
