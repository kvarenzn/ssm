//go:build unix

package term

import (
	"os"
	"syscall"
	"time"
)

func readByteWithTimeout(reader *os.File, timeout time.Duration) error {
	fd := int(reader.Fd())

	var readFds syscall.FdSet
	readFds.Bits[fd/64] |= 1 << (uint(fd) % 64)

	tv := syscall.NsecToTimeval(timeout.Nanoseconds())
	n, err := syscall.Select(fd+1, &readFds, nil, nil, &tv)
	if err != nil {
		return err
	}

	if n == 0 {
		return os.ErrDeadlineExceeded
	}

	_, err = reader.Read(oneByte)
	return err
}
