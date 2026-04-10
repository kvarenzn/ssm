// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

type SecondarySpriteTexture struct {
	Texture *PPtr
	Name    string
}

func NewSecondarySpriteTexture(reader *ObjectReader) *SecondarySpriteTexture {
	return &SecondarySpriteTexture{
		Texture: NewPPtr(reader),
		Name:    reader.CString(),
	}
}

type AABB struct {
	Center *Vector3f
	Extent *Vector3f
}

func NewAABB(reader *ObjectReader) *AABB {
	return &AABB{reader.Vector3f(), reader.Vector3f()}
}

type GfxPrimitiveType int32

const (
	GfxTriangles GfxPrimitiveType = iota
	GfxTriangleStrip
	GfxQuads
	GfxLines
	GfxLineStrip
	GfxPoints
)

type SubMesh struct {
	FirstByte     uint32
	IndexCount    uint32
	Topology      GfxPrimitiveType
	TriangleCount uint32
	BaseVertex    uint32
	FirstVertex   uint32
	VertexCount   uint32
	LocalAABB     *AABB
}

func NewSubMesh(reader *ObjectReader) *SubMesh {
	s := &SubMesh{
		FirstByte:  reader.U32(),
		IndexCount: reader.U32(),
		Topology:   GfxPrimitiveType(reader.S32()),
	}

	if reader.Version.LessThan(4) {
		s.TriangleCount = reader.U32()
	}

	if reader.Version.GreaterEqual(2017, 3) {
		s.BaseVertex = reader.U32()
	}

	if reader.Version.GreaterEqual(3) {
		s.FirstVertex = reader.U32()
		s.VertexCount = reader.U32()
		s.LocalAABB = NewAABB(reader)
	}

	return s
}

type StreamInfo struct {
	ChannelMask uint32
	Offset      uint32
	Stride      uint32
	Align       uint32
	DividerOp   uint8
	Frequency   uint16
}

func NewStreamInfo(reader *ObjectReader) *StreamInfo {
	s := &StreamInfo{
		ChannelMask: reader.U32(),
		Offset:      reader.U32(),
	}

	if reader.Version.LessThan(4) {
		s.Stride = reader.U32()
		s.Align = reader.U32()
	} else {
		s.Stride = uint32(reader.U8())
		s.DividerOp = reader.U8()
		s.Frequency = reader.U16()
	}

	return s
}

type ChannelInfo struct {
	Stream    uint8
	Offset    uint8
	Format    uint8
	Dimension uint8
}

func NewChannelInfo(reader *ObjectReader) *ChannelInfo {
	return &ChannelInfo{
		Stream:    reader.U8(),
		Offset:    reader.U8(),
		Format:    reader.U8(),
		Dimension: reader.U8() & 0xf,
	}
}

type VertexData struct {
	CurrentChannels uint32
	VertexCount     uint32
	Channels        []*ChannelInfo
	Streams         []*StreamInfo
	DataSize        []byte
}

