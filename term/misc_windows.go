//go:build windows

package term

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

func GetTerminalSize() (*TermSize, error) {
	info := windows.ConsoleScreenBufferInfo{}
	if err := windows.GetConsoleScreenBufferInfo(windows.Stdout, &info); err != nil {
		return nil, err
	}

	row := int(info.Window.Bottom - info.Window.Top + 1)
	col := int(info.Window.Right - info.Window.Left + 1)

	var x, y int
	fontInfo := ConsoleFontInfo{}
	if err := getCurrentConsoleFont(syscall.Handle(os.Stdout.Fd()), false, &fontInfo); err == nil {
		x = int(fontInfo.FontSize.X)
		y = int(fontInfo.FontSize.Y)
	}

	return &TermSize{
		Row:    row,
		Col:    col,
		Xpixel: col * x,
		Ypixel: row * y,
	}, nil
}

func getConsoleModePair() (uint32, uint32, error) {
	var in, out uint32
	if err := windows.GetConsoleMode(windows.Stdin, &in); err != nil {
		return 0, 0, err
	}

	if err := windows.GetConsoleMode(windows.Stdout, &out); err != nil {
		return 0, 0, err
	}

	return in, out, nil
}

func setConsoleModePair(in, out uint32) error {
	if err := windows.SetConsoleMode(windows.Stdin, in); err != nil {
		return err
	}

	if err := windows.SetConsoleMode(windows.Stdout, out); err != nil {
		return err
	}

	return nil
}

var (
	inputMode  uint32
	outputMode uint32
)

func enableVirtualTerminalSupport() error {
	in, out, err := getConsoleModePair()
	if err != nil {
		return err
	}

	inputMode, outputMode = in, out

	in &^= windows.ENABLE_ECHO_INPUT
	in &^= windows.ENABLE_LINE_INPUT
	in &^= windows.ENABLE_MOUSE_INPUT
	in |= windows.ENABLE_VIRTUAL_TERMINAL_INPUT

	out |= windows.ENABLE_PROCESSED_OUTPUT
	out |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING

	return setConsoleModePair(in, out)
}

func PrepareTerminal() error {
	if err := enableVirtualTerminalSupport(); err != nil {
		return err
	}

	HideCursor()

	return nil
}

func RestoreTerminal() error {
	if err := setConsoleModePair(inputMode, outputMode); err != nil {
		return err
	}

	ShowCursor()

	return nil
}
