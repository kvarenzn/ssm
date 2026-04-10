// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

func LZ4Decompress(data []byte, uncompressedSize int) []byte {
	result := make([]byte, uncompressedSize+0x100)
	reader := NewBinaryReaderFromBytes(data, false)
	dataSize := reader.Len()

	ptr := 0

	for {
		token := reader.U8()

		// read literal
		literalLength := int(token >> 4)
		if literalLength == 0b1111 {
			add := reader.U8()
			for add == 0xff {
				literalLength += 0xff
				add = reader.U8()
			}
			literalLength += int(add)
		}
		copy(result[ptr:], reader.Bytes(literalLength))
		ptr += literalLength

		if reader.Position() == dataSize {
			break
		}

		// read match copy operation
		offset := reader.U16()
		if offset == 0 {
			continue
		}
		matchLength := int(token & 0b1111)
		if matchLength == 0b1111 {
			add := reader.U8()
			for add == 0xff {
				matchLength += 0xff
				add = reader.U8()
			}
			matchLength += int(add)
		}
		matchLength += 4

		// supporting overlap copy
		begin := ptr - int(offset)
		for i := range matchLength {
			result[ptr+i] = result[begin+i]
		}
	}

	return result[:uncompressedSize]
}
