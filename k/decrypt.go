package k

import (
	"bytes"
	"fmt"
	"io"
)

var magicHeader = []byte{16, 0, 0, 0}

type assetFile struct {
	r     io.Reader
	count int
}

func NewSekaiAssetFile(reader io.Reader) (*assetFile, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(buf, magicHeader) {
		return nil, fmt.Errorf("Invalid magic header: expected %v, but got %v", magicHeader, buf)
	}

	return &assetFile{
		r: reader,
	}, nil
}

func (f *assetFile) Read(buf []byte) (n int, err error) {
	n, err = f.r.Read(buf)
	for i := range n {
		p := f.count + i
		if p >= 128 {
			break
		}

		if p%8 < 5 {
			buf[i] = ^buf[i]
		}
	}

	f.count += n
	return
}

func (f *assetFile) Close() error {
	if c, ok := f.r.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
