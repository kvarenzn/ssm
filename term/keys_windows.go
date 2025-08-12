//go:build windows

package term

import (
	"os"
	"syscall"
	"time"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
	waitObject0             = uintptr(0)
	waitTimeout             = uintptr(0x00000102)
)

func waitForHandle(handle syscall.Handle, timeout time.Duration) (bool, error) {
	ms := uint32(timeout.Milliseconds())
	ret, _, err := procWaitForSingleObject.Call(uintptr(handle), uintptr(ms))
	if ret == waitObject0 {
		return true, nil
	}

	if ret == waitTimeout {
		return false, nil
	}

	return false, err
}

func readByteWithTimeout(reader *os.File, timeout time.Duration) error {
	handle := syscall.Handle(reader.Fd())

	ready, err := waitForHandle(handle, timeout)
	if err != nil {
		return err
	}

	if !ready {
		return os.ErrDeadlineExceeded
	}

	_, err = reader.Read(oneByte)
	return err
}
