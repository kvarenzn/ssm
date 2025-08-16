package term

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
)

type GraphicsMethod uint8

const (
	HALF_BLOCK GraphicsMethod = iota
	OVERSTRIKED_DOTS
	KITTY_GRAPHICS_PROTOCOL
)

func GetGraphicsMethod() GraphicsMethod {
	emulator := os.Getenv("TERM")
	if emulator == "xterm-kitty" || emulator == "xterm-ghostty" || os.Getenv("WEZTERM_EXECUTABLE") != "" || os.Getenv("KONSOLE_VERSION") != "" {
		return KITTY_GRAPHICS_PROTOCOL
	} else if os.Getenv("TERM_PROGRAM") == "mintty" {
		return OVERSTRIKED_DOTS
	} else {
		return HALF_BLOCK
	}
}

func DecodeImage(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)
	if bytes.Equal(data[:4], []byte("\x89PNG")) {
		return png.Decode(reader)
	}

	return jpeg.Decode(reader)
}

func ReadImageBytes(img image.Image, hasAlphaChannel bool) []byte {
	bound := img.Bounds()
	width, height := bound.Dx(), bound.Dy()
	channels := 3
	if hasAlphaChannel {
		channels = 4
	}
	totalSize := width * height * channels
	buf := make([]byte, totalSize)
	off := 0
	for y := bound.Min.Y; y < bound.Max.Y; y++ {
		for x := bound.Min.X; x < bound.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			buf[off] = byte(r >> 8)
			off++
			buf[off] = byte(g >> 8)
			off++
			buf[off] = byte(b >> 8)
			off++
			if hasAlphaChannel {
				buf[off] = byte(a >> 8)
				off++
			}
		}
	}
	return buf
}

const (
	upperHalfBlock = "▀"
	lowerHalfBlock = "▄"
)

func DisplayImageUsingHalfBlock(i image.Image, upper bool, padLeft int) {
	bounds := i.Bounds()
	mx := bounds.Max
	mn := bounds.Min
	size := mx.Sub(mn)
	MoveHome()
	for y := mn.Y; y < mx.Y; y += 2 {
		print(strings.Repeat(" ", padLeft))
		for x := mn.X; x < mx.X; x++ {
			r1, g1, b1, _ := i.At(x, y).RGBA()
			r2, g2, b2, _ := i.At(x, y+1).RGBA()
			if upper {
				fmt.Printf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm"+upperHalfBlock, r1>>8, g1>>8, b1>>8, r2>>8, g2>>8, b2>>8)
			} else {
				fmt.Printf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm"+lowerHalfBlock, r2>>8, g2>>8, b2>>8, r1>>8, g1>>8, b1>>8)
			}
		}
		print("\x1b[0m")
		ClearToRight()
		println()
	}

	if size.Y%2 != 0 {
		print(strings.Repeat(" ", padLeft))
		y := mx.Y
		for x := mn.X; x < mx.X; x++ {
			r1, g1, b1, _ := i.At(x, y).RGBA()
			if upper {
				fmt.Printf("\x1b[38;2;%d;%d;%dm"+upperHalfBlock, r1>>8, g1>>8, b1>>8)
			} else {
				fmt.Printf("\x1b[38;2;%d;%d;%dm\x1b[7m"+lowerHalfBlock, r1>>8, g1>>8, b1>>8)
			}
		}
		print("\x1b[0m")
		ClearToRight()
		println()
	}
}

func DisplayImageUsingKittyProtocol(i image.Image, hasAlpha bool, offsetX, offsetY int) {
	const CHUNK_SIZE = 4096
	data := ReadImageBytes(i, hasAlpha)
	payload := base64.StdEncoding.EncodeToString(data)
	print("\x1b_Ga=T,")
	if !hasAlpha {
		print("f=24,")
	}
	bound := i.Bounds()
	width := bound.Dx()
	height := bound.Dy()
	fmt.Printf("s=%d,v=%d", width, height)
	if offsetX != 0 {
		fmt.Printf(",X=%d", offsetX)
	}
	if offsetY != 0 {
		fmt.Printf(",Y=%d", offsetY)
	}
	if len(payload) <= CHUNK_SIZE {
		print(";")
		print(payload)
		print("\x1b\\")
		return
	}

	print(",")
	for len(payload) > CHUNK_SIZE {
		print("m=1;")
		print(payload[:CHUNK_SIZE])
		payload = payload[CHUNK_SIZE:]
		print("\x1b\\\x1b_G")
	}
	print("m=0;")
	print(payload)
	print("\x1b\\")
}

func DisplayImageUsingOverstrikedDots(i image.Image, offsetX int, offsetY int, padLeft int) {
	const DOTS = "⠁⠈⠂⠐⠄⠠⡀⢀"
	offsetX %= 2
	offsetY %= 4

	padding := strings.Repeat(" ", padLeft)
	bounds := i.Bounds()
	for y := bounds.Min.Y - offsetY; y < bounds.Max.Y; y += 4 {
		print(padding)
		for x := bounds.Min.X - offsetX; x < bounds.Max.X; x += 2 {
			print(' ')
			for dy := range 4 {
				for dx := range 2 {
					r, g, b, a := i.At(x+dx, y+dy).RGBA()
					if a == 0 {
						continue
					}

					print("\x1b[D")                                     // <-
					fmt.Printf("\x1b[38;2;%d;%d;%dm", r>>8, g>>8, b>>8) // set color
					print("\x1b[?20h")
					print(DOTS[dy*2+dx])
				}
			}
		}
		print("\x1b[0m")
		ClearToRight()
		println()
	}
}
