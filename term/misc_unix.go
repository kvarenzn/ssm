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
				Col:        int(sz.Col),
				Row:        int(sz.Row),
				Xpixel:     int(sz.Xpixel),
				Ypixel:     int(sz.Ypixel),
				CellWidth:  int(sz.Xpixel) / int(sz.Col),
				CellHeight: int(sz.Ypixel) / int(sz.Row),
			}, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

var settings *unix.Termios

func Hello() error {
	return nil
}

func PrepareTerminal() error {
	if t, err := getTermios(); err != nil {
		return err
	} else {
		settings = t
	}

	UseAlternateScreenBuffer()
	HideCursor()

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

	return nil
}

func RestoreTerminal() error {
	if err := setTermios(settings); err != nil {
		return err
	}

	ShowCursor()
	UseNormalScreenBuffer()

	return nil
}

func Bye() error {
	return nil
}
