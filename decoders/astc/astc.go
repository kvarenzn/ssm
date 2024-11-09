package astc

import (
	"encoding/binary"
	"image"

	"github.com/kvarenzn/ssm/decoders"
)

func decodeBlock(data []byte, blockWidth, blockHeight int, buffer []byte) {
	if data[0] == 0xfc && data[1]&1 == 1 {
		var r, g, b, a byte
		if data[1]&2 != 0 {
			r = f32ToU8(u16ToF32(binary.LittleEndian.Uint16(data[8:])))
			g = f32ToU8(u16ToF32(binary.LittleEndian.Uint16(data[10:])))
			b = f32ToU8(u16ToF32(binary.LittleEndian.Uint16(data[12:])))
			a = f32ToU8(u16ToF32(binary.LittleEndian.Uint16(data[14:])))
		} else {
			r = data[9]
			g = data[11]
			b = data[13]
			a = data[15]
		}

		for i := 0; i < blockWidth*blockHeight; i++ {
			buffer[i*4+0] = r
			buffer[i*4+1] = g
			buffer[i*4+2] = b
			buffer[i*4+3] = a
		}
	} else if data[0]&0xc3 == 0xc0 && data[1]&1 == 1 || data[0]&0xf == 0 {
		for i := 0; i < blockWidth*blockHeight; i++ {
			buffer[i*4+0] = 0xff
			buffer[i*4+1] = 0x00
			buffer[i*4+2] = 0xff
			buffer[i*4+3] = 0xff
		}
	} else {
		blockData := BlockData{
			BlockWidth:  blockWidth,
			BlockHeight: blockHeight,
		}
		blockData.DecodeBlockParams(data)
		blockData.DecodeEndpoints(data)
		blockData.DecodeWeights(data)
		if blockData.PartCount > 1 {
			blockData.SelectPartition(data)
		}
		blockData.ApplicateColor(buffer)
	}
}

func Decode(data []byte, width, height, blockWidth, blockHeight int) (image.Image, error) {
	xBlocks := (width + blockWidth - 1) / blockWidth
	yBlocks := (height + blockHeight - 1) / blockHeight
	// buffer := make([]byte, blockWidth*blockHeight*4)
	buffer := make([]byte, 144*4)
	result := make([]byte, width*height*4)
	ptr := 0
	for blockY := 0; blockY < yBlocks; blockY++ {
		for blockX := 0; blockX < xBlocks; blockX, ptr = blockX+1, ptr+16 {
			decodeBlock(data[ptr:], blockWidth, blockHeight, buffer)
			decoders.CopyBlockBuffer(blockX, blockY, width, height, blockWidth, blockHeight, buffer, result)
		}
	}

	return &image.NRGBA{
		Pix:    result,
		Stride: 4 * width,
		Rect:   image.Rect(0, 0, width, height),
	}, nil
}
