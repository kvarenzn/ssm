// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package k

import (
	"bufio"
	"bytes"
	"io"
)

var magicHeader = []byte{16, 0, 0, 0}

type assetFile struct {
	r     io.Reader
	count int
}

func NewSekaiAssetFile(reader io.Reader) (io.Reader, error) {
	br := bufio.NewReader(reader)
	buf, err := br.Peek(4)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(buf, magicHeader) {
		return br, nil
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
