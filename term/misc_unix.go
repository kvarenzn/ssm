//go:build unix

package term

import (
	"os"

	"golang.org/x/sys/unix"
)

func GetTerminalSize() (*TermSize, error) {
	if f, err := os.OpenFile("/dev/tty", unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NDELAY|unix.O_RDWR, 0o666); err == nil {
		defer f.Close()
		if sz, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ); err == nil {
			return &TermSize{
				Row: sz.Row,
				Col: sz.Col,
			}, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func getTermios() (*unix.Termios, error) {
	if f, err := os.OpenFile("/dev/tty", unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NDELAY|unix.O_RDWR, 0o666); err == nil {
		defer f.Close()
		if t, err := unix.IoctlGetTermios(int(f.Fd()), unix.TCGETS); err == nil {
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
		return unix.IoctlSetTermios(int(f.Fd()), unix.TCSETS, t)
	} else {
		return err
	}
}

var settings *unix.Termios

func PrepareTerminal() error {
	if t, err := getTermios(); err != nil {
		return err
	} else {
		settings = t
	}

	if t, err := getTermios(); err != nil {
		return err
	} else {
		t.Iflag &= ^uint32(unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON)
		t.Lflag &= ^uint32(unix.ECHO | unix.ECHONL | unix.ICANON | unix.IEXTEN)
		t.Cflag &= ^uint32(unix.CSIZE | unix.PARENB)
		t.Cflag |= unix.CS8

		if err := setTermios(t); err != nil {
			return err
		}
	}

	HideCursor()

	return nil
}

func RestoreTerminal() error {
	if err := setTermios(settings); err != nil {
		return err
	}

	ShowCursor()

	return nil
}
