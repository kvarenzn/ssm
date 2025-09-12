// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build unix

package term

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func readByteWithTimeout(reader *os.File, timeout time.Duration) error {
	fd := int(reader.Fd())

	var readFds unix.FdSet
	readFds.Bits[fd/64] |= 1 << (uint(fd) % 64)

	tv := unix.NsecToTimeval(timeout.Nanoseconds())
	n, err := unix.Select(fd+1, &readFds, nil, nil, &tv)
	if err != nil {
		return err
	}

	if n == 0 {
		return os.ErrDeadlineExceeded
	}

	_, err = reader.Read(oneByte)
	return err
}
