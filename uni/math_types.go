// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

type Vector2f struct {
	X float32
	Y float32
}

func (reader *BinaryReader) Vector2f() *Vector2f {
	return &Vector2f{reader.F32(), reader.F32()}
}

func (reader *BinaryReader) Vector2fArray() []*Vector2f {
	return ReadArray(reader.Vector2f, reader.S32())
}

type Vector3f struct {
	X float32
	Y float32
	Z float32
}

func (reader *BinaryReader) Vector3f() *Vector3f {
	return &Vector3f{reader.F32(), reader.F32(), reader.F32()}
}

type Vector4f struct {
	X float32
	Y float32
	Z float32
	W float32
}

func (reader *BinaryReader) Vector4f() *Vector4f {
	return &Vector4f{reader.F32(), reader.F32(), reader.F32(), reader.F32()}
}

type Rectf struct {
	X      float32
	Y      float32
	Width  float32
	Height float32
}

func NewRectf(reader *BinaryReader) *Rectf {
	return &Rectf{reader.F32(), reader.F32(), reader.F32(), reader.F32()}
}

type Matrix4x4f struct {
	M00, M10, M20, M30 float32
	M01, M11, M21, M31 float32
	M02, M12, M22, M32 float32
	M03, M13, M23, M33 float32
}

func NewMatrix4x4f(values []float32) *Matrix4x4f {
	return &Matrix4x4f{values[0], values[1], values[2], values[3], values[4], values[5], values[6], values[7], values[8], values[9], values[10], values[11], values[12], values[13], values[14], values[15]}
}

func (reader *BinaryReader) Matrix4x4f() *Matrix4x4f {
	return NewMatrix4x4f(ReadArray(reader.F32, 16))
}

func (reader *BinaryReader) Matrix4x4fArray() []*Matrix4x4f {
	return ReadArray(reader.Matrix4x4f, reader.S32())
}
