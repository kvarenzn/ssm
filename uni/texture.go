// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

import (
	"fmt"

	"github.com/kvarenzn/ssm/optional"
)

type Texture struct {
	*NamedObject
	ForcedFallbackFormat   optional.Optional[int32]
	DownscaleFallback      optional.Optional[bool]
	IsAlphaChannelOptional optional.Optional[bool]
}

func NewTexture(reader *ObjectReader) *Texture {
	t := &Texture{
		NamedObject: NewNamedObject(reader),
	}

	if t.Version.GreaterEqual(2017, 3) {
		t.ForcedFallbackFormat = optional.Some(reader.S32())
		t.DownscaleFallback = optional.Some(reader.Bool())
		if t.Version.GreaterEqual(2020, 2) {
			t.IsAlphaChannelOptional = optional.Some(reader.Bool())
		}
		reader.Align(4)
	}

	return t
}

type TextureFormat int

const (
	Alpha8 TextureFormat = iota + 1
	ARGB4444
	RGB24
	RGBA32
	ARGB32
	ARGBFloat
	RGB565
	BGR24
	R16
	DXT1
	DXT3
	DXT5
	RGBA4444
	BGRA32
	RHalf
	RGHalf
	RGBAHalf
	RFloat
	RGFloat
	RGBAFloat
	YUY2
	RGB9e5Float
	RGBFloat
	BC6H
	BC7
	BC4
	BC5
	DXT1Crunched
	DXT5Crunched
	PVRTC_RGB2
	PVRTC_RGBA2
	PVRTC_RGB4
	PVRTC_RGBA4
	ETC_RGB4
	ATC_RGB4
	ATC_RGBA8
	EAC_R TextureFormat = iota + 5
	EAC_R_SIGNED
	EAC_RG
	EAC_RG_SIGNED
	ETC2_RGB
	ETC2_RGBA1
	ETC2_RGBA8
	ASTC_RGB_4x4
	ASTC_RGB_5x5
	ASTC_RGB_6x6
	ASTC_RGB_8x8
	ASTC_RGB_10x10
	ASTC_RGB_12x12
	ASTC_RGBA_4x4
	ASTC_RGBA_5x5
	ASTC_RGBA_6x6
	ASTC_RGBA_8x8
	ASTC_RGBA_10x10
	ASTC_RGBA_12x12
	ETC_RGB4_3DS
	ETC_RGBA8_3DS
	RG16
	R8
	ETC_RGB4Crunched
	ETC2_RGBA8Crunched
	ASTC_HDR_4x4
	ASTC_HDR_5x5
	ASTC_HDR_6x6
	ASTC_HDR_8x8
	ASTC_HDR_10x10
	ASTC_HDR_12x12
	RG32
	RGB48
	RGBA64
)

