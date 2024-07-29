package etc

import (
	"encoding/binary"
	"image"

	"github.com/kvarenzn/ssm/decoders"
)

func decode2Block(data []byte, out []byte) {
	j := uint16(data[6])<<8 | uint16(data[7])
	k := uint32(data[4])<<8 | uint32(data[5])
	var c [3][3]byte

	if data[3]&2 != 0 {
		r := data[0] & 0xf8
		dr := int16(data[0])<<3&0x18 - int16(data[0])<<3&0x20
		g := data[1] & 0xf8
		dg := int16(data[1])<<3&0x18 - int16(data[1])<<3&0x20
		b := data[2] & 0xf8
		db := int16(data[2])<<3&0x18 - int16(data[2])<<3&0x20

		if int16(r)+dr < 0 || int16(r)+dr > 0xff {
			c[0][0] = data[0]<<3&0xc0 | data[0]<<4&0x30 | data[0]>>1&0xc | data[0]&3
			c[0][1] = data[1]&0xf0 | data[1]>>4
			c[0][2] = data[1]&0x0f | data[1]<<4
			c[1][0] = data[2]&0xf0 | data[2]>>4
			c[1][1] = data[2]&0x0f | data[2]<<4
			c[1][2] = data[3]&0xf0 | data[3]>>4

			d := int16(etc2DistanceTable[data[3]>>1&6|data[3]&1])
			colors := [4][]byte{
				applicateColorRaw(c[0]),
				applicateColor(c[1], d),
				applicateColorRaw(c[1]),
				applicateColor(c[1], -d),
			}

			k <<= 1
			for i := 0; i < 16; i, j, k = i+1, j>>1, k>>1 {
				copy(out[writeOrderTable[i]*4:], colors[k&2|uint32(j&1)])
			}
		} else if int16(g)+dg < 0 || int16(g)+dg > 0xff {
			c[0][0] = data[0]<<1&0xf0 | data[0]>>3&0xf
			c[0][1] = data[0]<<5&0xe0 | data[1]&0x10
			c[0][1] |= c[0][1] >> 4
			c[0][2] = data[1]&8 | data[1]<<1&6 | data[2]>>7
			c[0][2] |= c[0][2] << 4
			c[1][0] = data[2]<<1&0xf0 | data[2]>>3&0xf
			c[1][1] = data[2]<<5&0xe0 | data[3]>>3&0x10
			c[1][1] |= c[1][1] >> 4
			c[1][2] = data[3]<<1&0xf0 | data[3]>>3&0xf
			d := data[3]&4 | data[3]<<1&2
			if c[0][0] > c[1][0] || c[0][0] == c[1][0] && (c[0][1] > c[1][1] || c[0][1] == c[1][1] && c[0][2] >= c[1][2]) {
				d++
			}

			dd := int16(etc2DistanceTable[d])
			colors := [4][]byte{
				applicateColor(c[0], dd),
				applicateColor(c[0], -dd),
				applicateColor(c[1], dd),
				applicateColor(c[1], -dd),
			}

			k <<= 1
			for i := 0; i < 16; i, j, k = i+1, j>>1, k>>1 {
				copy(out[writeOrderTable[i]*4:], colors[k&2|uint32(j&1)])
			}
		} else if int16(b)+db < 0 || int16(b)+db > 0xff {
			c[0][0] = data[0]<<1&0xfc | data[0]>>5&3
			c[0][1] = data[0]<<7&0x80 | data[1]&0x7e | data[0]&1
			c[0][2] = data[1]<<7&0x80 | data[2]<<2&0x60 | data[2]<<3&0x18 | data[3]>>5&4
			c[0][2] |= c[0][2] >> 6
			c[1][0] = data[3]<<1&0xf8 | data[3]<<2&4 | data[3]>>5&3
			c[1][1] = data[4]&0xfe | data[4]>>7
			c[1][2] = data[4]<<7&0x80 | data[5]>>1&0x7c
			c[1][2] |= c[1][2] >> 6
			c[2][0] = data[5]<<5&0xe0 | data[6]>>3&0x1c | data[5]>>1&3
			c[2][1] = data[6]<<3&0xf8 | data[7]>>5&0x6 | data[6]>>4&1
			c[2][2] = data[7]<<2 | data[7]>>4&3
			for y, i := 0, 0; y < 4; y++ {
				for x := 0; x < 4; x, i = x+1, i+1 {
					out[i*4+0] = clampU8((x*(int(c[1][0])-int(c[0][0])) + y*(int(c[2][0])-int(c[0][0])) + 4*int(c[0][0]) + 2) >> 2)
					out[i*4+1] = clampU8((x*(int(c[1][1])-int(c[0][1])) + y*(int(c[2][1])-int(c[0][1])) + 4*int(c[0][1]) + 2) >> 2)
					out[i*4+2] = clampU8((x*(int(c[1][2])-int(c[0][2])) + y*(int(c[2][2])-int(c[0][2])) + 4*int(c[0][2]) + 2) >> 2)
					out[i*4+3] = 0xff
				}
			}
		} else {
			code := [2]byte{data[3] >> 5, data[3] >> 2 & 7}
			table := etc1SubblockTable[data[3]&1]
			c[0][0] = r | r>>5
			c[0][1] = g | g>>5
			c[0][2] = b | b>>5
			c[1][0] = byte(int16(r) + dr)
			c[1][1] = byte(int16(g) + dg)
			c[1][2] = byte(int16(b) + db)
			c[1][0] |= c[1][0] >> 5
			c[1][1] |= c[1][1] >> 5
			c[1][2] |= c[1][2] >> 5

			for i := 0; i < 16; i, j, k = i+1, j>>1, k>>1 {
				s := table[i]
				m := int16(etc1ModifierTable[code[s]][j&1])
				if k&1 != 0 {
					m = -m
				}
				copy(out[writeOrderTable[i]*4:], applicateColor(c[s], m))
			}
		}
	} else {
		code := [2]byte{data[3] >> 5, data[3] >> 2 & 7}
		table := etc1SubblockTable[data[3]&1]
		c[0][0] = data[0]&0xf0 | data[0]>>4
		c[1][0] = data[0]&0x0f | data[0]<<4
		c[0][1] = data[1]&0xf0 | data[1]>>4
		c[1][1] = data[1]&0x0f | data[1]<<4
		c[0][2] = data[2]&0xf0 | data[2]>>4
		c[1][2] = data[2]&0x0f | data[2]<<4
		for i := 0; i < 16; i, j, k = i+1, j>>1, k>>1 {
			s := table[i]
			m := int16(etc1ModifierTable[code[s]][j&1])
			if k&1 != 0 {
				m = -m
			}
			copy(out[writeOrderTable[i]*4:], applicateColor(c[s], m))
		}
	}
}

