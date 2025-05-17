package term

import (
	"os"

	"golang.org/x/sys/unix"
)

func GetTerminalSize() (*unix.Winsize, error) {
	if f, err := os.OpenFile("/dev/tty", unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NDELAY|unix.O_RDWR, 0o666); err == nil {
		defer f.Close()
		if sz, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ); err == nil {
			return sz, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func GetTermios() (*unix.Termios, error) {
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

func SetTermios(t *unix.Termios) error {
	if f, err := os.OpenFile("/dev/tty", unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NDELAY|unix.O_RDWR, 0o666); err == nil {
		defer f.Close()
		return unix.IoctlSetTermios(int(f.Fd()), unix.TCSETS, t)
	} else {
		return err
	}
}

func PrepareTerminal() error {
	t, err := GetTermios()
	if err != nil {
		return err
	}

	t.Iflag &= ^uint32(unix.IGNBRK | unix.BRKINT | unix.ISTRIP | unix.INLCR | unix.ICRNL | unix.IXON)
	t.Lflag &= ^uint32(unix.ECHO | unix.ECHONL | unix.ICANON | unix.IEXTEN)
	t.Cflag &= ^uint32(unix.CSIZE | unix.PARENB)
	t.Cflag |= unix.CS8

	return SetTermios(t)
}