func NewVertexData(reader *ObjectReader) *VertexData {
	v := &VertexData{}
	if reader.Version.LessThan(2018) {
		v.CurrentChannels = reader.U32()
	}

	v.VertexCount = reader.U32()

	if reader.Version.GreaterEqual(4) {
		v.Channels = ReadArray(func() *ChannelInfo {
			return NewChannelInfo(reader)
		}, reader.S32())
	}

	if reader.Version.LessThan(5) {
		streamsCount := int32(4)
		if reader.Version.GreaterEqual(4) {
			streamsCount = reader.S32()
		}

		v.Streams = ReadArray(func() *StreamInfo {
			return NewStreamInfo(reader)
		}, streamsCount)

		if reader.Version.LessThan(4) {
			v.Channels = make([]*ChannelInfo, 6)
			for i := range 6 {
				v.Channels[i] = &ChannelInfo{}
			}

			for s, stream := range v.Streams {
				mask := stream.ChannelMask
				offset := uint8(0)
				for i, channel := range v.Channels {
					if (mask>>i)&1 == 0 {
						continue
					}

					channel.Stream = uint8(s)
					channel.Offset = offset
					switch i {
					case 0:
						fallthrough
					case 1:
						channel.Format = 0
						channel.Dimension = 3
					case 2:
						channel.Format = 2
						channel.Dimension = 4
					case 3:
						fallthrough
					case 4:
						channel.Format = 0
						channel.Dimension = 2
					case 5:
						channel.Format = 0
						channel.Dimension = 4
					}

					offset += channel.Dimension * uint8(GetVertexFormatSize(ToVertexFormat(int(channel.Format), reader.Version)))
				}
			}
		}
	} else {
		streamCount := uint8(0)
		for _, c := range v.Channels {
			if streamCount < c.Stream {
				streamCount = c.Stream
			}
		}
		streamCount++
		v.Streams = make([]*StreamInfo, streamCount)

		offset := uint32(0)
		for s := range streamCount {
			channelMask := uint32(0)
			stride := uint32(0)

			for c, channel := range v.Channels {
				if channel.Stream != s || channel.Dimension <= 0 {
					continue
				}

				channelMask |= 1 << c
				stride += uint32(channel.Dimension) * uint32(GetVertexFormatSize(ToVertexFormat(int(channel.Format), reader.Version)))
			}

			v.Streams[s] = &StreamInfo{
				ChannelMask: channelMask,
				Offset:      offset,
				Stride:      stride,
				DividerOp:   0,
				Frequency:   0,
			}
			offset += v.VertexCount * stride
			offset = (offset + 15) & ^uint32(15)
		}
	}

	v.DataSize = reader.ByteArray()
	reader.Align(4)

	return v
}

type VertexChannelFormat int

const (
	VCFFloat VertexChannelFormat = iota
	VCFFloat16
	VCFColor
	VCFByte
	VCFUInt32
)

type VertexFormat2017 int

const (
	VF2017Float VertexFormat2017 = iota
	VF2017Float16
	VF2017Color
	VF2017UNorm8
	VF2017SNorm8
	VF2017UNorm16
	VF2017SNorm16
	VF2017UInt8
	VF2017SInt8
	VF2017UInt16
	VF2017SInt16
	VF2017UInt32
	VF2017SInt32
)

type VertexFormat int

const (
	VFFloat VertexFormat = iota
	VFFloat16
	VFUNorm8
	VFSNorm8
	VFUNorm16
	VFSNorm16
	VFUInt8
	VFSInt8
	VFUInt16
	VFSInt16
	VFUInt32
	VFSInt32
)

func ToVertexFormat(format int, version Version) VertexFormat {
	if version.LessThan(2017) {
		switch VertexChannelFormat(format) {
		case VCFFloat:
			return VFFloat
		case VCFFloat16:
			return VFFloat16
		case VCFColor:
			return VFUNorm8
		case VCFByte:
			return VFUInt8
		case VCFUInt32:
			return VFUInt32
		default:
			panic("?")
		}
	} else if version.LessThan(2019) {
		switch VertexFormat2017(format) {
		case VF2017Float:
			return VFFloat
		case VF2017Float16:
			return VFFloat16
		case VF2017Color:
			fallthrough
		case VF2017UNorm8:
			return VFUNorm8
		case VF2017SNorm8:
			return VFSNorm8
		case VF2017UNorm16:
			return VFUNorm16
		case VF2017SNorm16:
			return VFSNorm16
		case VF2017UInt8:
			return VFUInt8
		case VF2017SInt8:
			return VFSInt8
		case VF2017UInt16:
			return VFUInt16
		case VF2017SInt16:
			return VFSInt16
		case VF2017UInt32:
			return VFUInt16
		case VF2017SInt32:
			return VFSInt16
		default:
			panic("?")
		}
	} else {
		return VertexFormat(format)
	}
}

func GetVertexFormatSize(format VertexFormat) uint {
	switch format {
	case VFFloat:
		fallthrough
	case VFUInt32:
		fallthrough
	case VFSInt32:
		return 4

	case VFFloat16:
		fallthrough
	case VFUNorm16:
		fallthrough
	case VFSNorm16:
		fallthrough
	case VFUInt16:
		fallthrough
	case VFSInt16:
		return 2

	case VFUNorm8:
		fallthrough
	case VFSNorm8:
		fallthrough
	case VFUInt8:
		fallthrough
	case VFSInt8:
		return 1

	default:
		panic("?")
	}
}

type SpriteVertex struct {
	Pos *Vector3f
	UV  *Vector2f
}

