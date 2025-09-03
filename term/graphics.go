package term

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"

	"golang.org/x/image/draw"
)

type GraphicsMethod uint8

const (
	HALF_BLOCK GraphicsMethod = iota
	OVERSTRIKED_DOTS
	SIXEL_PROTOCOL
	ITERM2_GRAPHICS_PROTOCOL
	KITTY_GRAPHICS_PROTOCOL
)

func GetGraphicsMethod() GraphicsMethod {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	if term == "xterm-kitty" || term == "xterm-ghostty" || os.Getenv("KONSOLE_VERSION") != "" || termProgram == "WezTerm" {
		return KITTY_GRAPHICS_PROTOCOL
	} else if termProgram == "iTerm.app" || term == "mlterm" || termProgram == "mintty" {
		return ITERM2_GRAPHICS_PROTOCOL
	} else if term == "foot" || os.Getenv("WT_SESSION") != "" {
		return SIXEL_PROTOCOL
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

func repeats(b byte, n int) []byte {
	return bytes.Repeat([]byte{b}, n)
}

func DisplayImageUsingHalfBlock(i image.Image, upper bool, padLeft int) {
	buffer := bufio.NewWriter(os.Stdout)

	bounds := i.Bounds()
	mx := bounds.Max
	mn := bounds.Min
	size := mx.Sub(mn)
	MoveHome()
	for y := mn.Y; y < mx.Y; y += 2 {
		buffer.Write(repeats(' ', padLeft))
		for x := mn.X; x < mx.X; x++ {
			r1, g1, b1, _ := i.At(x, y).RGBA()
			r2, g2, b2, _ := i.At(x, y+1).RGBA()
			if upper {
				fmt.Fprintf(buffer, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm"+upperHalfBlock, r1>>8, g1>>8, b1>>8, r2>>8, g2>>8, b2>>8)
			} else {
				fmt.Fprintf(buffer, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm"+lowerHalfBlock, r2>>8, g2>>8, b2>>8, r1>>8, g1>>8, b1>>8)
			}
		}
		buffer.WriteString("\x1b[0m")
		FClearToRight(buffer)
		buffer.WriteByte('\n')
	}

	if size.Y%2 != 0 {
		buffer.Write(repeats(' ', padLeft))
		y := mx.Y
		for x := mn.X; x < mx.X; x++ {
			r1, g1, b1, _ := i.At(x, y).RGBA()
			if upper {
				fmt.Fprintf(buffer, "\x1b[38;2;%d;%d;%dm"+upperHalfBlock, r1>>8, g1>>8, b1>>8)
			} else {
				fmt.Fprintf(buffer, "\x1b[38;2;%d;%d;%dm\x1b[7m"+lowerHalfBlock, r1>>8, g1>>8, b1>>8)
			}
		}
		buffer.WriteString("\x1b[0m")
		FClearToRight(buffer)
		buffer.WriteByte('\n')
	}
}

func b64encode(data []byte) []byte {
	total := base64.StdEncoding.EncodedLen(len(data))
	encoded := make([]byte, total)
	base64.StdEncoding.Encode(encoded, data)
	return encoded
}

func KittyImageProtocol(buffer *bufio.Writer, i image.Image, hasAlpha bool, offsetX, offsetY int) {
	const CHUNK_SIZE = 4096
	data := ReadImageBytes(i, hasAlpha)
	payload := b64encode(data)
	buffer.WriteString("\x1b_Ga=T,")
	if !hasAlpha {
		buffer.WriteString("f=24,")
	}
	bound := i.Bounds()
	width := bound.Dx()
	height := bound.Dy()
	fmt.Fprintf(buffer, "s=%d,v=%d", width, height)
	if offsetX != 0 {
		fmt.Fprintf(buffer, ",X=%d", offsetX)
	}
	if offsetY != 0 {
		fmt.Fprintf(buffer, ",Y=%d", offsetY)
	}
	if len(payload) <= CHUNK_SIZE {
		buffer.WriteByte(';')
		buffer.Write(payload)
		buffer.WriteByte('\x1b')
		buffer.WriteByte('\\')
		return
	}

	buffer.WriteByte(',')
	for len(payload) > CHUNK_SIZE {
		buffer.WriteString("m=1;")
		buffer.Write(payload[:CHUNK_SIZE])
		payload = payload[CHUNK_SIZE:]
		buffer.WriteString("\x1b\\\x1b_G")
	}
	buffer.WriteString("m=0;")
	buffer.Write(payload)
	buffer.WriteString("\x1b\\")
}

func DisplayImageUsingKittyProtocol(i image.Image, size *TermSize, jacketHeight int) {
	padLeftPixels := (size.Xpixel - i.Bounds().Dx()) / 2
	buffer := bufio.NewWriter(os.Stdout)
	FMoveHome(buffer)
	buffer.Write(repeats(' ', padLeftPixels/size.CellWidth))
	FClearToRight(buffer)
	KittyImageProtocol(buffer, i, true, padLeftPixels%size.CellWidth, 0)
	buffer.Flush()
}

func DisplayImageUsingITerm2Protocol(i image.Image, size *TermSize, jacketHeight int) {
	buffer := bufio.NewWriter(os.Stdout)

	bounds := i.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	ih := jacketHeight * size.CellHeight
	iw := ih * dx / dy
	iww := iw
	if iw%size.CellWidth != 0 {
		iww += size.CellWidth - iw%size.CellWidth
	}
	iwc := iww / size.CellWidth

	FMoveHome(buffer)
	buffer.Write(bytes.Repeat([]byte{' '}, max((size.Col-iwc)/2, 0)))
	buffer.WriteString("\x1b]1337;File=inline=1")
	fmt.Fprintf(buffer, ";width=%d", iwc)
	fmt.Fprintf(buffer, ";height=%d", jacketHeight)
	buffer.WriteString(";preserveAspectRatio=0")
	buffer.WriteByte(':')
	out := image.NewNRGBA(image.Rect(0, 0, iww, ih))
	draw.BiLinear.Scale(out, image.Rect((iww-iw)/2, 0, iw, ih), i, bounds, draw.Src, nil)
	buf := bytes.NewBuffer(nil)
	png.Encode(buf, out)
	buffer.Write(b64encode(buf.Bytes()))
	buffer.WriteByte('\a')

	buffer.Flush()
}

const (
	wuBins        = 32
	paletteColors = 256
)

func DisplayImageUsingSixelProtocol(i image.Image, size *TermSize, jacketHeight int) {
	buffer := bufio.NewWriter(os.Stdout)

	bounds := i.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	ih := jacketHeight * size.CellHeight
	iw := ih * dx / dy
	iwc := iw / size.CellWidth

	FMoveHome(buffer)
	buffer.Write(repeats(' ', max((size.Col-iwc)/2, 0)))
	q := NewWuQuantizer(wuBins)
	q.buildHistogram(i)
	q.buildMoments()
	palette, boxes := q.Quantize(paletteColors)
	indexes := q.mapImageToPalette(i, boxes, palette)
	sixelOutput(buffer, i.Bounds(), palette, indexes)
	buffer.Flush()
}

var _DOTS = []int{1, 8, 2, 16, 4, 32, 64, 128}

func DisplayImageUsingOverstrikedDots(i image.Image, offsetX int, offsetY int, padLeft int) {
	buffer := bufio.NewWriter(os.Stdout)

	offsetX %= 2
	offsetY %= 4

	padding := repeats(' ', padLeft)
	bounds := i.Bounds()
	FMoveHome(buffer)
	for y := bounds.Min.Y - offsetY; y < bounds.Max.Y; y += 4 {
		buffer.Write(padding)
		for x := bounds.Min.X - offsetX; x < bounds.Max.X; x += 2 {
			buffer.WriteByte(' ')
			for dy := range 4 {
				for dx := range 2 {
					r, g, b, a := i.At(x+dx, y+dy).RGBA()
					if a == 0 {
						continue
					}

					buffer.WriteString("\x1b[D")                                 // <-
					fmt.Fprintf(buffer, "\x1b[38;2;%d;%d;%dm", r>>8, g>>8, b>>8) // set color
					buffer.WriteString("\x1b[?20h")
					buffer.WriteRune(rune(0x2800 + _DOTS[dy<<1|dx]))
				}
			}
		}
		buffer.WriteString("\x1b[0m")
		FClearToRight(buffer)
		buffer.WriteByte('\n')
	}
}
