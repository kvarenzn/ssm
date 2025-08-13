//go:build windows

package term

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procPeekConsoleInputW   = kernel32.NewProc("PeekConsoleInputW")
	procReadConsoleInputW   = kernel32.NewProc("ReadConsoleInputW")
	procWriteConsoleInputW  = kernel32.NewProc("WriteConsoleInputW")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
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

func peekConsoleInput(consoleInput syscall.Handle, buffer *InputRecord, length DWORD, numberOfEventsRead LPDWORD) error {
	r1, _, err := procPeekConsoleInputW.Call(
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
	r1, _, err := procReadConsoleInputW.Call(
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
	r1, _, err := procWriteConsoleInputW.Call(
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

const (
	WAIT_ABANDONED = 0x00000080
	WAIT_OBJECT_0  = 0x00000000
	WAIT_TIMEOUT   = 0x00000102
	WAIT_FAILED    = 0xffffffff
)

func waitForHandle(handle syscall.Handle, timeout time.Duration) (bool, error) {
	ms := uint32(timeout.Milliseconds())
	ret, _, err := procWaitForSingleObject.Call(uintptr(handle), uintptr(ms))
	if ret == WAIT_OBJECT_0 {
		return true, nil
	}

	if ret == WAIT_TIMEOUT {
		return false, nil
	}

	return false, err
}
