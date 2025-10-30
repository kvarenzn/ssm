// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

type TextAsset struct {
	*NamedObject
	Content []byte
}

func NewTextAsset(reader *ObjectReader) *TextAsset {
	t := &TextAsset{
		NamedObject: NewNamedObject(reader),
	}

	t.Content = reader.Bytes(int(reader.S32()))

	return t
}