func NewSpriteVertex(reader *ObjectReader) *SpriteVertex {
	s := &SpriteVertex{
		Pos: reader.Vector3f(),
	}

	if reader.Version.LessEqual(4, 3) {
		s.UV = reader.Vector2f()
	}

	return s
}

type BoneWeights4 struct {
	Weight    []float32
	BoneIndex []int32
}

func NewBoneWeights4(reader *ObjectReader) *BoneWeights4 {
	return &BoneWeights4{
		Weight:    ReadArray(reader.F32, 4),
		BoneIndex: ReadArray(reader.S32, 4),
	}
}

type SpritePackingRotation uint32

const (
	SPRNone SpritePackingRotation = iota
	SPRFlipHorizontal
	SPRFlipVertical
	SPRRotate180
	SPRRotate90
)

type SpritePackingMode uint32

const (
	SPMTight SpritePackingMode = iota
	SPMRectangle
)

type SpriteMeshType uint32

const (
	SMTFullRect SpriteMeshType = iota
	SMTTight
)

type SpriteSettings struct {
	SettingsRaw uint32

	Packed          uint32
	PackingMode     SpritePackingMode
	PackingRotation SpritePackingRotation
	MeshType        SpriteMeshType
}

func NewSpriteSettings(reader *ObjectReader) *SpriteSettings {
	settingsRaw := reader.U32()
	return &SpriteSettings{
		SettingsRaw: settingsRaw,

		Packed:          settingsRaw & 1,
		PackingMode:     SpritePackingMode((settingsRaw >> 1) & 1),
		PackingRotation: SpritePackingRotation((settingsRaw >> 2) & 0xf),
		MeshType:        SpriteMeshType((settingsRaw >> 6) & 1),
	}
}

type SpriteRenderData struct {
	Texture             *PPtr
	AlphaTexture        *PPtr
	SecondaryTextures   []*SecondarySpriteTexture
	SubMeshes           []*SubMesh
	IndexBuffer         []byte
	VertexData          *VertexData
	Vertices            []*SpriteVertex
	Indices             []uint16
	Bindpose            []*Matrix4x4f
	SourceSkin          []*BoneWeights4
	TextureRect         *Rectf
	TextureRectOffset   *Vector2f
	AtlasRectOffset     *Vector2f
	SettingsRaw         *SpriteSettings
	UVTransform         *Vector4f
	DownscaleMultiplier float32
}

func NewSpriteRenderData(reader *ObjectReader) *SpriteRenderData {
	d := &SpriteRenderData{
		Texture: NewPPtr(reader),
	}

	if reader.Version.GreaterEqual(5, 2) {
		d.AlphaTexture = NewPPtr(reader)
	}

	if reader.Version.GreaterEqual(2019) {
		d.SecondaryTextures = ReadArray(func() *SecondarySpriteTexture {
			return NewSecondarySpriteTexture(reader)
		}, reader.S32())
	}

	if reader.Version.GreaterEqual(5, 6) {
		d.SubMeshes = ReadArray(func() *SubMesh {
			return NewSubMesh(reader)
		}, reader.S32())
		d.IndexBuffer = reader.ByteArray()
		reader.Align(4)

		d.VertexData = NewVertexData(reader)
	} else {
		d.Vertices = ReadArray(func() *SpriteVertex {
			return NewSpriteVertex(reader)
		}, reader.S32())
		d.Indices = reader.U16Array()
		reader.Align(4)
	}

	if reader.Version.GreaterEqual(2018) {
		d.Bindpose = reader.Matrix4x4fArray()

		if reader.Version.LessThan(2018, 2) {
			d.SourceSkin = ReadArray(func() *BoneWeights4 {
				return NewBoneWeights4(reader)
			}, reader.S32())
		}
	}

	d.TextureRect = NewRectf(reader.BinaryReader)
	d.TextureRectOffset = reader.Vector2f()
	if reader.Version.GreaterEqual(5, 6) {
		d.AtlasRectOffset = reader.Vector2f()
	}

	d.SettingsRaw = NewSpriteSettings(reader)
	if reader.Version.GreaterEqual(4, 5) {
		d.UVTransform = reader.Vector4f()
	}

	if reader.Version.GreaterEqual(2017) {
		d.DownscaleMultiplier = reader.F32()
	}

	return d
}

