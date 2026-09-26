package o

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"slices"
)

func nonceOf(bundleName string) []byte {
	h := sha256.New()
	h.Write(nonceSeed)
	h.Write([]byte(bundleName))
	return h.Sum(nil)[:8]
}

func aesECBEncrypt(block cipher.Block, data []byte) []byte {
	bs := block.BlockSize()

	result := make([]byte, len(data))
	for i := 0; i < len(data); i += bs {
		block.Encrypt(result[i:i+bs], data[i:i+bs])
	}

	return result
}

func aesBlocks(inputs [][]byte) ([]byte, error) {
	key, err := aes.NewCipher(assetBundleCryptKey)
	if err != nil {
		return nil, err
	}

	result := bytes.NewBuffer(nil)
	for i := range len(inputs) {
		result.Write(aesECBEncrypt(key, inputs[i]))
	}

	return result.Bytes(), nil
}

func DecryptAssetBundle(data []byte, nonce []byte) ([]byte, error) {
	if len(nonce) != 8 {
		return nil, fmt.Errorf("nonce的长度必须为8")
	}

	n := min(len(data), encryptedHeaderSize)
	decryptedHeader := make([]byte, n)
	copy(decryptedHeader, data)
	l := (n + 15) / 16
	blocks := make([][]byte, l)
	for i := range l {
		blocks[i] = make([]byte, 8+8)
		copy(blocks[i], nonce)
		binary.BigEndian.PutUint64(blocks[i][8:], uint64(i))
	}
	ks, err := aesBlocks(blocks)
	if err != nil {
		return nil, err
	}

	for i := range decryptedHeader {
		decryptedHeader[i] ^= ks[i]
	}

	return slices.Concat(decryptedHeader, data[n:]), nil
}
