//go:build windows

package term

import (
	"os"
	"syscall"
	"time"
)

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