func decode2A8Block(data []byte, out []byte) {
	if data[1]&0xf0 != 0 {
		multiplier := data[1] >> 4
		table := etc2AlphaModTable[data[1]&0xf]
		l := binary.BigEndian.Uint64(data)
		for i := 0; i < 16; i, l = i+1, l>>3 {
			out[writeOrderTable[i]*4+3] = clampU8(int(data[0]) + int(multiplier)*int(table[l&7]))
		}
	} else {
		for i := 0; i < 16; i++ {
			out[i*4+3] = data[0]
		}
	}
}

func Decode2(data []byte, width, height int) (image.Image, error) {
	xBlocks := (width + 3) / 4
	yBlocks := (height + 3) / 4

	buffer := make([]byte, 4*4*4)
	result := make([]byte, width*height*4)

	ptr := 0
	for blockY := 0; blockY < yBlocks; blockY++ {
		for blockX := 0; blockX < xBlocks; blockX, ptr = blockX+1, ptr+8 {
			decode2Block(data[ptr:], buffer)
			decoders.CopyBlockBuffer(blockX, blockY, width, height, 4, 4, buffer, result)
		}
	}

	return &image.RGBA{
		Pix:    result,
		Stride: width * 4,
		Rect:   image.Rect(0, 0, width, height),
	}, nil
}

func Decode2A8(data []byte, width, height int) (image.Image, error) {
	xBlocks := (width + 3) / 4
	yBlocks := (height + 3) / 4

	buffer := make([]byte, 4*4*4)
	result := make([]byte, width*height*4)

	ptr := 0
	for blockY := 0; blockY < yBlocks; blockY++ {
		for blockX := 0; blockX < xBlocks; blockX, ptr = blockX+1, ptr+16 {
			decode2Block(data[ptr+8:], buffer)
			decode2A8Block(data[ptr:], buffer)
			decoders.CopyBlockBuffer(blockX, blockY, width, height, 4, 4, buffer, result)
		}
	}

	return &image.RGBA{
		Pix:    result,
		Stride: width * 4,
		Rect:   image.Rect(0, 0, width, height),
	}, nil
}
