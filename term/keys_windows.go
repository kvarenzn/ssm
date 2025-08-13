//go:build windows

package term

import (
	"os"
	"time"

	"golang.org/x/sys/windows"
)

func readByteWithTimeout(reader *os.File, timeout time.Duration) error {
	handle := windows.Handle(reader.Fd())

	ev, err := windows.WaitForSingleObject(handle, uint32(timeout.Milliseconds()))
	if err != nil {
		return err
	}

	if ev != WAIT_OBJECT_0 {
		return os.ErrDeadlineExceeded
	}

	_, err = reader.Read(oneByte)
	return err
}
