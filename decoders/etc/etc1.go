package etc

import (
	"image"

	"github.com/kvarenzn/ssm/decoders"
)

func decode1Block(data []byte, out []byte) {
	code := []byte{data[3] >> 5, data[3] >> 2 & 7}
	table := etc1SubblockTable[data[3]&1]

	var c [2][3]byte

	if data[3]&2 != 0 {
		c[0][0] = data[0] & 0xf8
		c[0][1] = data[1] & 0xf8
		c[0][2] = data[2] & 0xf8
		c[1][0] = c[0][0] + data[0]<<3&0x18 - data[0]<<3&0x20
		c[1][1] = c[0][1] + data[1]<<3&0x18 - data[1]<<3&0x20
		c[1][2] = c[0][2] + data[2]<<3&0x18 - data[2]<<3&0x20
		c[0][0] |= c[0][0] >> 5
		c[0][1] |= c[0][1] >> 5
		c[0][2] |= c[0][2] >> 5
		c[1][0] |= c[1][0] >> 5
		c[1][1] |= c[1][1] >> 5
		c[1][2] |= c[1][2] >> 5
	} else {
		c[0][0] = data[0]&0xf0 | data[0]>>4
		c[1][0] = data[0]&0x0f | data[0]<<4
		c[0][1] = data[1]&0xf0 | data[1]>>4
		c[1][1] = data[1]&0x0f | data[1]<<4
		c[0][2] = data[2]&0xf0 | data[2]>>4
		c[1][2] = data[2]&0x0f | data[2]<<4
	}

	j := uint16(data[6])<<8 | uint16(data[7])
	k := uint16(data[4])<<8 | uint16(data[5])

	for i := 0; i < 16; i, j, k = i+1, j>>1, k>>1 {
		s := table[i]
		m := int16(etc1ModifierTable[code[s]][j&1])

		if k&1 != 0 {
			m = -m
		}
		pos := writeOrderTable[i]
		copy(out[pos*4:], applicateColor(c[s], m))
	}
}

func Decode1(data []byte, width, height int) (image.Image, error) {
	xBlocks := (width + 3) / 4
	yBlocks := (height + 3) / 4

	buffer := make([]byte, 16*4)

	result := make([]byte, width*height*4)
	ptr := 0
	for blockY := range yBlocks {
		for blockX := 0; blockX < xBlocks; blockX, ptr = blockX+1, ptr+8 {
			decode1Block(data[ptr:], buffer)
			decoders.CopyBlockBuffer(blockX, blockY, width, height, 4, 4, buffer, result)
		}
	}

	return &image.NRGBA{
		Pix:    result,
		Stride: width * 4,
		Rect:   image.Rect(0, 0, width, height),
	}, nil
}
