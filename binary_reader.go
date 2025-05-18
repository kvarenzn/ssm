package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/kvarenzn/ssm/log"
)

type BinaryReader struct {
	reader    *bytes.Reader
	bigEndian bool
}

func NewBinaryReaerFromReader(reader *bytes.Reader, bigEndian bool) *BinaryReader {
	reader.Seek(0, io.SeekStart)
	return &BinaryReader{
		reader:    reader,
		bigEndian: bigEndian,
	}
}

func NewBinaryReaderFromBytes(data []byte, bigEndian bool) *BinaryReader {
	return &BinaryReader{
		reader:    bytes.NewReader(data),
		bigEndian: bigEndian,
	}
}

func (r *BinaryReader) Bool() bool {
	data, err := r.reader.ReadByte()
	if err != nil {
		log.Fatal(err)
	}

	if data == 0 {
		return false
	}

	return true
}

func (r *BinaryReader) Position() int64 {
	return r.reader.Size() - int64(r.reader.Len())
}

func (r *BinaryReader) SeekTo(newPosition int64) error {
	pos, err := r.reader.Seek(newPosition, io.SeekStart)
	if err != nil {
		return err
	}

	if pos != newPosition {
		return fmt.Errorf("seek failed")
	}
	return nil
}

func (r *BinaryReader) Skip(offset int64) {
	r.reader.Seek(offset, io.SeekCurrent)
}

func (r *BinaryReader) Len() int64 {
	return r.reader.Size()
}

func (r *BinaryReader) IsBigEndian() bool {
	return r.bigEndian
}

func (r *BinaryReader) SetBigEndian(bigEndian bool) {
	r.bigEndian = bigEndian
}

func (r *BinaryReader) Align(size int64) int64 {
	pos := r.Position()
	r.reader.Seek((-pos%size+size)%size, io.SeekCurrent)
	return pos
}

func (r *BinaryReader) Bytes(count int) []byte {
	result := make([]byte, count)
	c, err := r.reader.Read(result)
	if err != nil {
		if err == io.EOF {
			return result
		}
		log.Fatal(err)
	}

	if c != count {
		log.Fatalf("expect to read %d bytes, but got %d bytes", count, c)
	}
	return result
}

func (r *BinaryReader) FixedString(length int) string {
	data := r.Bytes(length)
	return string(data)
}

func (r *BinaryReader) CString() string {
	return string(r.Chars())
}

func (r *BinaryReader) Chars() []byte {
	result := []byte{}
	b, err := r.reader.ReadByte()
	if err != nil {
		log.Fatal(err)
	}

	for b != 0 {
		result = append(result, b)
		b, err = r.reader.ReadByte()
		if err != nil {
			log.Fatal(err)
		}
	}

	return result
}

func (r *BinaryReader) CharsWithMaxSize(maxSize int) []byte {
	result := []byte{}

	b, err := r.reader.ReadByte()
	if err != nil {
		if err == io.EOF {
			return []byte{}
		}

		log.Fatal(err)
	}

	size := 0

	for b != 0 && size < maxSize {
		result = append(result, b)
		b, err = r.reader.ReadByte()
		if err != nil {
			log.Fatal(err)
		}
		size++
	}

	return result
}

func (r *BinaryReader) AlignedString() string {
	res := r.FixedString(int(r.S32()))
	r.Align(4)
	return res
}

func (r *BinaryReader) U8() uint8 {
	b, err := r.reader.ReadByte()
	if err != nil {
		log.Fatal(err)
	}

	return b
}

func (r *BinaryReader) S8() int8 {
	return int8(r.U8())
}

func (r *BinaryReader) U16() uint16 {
	if r.bigEndian {
		return binary.BigEndian.Uint16(r.Bytes(2))
	} else {
		return binary.LittleEndian.Uint16(r.Bytes(2))
	}
}

func (r *BinaryReader) S16() int16 {
	return int16(r.U16())
}

func (r *BinaryReader) U32() uint32 {
	if r.bigEndian {
		return binary.BigEndian.Uint32(r.Bytes(4))
	} else {
		return binary.LittleEndian.Uint32(r.Bytes(4))
	}
}

func (r *BinaryReader) S32() int32 {
	return int32(r.U32())
}

func (r *BinaryReader) U64() uint64 {
	if r.bigEndian {
		return binary.BigEndian.Uint64(r.Bytes(8))
	} else {
		return binary.LittleEndian.Uint64(r.Bytes(8))
	}
}

func (r *BinaryReader) S64() int64 {
	return int64(r.U64())
}

func (r *BinaryReader) F32() float32 {
	return math.Float32frombits(r.U32())
}

func (r *BinaryReader) F64() float64 {
	return math.Float64frombits(r.U64())
}