func (f TextureFormat) String() string {
	switch f {
	case Alpha8:
		return "Alpha8"
	case ARGB4444:
		return "ARGB4444"
	case RGB24:
		return "RGB24"
	case RGBA32:
		return "RGBA32"
	case ARGB32:
		return "ARGB32"
	case ARGBFloat:
		return "ARGBFloat"
	case RGB565:
		return "RGB565"
	case BGR24:
		return "BGR24"
	case R16:
		return "R16"
	case DXT1:
		return "DXT1"
	case DXT3:
		return "DXT3"
	case DXT5:
		return "DXT5"
	case RGBA4444:
		return "RGBA4444"
	case BGRA32:
		return "BGRA32"
	case RHalf:
		return "RHalf"
	case RGHalf:
		return "RGHalf"
	case RGBAHalf:
		return "RGBAHalf"
	case RFloat:
		return "RFloat"
	case RGFloat:
		return "RGFloat"
	case RGBAFloat:
		return "RGBAFloat"
	case YUY2:
		return "YUY2"
	case RGB9e5Float:
		return "RGB9e5Float"
	case RGBFloat:
		return "RGBFloat"
	case BC6H:
		return "BC6H"
	case BC7:
		return "BC7"
	case BC4:
		return "BC4"
	case BC5:
		return "BC5"
	case DXT1Crunched:
		return "DXT1Crunched"
	case DXT5Crunched:
		return "DXT5Crunched"
	case PVRTC_RGB2:
		return "PVRTC_RGB2"
	case PVRTC_RGBA2:
		return "PVRTC_RGBA2"
	case PVRTC_RGB4:
		return "PVRTC_RGB4"
	case PVRTC_RGBA4:
		return "PVRTC_RGBA4"
	case ETC_RGB4:
		return "ETC_RGB4"
	case ATC_RGB4:
		return "ATC_RGB4"
	case ATC_RGBA8:
		return "ATC_RGBA8"
	case EAC_R:
		return "EAC_R"
	case EAC_R_SIGNED:
		return "EAC_R_SIGNED"
	case EAC_RG:
		return "EAC_RG"
	case EAC_RG_SIGNED:
		return "EAC_RG_SIGNED"
	case ETC2_RGB:
		return "ETC2_RGB"
	case ETC2_RGBA1:
		return "ETC2_RGBA1"
	case ETC2_RGBA8:
		return "ETC2_RGBA8"
	case ASTC_RGB_4x4:
		return "ASTC_RGB_4x4"
	case ASTC_RGB_5x5:
		return "ASTC_RGB_5x5"
	case ASTC_RGB_6x6:
		return "ASTC_RGB_6x6"
	case ASTC_RGB_8x8:
		return "ASTC_RGB_8x8"
	case ASTC_RGB_10x10:
		return "ASTC_RGB_10x10"
	case ASTC_RGB_12x12:
		return "ASTC_RGB_12x12"
	case ASTC_RGBA_4x4:
		return "ASTC_RGBA_4x4"
	case ASTC_RGBA_5x5:
		return "ASTC_RGBA_5x5"
	case ASTC_RGBA_6x6:
		return "ASTC_RGBA_6x6"
	case ASTC_RGBA_8x8:
		return "ASTC_RGBA_8x8"
	case ASTC_RGBA_10x10:
		return "ASTC_RGBA_10x10"
	case ASTC_RGBA_12x12:
		return "ASTC_RGBA_12x12"
	case ETC_RGB4_3DS:
		return "ETC_RGB4_3DS"
	case ETC_RGBA8_3DS:
		return "ETC_RGBA8_3DS"
	case RG16:
		return "RG16"
	case R8:
		return "R8"
	case ETC_RGB4Crunched:
		return "ETC_RGB4Crunched"
	case ETC2_RGBA8Crunched:
		return "ETC2_RGBA8Crunched"
	case ASTC_HDR_4x4:
		return "ASTC_HDR_4x4"
	case ASTC_HDR_5x5:
		return "ASTC_HDR_5x5"
	case ASTC_HDR_6x6:
		return "ASTC_HDR_6x6"
	case ASTC_HDR_8x8:
		return "ASTC_HDR_8x8"
	case ASTC_HDR_10x10:
		return "ASTC_HDR_10x10"
	case ASTC_HDR_12x12:
		return "ASTC_HDR_12x12"
	case RG32:
		return "RG32"
	case RGB48:
		return "RGB48"
	case RGBA64:
		return "RGBA64"
	}

	return fmt.Sprintf("Unknown Format (%d)", f)
}

type TextureSettings struct {
	FilterMode int32
	Aniso      int32
	MipBias    float32
	WrapMode   int32
}

func NewTextureSettings(reader *ObjectReader) *TextureSettings {
	s := &TextureSettings{}

	s.FilterMode = reader.S32()
	s.Aniso = reader.S32()
	s.MipBias = reader.F32()

	s.WrapMode = reader.S32()
	if reader.Version.GreaterEqual(2017) {
		reader.S32() // WrapV
		reader.S32() // WrapW
	}

	return s
}

type StreamingInfo struct {
	Offset int64
	Size   uint32
	Path   string
}

func NewStreamingInfo(reader *ObjectReader) *StreamingInfo {
	i := &StreamingInfo{}
	if reader.Version.GreaterEqual(2020, 1) {
		i.Offset = reader.S64()
	} else {
		i.Offset = int64(reader.U32())
	}

	i.Size = reader.U32()
	i.Path = reader.AlignedString()

	return i
}

