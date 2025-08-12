//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/kvarenzn/ssm/log"
)

type (
	WORD    uint16
	DWORD   uint32
	LPDWORD *uint32
	SHORT   int16

	COORD struct {
		X SHORT
		Y SHORT
	}

	WindowBufferSizeRecord struct {
		Size COORD
	}

	InputRecord struct {
		EventType WORD
		Padding   [2]byte
		Event     [16]byte
	}
)

const (
	KEY_EVENT                = 0x0001
	MOUSE_EVENT              = 0x0002
	WINDOW_BUFFER_SIZE_EVENT = 0x0004
)

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procPeekConsoleInputW  = kernel32.NewProc("PeekConsoleInputW")
	procReadConsoleInputW  = kernel32.NewProc("ReadConsoleInputW")
	procWriteConsoleInputW = kernel32.NewProc("WriteConsoleInputW")
)

func peekConsoleInput(consoleInput syscall.Handle, buffer *InputRecord, length DWORD, numberOfEventsRead LPDWORD) error {
	r1, _, err := syscall.SyscallN(
		procPeekConsoleInputW.Addr(),
		uintptr(consoleInput),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(length),
		uintptr(unsafe.Pointer(numberOfEventsRead)),
	)
	if r1 == 0 {
		return err
	}

	return nil
}

func readConsoleInput(consoleInput syscall.Handle, buffer *InputRecord, length DWORD, numberOfEventsRead LPDWORD) error {
	r1, _, err := syscall.SyscallN(
		procReadConsoleInputW.Addr(),
		uintptr(consoleInput),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(length),
		uintptr(unsafe.Pointer(numberOfEventsRead)),
	)
	if r1 == 0 {
		return err
	}

	return nil
}

func writeConsoleInput(consoleInput syscall.Handle, buffer *InputRecord, length DWORD, numberOfEventsWritten LPDWORD) error {
	r1, _, err := syscall.SyscallN(
		procWriteConsoleInputW.Addr(),
		uintptr(consoleInput),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(length),
		uintptr(unsafe.Pointer(numberOfEventsWritten)),
	)
	if r1 == 0 {
		return err
	}

	return nil
}

func watchResize(sigCh chan<- os.Signal) {
	h := syscall.Handle(os.Stdin.Fd())

	var inputRec [1]InputRecord
	var eventsRead uint32

	for {
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
