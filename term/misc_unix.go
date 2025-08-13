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
				Row:    int(sz.Row),
				Col:    int(sz.Col),
				Xpixel: int(sz.Xpixel),
				Ypixel: int(sz.Ypixel),
			}, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
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
		t.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
		t.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.IEXTEN
		t.Cflag &^= unix.CSIZE | unix.PARENB
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