type Sprite struct {
	*NamedObject
	Rect          *Rectf
	Offset        *Vector2f
	Border        *Vector4f
	PixelsToUnits float32
	Pivot         *Vector2f
	Extrude       uint32
	IsPolygon     bool
	RenderDataKey *struct {
		Key   []byte
		Value int64
	}
	AtlasTags    []string
	SpriteAtlas  *PPtr
	RenderData   *SpriteRenderData
	PhysicsShape [][]*Vector2f
}

func NewSprite(reader *ObjectReader) *Sprite {
	s := &Sprite{
		NamedObject: NewNamedObject(reader),
		Rect:        NewRectf(reader.BinaryReader),
		Offset:      reader.Vector2f(),
	}

	if reader.Version.GreaterEqual(4, 5) {
		s.Border = reader.Vector4f()
	}

	s.PixelsToUnits = reader.F32()
	if reader.Version.GreaterEqual(5, 4, 1, 3) {
		s.Pivot = reader.Vector2f()
	}

	s.Extrude = reader.U32()
	if reader.Version.GreaterEqual(5, 3) {
		s.IsPolygon = reader.Bool()
		reader.Align(4)
	}

	if reader.Version.GreaterEqual(2017) {
		key := reader.Bytes(16)
		value := reader.S64()
		s.RenderDataKey = &struct {
			Key   []byte
			Value int64
		}{
			Key:   key,
			Value: value,
		}

		s.AtlasTags = reader.StringArray()
		s.SpriteAtlas = NewPPtr(reader)
	}

	s.RenderData = NewSpriteRenderData(reader)

	if reader.Version.GreaterEqual(2017) {
		s.PhysicsShape = ReadArray(func() []*Vector2f {
			return reader.Vector2fArray()
		}, reader.S32())
	}

	return s
}

type SpriteAtlasData struct {
	Texture             *PPtr
	AlphaTexture        *PPtr
	TextureRect         *Rectf
	TextureRectOffset   *Vector2f
	AtlasRectOffset     *Vector2f
	UVTransform         *Vector4f
	DownscaleMultiplier float32
	SettingsRaw         *SpriteSettings
	SecondaryTextures   []*SecondarySpriteTexture
}

func NewSpriteAtlasData(reader *ObjectReader) *SpriteAtlasData {
	s := &SpriteAtlasData{
		Texture:           NewPPtr(reader),
		AlphaTexture:      NewPPtr(reader),
		TextureRect:       NewRectf(reader.BinaryReader),
		TextureRectOffset: reader.Vector2f(),
	}

	if reader.Version.GreaterEqual(2017, 2) {
		s.AtlasRectOffset = reader.Vector2f()
	}

	s.UVTransform = reader.Vector4f()
	s.DownscaleMultiplier = reader.F32()
	s.SettingsRaw = NewSpriteSettings(reader)
	if reader.Version.GreaterEqual(2020, 2) {
		s.SecondaryTextures = ReadArray(func() *SecondarySpriteTexture {
			return NewSecondarySpriteTexture(reader)
		}, reader.S32())
		reader.Align(4)
	}

	return s
}

type SpriteAtlas struct {
	*NamedObject
	PackedSprites            []*PPtr
	PackedSpriteNamesToIndex []string
	RenderDataMap            map[*struct {
		Key   []byte
		Value int64
	}]*SpriteAtlasData
	Tag       string
	IsVariant bool
}

func NewSpriteAtlas(reader *ObjectReader) *SpriteAtlas {
	s := &SpriteAtlas{
		NamedObject: NewNamedObject(reader),
		PackedSprites: ReadArray(func() *PPtr {
			return NewPPtr(reader)
		}, reader.S32()),
		PackedSpriteNamesToIndex: reader.StringArray(),
	}

	s.RenderDataMap = map[*struct {
		Key   []byte
		Value int64
	}]*SpriteAtlasData{}

	for range reader.S32() {
		guid := reader.Bytes(16)
		id := reader.S64()
		value := NewSpriteAtlasData(reader)
		s.RenderDataMap[&struct {
			Key   []byte
			Value int64
		}{
			Key:   guid,
			Value: id,
		}] = value
	}
	s.Tag = reader.AlignedString()
	s.IsVariant = reader.Bool()
	reader.Align(4)
	return s
}
