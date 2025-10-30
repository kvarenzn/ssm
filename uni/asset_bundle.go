// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

type AssetInfo struct {
	PreloadIndex int32
	PreloadSize  int32
	Asset        *PPtr
}

func NewAssetInfo(reader *ObjectReader) *AssetInfo {
	return &AssetInfo{
		PreloadIndex: reader.S32(),
		PreloadSize:  reader.S32(),
		Asset:        NewPPtr(reader),
	}
}

type AssetBundle struct {
	*NamedObject
	PreloadTable []*PPtr
	Container    []*struct {
		Key   string
		Value *AssetInfo
	}
}

func NewAssetBundle(reader *ObjectReader) *AssetBundle {
	b := &AssetBundle{
		NamedObject: NewNamedObject(reader),
	}

	for range reader.S32() {
		b.PreloadTable = append(b.PreloadTable, NewPPtr(reader))
	}

	for range reader.S32() {
		b.Container = append(b.Container, &struct {
			Key   string
			Value *AssetInfo
		}{
			Key:   reader.AlignedString(),
			Value: NewAssetInfo(reader),
		})
	}

	return b
}
