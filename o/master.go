package o

import (
	"bytes"
	"compress/gzip"
	"io"
)

func DecryptMaster(data []byte, gunzip bool) []byte {
	const blockSize = 32

	body := data[64:]
	w, nr := expand(masterKey, 8)
	out := bytes.NewBuffer(nil)
	prev := masterIV
	for i := 0; i < len(body)-blockSize+1; i += blockSize {
		dec := decodeBlock(body[i:i+blockSize], w, nr, 8)
		for j := range min(len(dec), len(prev)) {
			out.WriteByte(dec[j] ^ prev[j])
		}
		prev = body[i : i+blockSize]
	}

	res := out.Bytes()
	if gunzip {
		r, err := gzip.NewReader(bytes.NewReader(res))
		if err != nil {
			return res
		}

		// Python 用 zlib.decompressobj(31)，遇到 gzip 流结束后的填充字节会停止；
		// Go 的 gzip.Reader 默认会尝试读取下一个成员并对填充报错，这里关闭多流模式。
		r.Multistream(false)
		decomp, err := io.ReadAll(r)
		if err != nil {
			return res
		}

		return decomp
	}

	return res
}
