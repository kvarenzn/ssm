// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"image"

	"github.com/kvarenzn/ssm/decoders/astc"
	"github.com/kvarenzn/ssm/decoders/etc"
	"github.com/kvarenzn/ssm/uni"
)

func DecodeTexture2D(texture *uni.Texture2D) (image.Image, error) {
	if texture.ImageData.Size == 0 || texture.Width == 0 || texture.Height == 0 {
		return nil, fmt.Errorf("Invalid texture data")
	}

	switch texture.Format {
	case uni.Alpha8:
		return &image.Alpha{
			Pix:    texture.ImageData.GetData(),
			Stride: int(texture.Width),
			Rect:   image.Rect(0, 0, int(texture.Width), int(texture.Height)),
		}, nil
	case uni.RGB24:
		data := texture.ImageData.GetData()
		stride := int(texture.Width) * 4
		buffer := make([]byte, texture.Width*texture.Height*4)
		ptr := int(texture.Height-1) * stride
		i := 0
		for y := int(texture.Height) - 1; y >= 0; y-- {
			for x := 0; x < int(texture.Width); x++ {
				copy(buffer[ptr:ptr+3], data[i:i+3])
				buffer[ptr+3] = 255
				i += 3
				ptr += 4
			}
			ptr -= 2 * stride
		}
		return &image.NRGBA{
			Pix:    buffer,
			Stride: stride,
			Rect:   image.Rect(0, 0, int(texture.Width), int(texture.Height)),
		}, nil
	case uni.RGBA32:
		data := texture.ImageData.GetData()
		lineWidth := int(texture.Width) * 4
		buffer := make([]byte, lineWidth)
		for i := 0; i < int(texture.Height)/2; i++ {
			copy(buffer, data[i*lineWidth:])
			copy(data[i*lineWidth:(i+1)*lineWidth], data[(int(texture.Height)-1-i)*lineWidth:])
			copy(data[(int(texture.Height)-1-i)*lineWidth:], buffer)
		}
		return &image.NRGBA{
			Pix:    data,
			Stride: int(texture.Width) * 4,
			Rect:   image.Rect(0, 0, int(texture.Width), int(texture.Height)),
		}, nil
	case uni.ARGB32:
		data := texture.ImageData.GetData()
		for i := 0; i < len(data); i += 4 {
			data[i+0], data[i+1], data[i+2], data[i+3] = data[i+1], data[i+2], data[i+3], data[i+0]
			data[i+2], data[i+3] = data[i+3], data[i+2]
		}
		return &image.NRGBA{
			Pix:    data,
			Stride: int(texture.Width) * 4,
			Rect:   image.Rect(0, 0, int(texture.Width), int(texture.Height)),
		}, nil
	case uni.ETC_RGB4:
		return etc.Decode1(texture.ImageData.GetData(), int(texture.Width), int(texture.Height))
	case uni.ETC2_RGB:
		return etc.Decode2(texture.ImageData.GetData(), int(texture.Width), int(texture.Height))
	case uni.ETC2_RGBA8:
		return etc.Decode2A8(texture.ImageData.GetData(), int(texture.Width), int(texture.Height))
	case uni.ASTC_RGB_4x4:
		return astc.Decode(texture.ImageData.GetData(), int(texture.Width), int(texture.Height), 4, 4)
	case uni.ASTC_RGB_6x6:
		return astc.Decode(texture.ImageData.GetData(), int(texture.Width), int(texture.Height), 6, 6)
	case uni.ASTC_RGB_8x8:
		return astc.Decode(texture.ImageData.GetData(), int(texture.Width), int(texture.Height), 8, 8)
	case uni.ASTC_RGB_10x10:
		return astc.Decode(texture.ImageData.GetData(), int(texture.Width), int(texture.Height), 10, 10)
	case uni.ASTC_RGB_12x12:
		return astc.Decode(texture.ImageData.GetData(), int(texture.Width), int(texture.Height), 12, 12)
	}

	return nil, fmt.Errorf("texture format %s not supported now", texture.Format)
}