type Texture2D struct {
	*Texture
	Width                    int32
	Height                   int32
	CompleteImageSize        int32
	MipsStripped             optional.Optional[int32]
	Format                   TextureFormat
	Mipmap                   optional.Optional[bool]
	MipCount                 optional.Optional[int32]
	IsReadable               optional.Optional[bool]
	IsPreProcessed           optional.Optional[bool]
	IgnoreMasterTextureLimit optional.Optional[bool]
	IgnoreMipmapLimit        optional.Optional[bool]
	MipmapLimitGroupName     optional.Optional[string]
	ReadAllowed              optional.Optional[bool]
	StreamingMipmaps         optional.Optional[bool]
	StreamingMipmapsPriority optional.Optional[int32]
	ImageCount               int32
	TextureDimension         int32
	LightmapFormat           optional.Optional[int32]
	ColorSpace               optional.Optional[int32]
	TextureSettings          *TextureSettings
	ImageData                *ResourceReader
	Info                     *StreamingInfo
}

func NewTexture2D(reader *ObjectReader) *Texture2D {
	t := &Texture2D{
		Texture: NewTexture(reader),
	}

	t.Width = reader.S32()
	t.Height = reader.S32()

	t.CompleteImageSize = reader.S32()
	if t.Version.GreaterEqual(2020, 1) {
		t.MipsStripped = optional.Some(reader.S32())
	}

	t.Format = TextureFormat(reader.S32())
	if t.Version.GreaterEqual(5, 2) {
		t.MipCount = optional.Some(reader.S32())
	} else {
		t.Mipmap = optional.Some(reader.Bool())
	}

	if t.Version.GreaterEqual(2, 6) {
		t.IsReadable = optional.Some(reader.Bool())
	}

	if t.Version.GreaterEqual(2020, 1) {
		t.IsPreProcessed = optional.Some(reader.Bool())
	}

	if t.Version.GreaterEqual(2022, 2) {
		t.IgnoreMipmapLimit = optional.Some(reader.Bool())
		reader.Align(4)
	} else if t.Version.GreaterEqual(2019, 3) {
		t.IgnoreMasterTextureLimit = optional.Some(reader.Bool())
	}

	if (t.SerializedType != nil && t.SerializedType.Type != nil && t.SerializedType.Type.ContainsNamePath("Base.m_MipmapLimitGroupName.Array.data")) || t.Version.GreaterEqual(2022, 2) {
		t.MipmapLimitGroupName = optional.Some(reader.AlignedString())
	}

	if t.Version.GreaterEqual(3) && t.Version.LessEqual(5, 4, 999) {
		t.ReadAllowed = optional.Some(reader.Bool())
	}

	if t.Version.GreaterEqual(2018, 2) {
		t.StreamingMipmaps = optional.Some(reader.Bool())
	}

	reader.Align(4)
	if t.Version.GreaterEqual(2018, 2) {
		t.StreamingMipmapsPriority = optional.Some(reader.S32())
	}

	t.ImageCount = reader.S32()
	t.TextureDimension = reader.S32()

	t.TextureSettings = NewTextureSettings(reader)

	if t.Version.GreaterEqual(3) {
		t.LightmapFormat = optional.Some(reader.S32())
	}

	if t.Version.GreaterEqual(3, 5) {
		t.ColorSpace = optional.Some(reader.S32())
	}

	if t.Version.GreaterEqual(2020, 2) {
		reader.Bytes(int(reader.S32())) // PlatformBlob
		reader.Align(4)
	}

	imageDataSize := reader.S32()
	if imageDataSize == 0 && t.Version.GreaterEqual(5, 3) {
		t.Info = NewStreamingInfo(reader)
	}

	if t.Info == nil || t.Info.Path == "" {
		t.ImageData = NewResourceReader(reader.BinaryReader, reader.Position(), int64(imageDataSize))
	} else {
		t.ImageData = NewResourceReaderWithAssetFile(t.Info.Path, t.AssetFile, t.Info.Offset, int64(t.Info.Size))
	}

	return t
}
