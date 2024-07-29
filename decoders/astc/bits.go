package astc

import (
	"encoding/binary"
	"math"
	"math/bits"
)

func bitReverseU8(b byte, c int) byte {
	return bits.Reverse8(b) >> (8 - c)
}

func bitReverseU64(l uint64, c int) uint64 {
	return bits.Reverse64(l) >> (64 - c)
}

func u32(buf []byte) uint32 {
	switch len(buf) {
	case 0:
		return 0
	case 1:
		buf = append(buf, 0, 0, 0)
	case 2:
		buf = append(buf, 0, 0)
	case 3:
		buf = append(buf, 0)
	default:
		buf = buf[:4]
	}

	return binary.LittleEndian.Uint32(buf)
}

func u64(buf []byte) uint64 {
	switch len(buf) {
	case 0:
		return 0
	case 1:
		buf = append(buf, 0, 0, 0, 0, 0, 0, 0)
	case 2:
		buf = append(buf, 0, 0, 0, 0, 0, 0)
	case 3:
		buf = append(buf, 0, 0, 0, 0, 0)
	case 4:
		buf = append(buf, 0, 0, 0, 0)
	case 5:
		buf = append(buf, 0, 0, 0)
	case 6:
		buf = append(buf, 0, 0)
	case 7:
		buf = append(buf, 0)
	default:
		buf = buf[:8]
	}

	return binary.LittleEndian.Uint64(buf)
}

func getBits(buf []byte, bit, l int) int {
	v := buf[bit/8:]
	return int(u32(v)) >> (bit % 8) & (1<<l - 1)
}

func getBits64(buf []byte, bit, l int) uint64 {
	var mask uint64
	if l == 64 {
		mask = 0xffff_ffff_ffff_ffff
	} else {
		mask = (1 << uint64(l)) - 1
	}
	if l < 1 {
		return 0
	} else if bit >= 64 {
		return u64(buf[8:]) >> (bit - 64) & mask
	} else if bit <= 0 {
		return u64(buf) << -bit & mask
	} else if bit+l <= 64 {
		return u64(buf) >> bit & mask
	} else {
		return (u64(buf)>>bit | u64(buf[8:])<<(64-bit)) & mask
	}
}

func clampU8(n int) byte {
	if n > 0xff {
		return 0xff
	} else if n < 0 {
		return 0
	} else {
		return byte(n)
	}
}

func clampHdr(n int) uint16 {
	if n < 0 {
		return 0
	} else if n > 0xfff {
		return 0xfff
	} else {
		return uint16(n)
	}
}

func bitTransferSigned(a, b *int) {
	*b = *b>>1 | *a&0x80
	*a = *a >> 1 & 0x3f
	if *a&0x20 != 0 {
		*a -= 0x40
	}
}

func u16ToF32(h uint16) float32 {
	w := uint32(h) << 16
	sign := w & 0x8000_0000
	ww := w + w
	expOffset := uint32(0xe0) << 23
	expScale := float32(0x1.0p-112)
	normalizedValue := math.Float32frombits(ww>>4+expOffset) * expScale
	magicMask := uint32(126) << 23
	magicBias := float32(0.5)
	denormallizedValue := math.Float32frombits(ww>>17|magicMask) - magicBias
	denormallizedCutoff := uint32(1) << 27
	var result uint32
	if ww < denormallizedCutoff {
		result = math.Float32bits(denormallizedValue)
	} else {
		result = math.Float32bits(normalizedValue)
	}

	return math.Float32frombits(result | sign)
}

func f32ToU8(f float32) byte {
	c := math.Round(float64(f) * 0xff)
	if c < 0 {
		return 0
	} else if c > 0xff {
		return 0xff
	} else {
		return byte(c)
	}
}
