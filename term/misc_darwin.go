//go:build darwin

package term

import (
	"os"

	"golang.org/x/sys/unix"
)

func getTermios() (*unix.Termios, error) {
	if f, err := os.OpenFile("/dev/tty", unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NDELAY|unix.O_RDWR, 0o666); err == nil {
		defer f.Close()
		if t, err := unix.IoctlGetTermios(int(f.Fd()), unix.TIOCGETA); err == nil {
			return t, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func setTermios(t *unix.Termios) error {
	if f, err := os.OpenFile("/dev/tty", unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NDELAY|unix.O_RDWR, 0o666); err == nil {
		defer f.Close()
		return unix.IoctlSetTermios(int(f.Fd()), unix.TIOCSETA, t)
	} else {
		return err
	}
}
