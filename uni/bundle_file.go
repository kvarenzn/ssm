// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

import (
	"bytes"
	"fmt"

	"github.com/pierrec/lz4/v4"
)

type BundleFileHeader struct {
	Signature                  string
	Version                    int
	UnityVersion               string
	UnityRevision              string
	Size                       int
	CompressedBlocksInfoSize   int
	UncompressedBlocksInfoSize int
	Flags                      int
}

type BundleFileStorageBlock struct {
	CompressedSize   uint32
	UncompressedSize uint32
	Flags            uint16
}

type BundleFileNode struct {
	Offset int
	Size   int
	Flags  int
	Path   string
}

type BundleFile struct {
	Header        BundleFileHeader
	BlocksInfo    []BundleFileStorageBlock
	DirectoryInfo []BundleFileNode
	Files         []StreamFile
	Reader        *FileReader
}

func NewBundleFile(reader *FileReader) (*BundleFile, error) {
	file := &BundleFile{
		Reader: reader,
		Header: BundleFileHeader{
			Signature:     reader.CString(),
			Version:       int(reader.U32()),
			UnityVersion:  reader.CString(),
			UnityRevision: reader.CString(),
		},
	}

	signature := file.Header.Signature
	if signature == "UnityArchive" {
		return nil, fmt.Errorf("Not supported")
	} else if (signature == "UnityWeb" || signature == "UnityRaw") && file.Header.Version != 6 {
		return nil, fmt.Errorf("Not supported")
	} else if signature == "UnityFS" || ((signature == "UnityWeb" || signature == "UnityRaw") && file.Header.Version == 6) {
		file.Header.Size = int(reader.S64())

		if file.Header.Size != int(reader.Len()) {
			return nil, fmt.Errorf("file corrupted: expected %d byte(s), but got %d byte(s)", file.Header.Size, reader.reader.Len())
		}

		file.Header.CompressedBlocksInfoSize = int(reader.U32())
		file.Header.UncompressedBlocksInfoSize = int(reader.U32())
		file.Header.Flags = int(reader.U32())

		if signature != "UnityFS" {
			reader.Skip(1)
		}

		if file.Header.Version >= 7 {
			reader.Align(16)
		}

		var blockInfoBytes []byte
		if file.Header.Flags&0x80 != 0 {
			position := reader.Position()
			reader.SeekTo(-int64(file.Header.CompressedBlocksInfoSize))
			blockInfoBytes = reader.Bytes(file.Header.CompressedBlocksInfoSize)
			reader.SeekTo(position)
		} else {
			blockInfoBytes = reader.Bytes(file.Header.CompressedBlocksInfoSize)
		}

		uncompressedSize := file.Header.UncompressedBlocksInfoSize

		var uncompressedData []byte
		switch file.Header.Flags & 0x3f {
		case 1:
			return nil, fmt.Errorf("LZMA unsupported")
		case 2:
			fallthrough
		case 3:
			uncompressedData = make([]byte, uncompressedSize+0x100)
			length, err := lz4.UncompressBlock(blockInfoBytes, uncompressedData)
			if err != nil {
				return nil, err
			}

			if length != uncompressedSize {
				return nil, fmt.Errorf("LZ4 decompression error: size not correct")
			}

			uncompressedData = uncompressedData[:uncompressedSize]
		default:
			uncompressedData = blockInfoBytes
		}

		ucReader := NewBinaryReaderFromBytes(uncompressedData, true)
		ucReader.Skip(16) // uncompressed data hash
		file.BlocksInfo = []BundleFileStorageBlock{}
		blocksCount := int(ucReader.S32())
		for range blocksCount {
			file.BlocksInfo = append(file.BlocksInfo, BundleFileStorageBlock{
				UncompressedSize: ucReader.U32(),
				CompressedSize:   ucReader.U32(),
				Flags:            ucReader.U16(),
			})
		}

		directoryCount := int(ucReader.S32())
		for range directoryCount {
			file.DirectoryInfo = append(file.DirectoryInfo, BundleFileNode{
				Offset: int(ucReader.S64()),
				Size:   int(ucReader.S64()),
				Flags:  int(ucReader.U32()),
				Path:   ucReader.CString(),
			})
		}

		blockStream := bytes.NewBuffer(nil)

		if file.Header.Flags&0x200 != 0 {
			reader.Align(16)
		}

		for _, block := range file.BlocksInfo {
			switch block.Flags & 0x3f {
			case 1:
				return nil, fmt.Errorf("LZMA unsupported")
			case 2:
				fallthrough
			case 3:
				compressedData := reader.Bytes(int(block.CompressedSize))
				uncompressedData := make([]byte, block.UncompressedSize+0x100)
				length, err := lz4.UncompressBlock(compressedData, uncompressedData)
				if err != nil {
					return nil, err
				}
				if length != int(block.UncompressedSize) {
					return nil, fmt.Errorf("Uncompressed size not match")
				}
				blockStream.Write(uncompressedData[:block.UncompressedSize])
			default:
				blockStream.Write(reader.Bytes(int(block.CompressedSize)))
			}
		}

		uncompressed := blockStream.Bytes()
		file.Files = []StreamFile{}
		for _, node := range file.DirectoryInfo {
			file.Files = append(file.Files, StreamFile{
				Path:   node.Path,
				Stream: uncompressed[node.Offset : node.Offset+node.Size],
			})
		}
	}

	return file, nil
}
