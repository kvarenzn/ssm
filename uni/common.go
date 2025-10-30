// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/kvarenzn/ssm/optional"
)

type TypeTreeNode struct {
	Type        string
	Name        string
	ByteSize    int
	Index       int
	TypeFlags   int
	Version     int
	MetaFlag    optional.Optional[int]
	Level       int
	RefTypeHash optional.Optional[uint64]
}

type TypeTree struct {
	Nodes []*TypeTreeNode
}

func (t *TypeTree) ContainsNamePath(namePath string) bool {
	paths := []string{}
	namePaths := []string{}

	for _, node := range t.Nodes {
		level := node.Level
		nam := node.Name

		namePaths = namePaths[:level]

		if level > 0 {
			nam = namePaths[level-1] + "." + nam
		}

		namePaths = append(namePaths, nam)

		paths = append(paths, namePaths[level])
	}

	for _, p := range paths {
		if strings.Contains(p, namePath) {
			return true
		}
	}
	return false
}

type SerializedType struct {
	ClassID          ClassID
	IsStrippedType   optional.Optional[bool]
	ScriptTypeIndex  optional.Optional[uint16]
	Type             *TypeTree
	ScriptID         []byte
	OldTypeHash      []byte
	TypeDependencies []int32
	ClassName        string
	Namespace        string
	AsmName          string
}

type ObjectInfo struct {
	ByteStart      int64
	ByteSize       uint32
	TypeID         int
	ClassID        ClassID
	PathID         int64
	IsDestroyed    optional.Optional[uint16]
	Stripped       optional.Optional[bool]
	SerializedType *SerializedType
}

type ObjectReader struct {
	*BinaryReader
	AssetFile      *SerializedFile
	PathID         int64
	ByteStart      int64
	ByteSize       uint32
	ClassID        ClassID
	SerializedType *SerializedType
	Version        Version
	Platform       int
	FormatVersion  int
}

func NewObjectReader(reader *BinaryReader, assetFile *SerializedFile, objectInfo *ObjectInfo) *ObjectReader {
	return &ObjectReader{
		BinaryReader:   NewBinaryReaerFromReader(reader.reader, reader.bigEndian),
		AssetFile:      assetFile,
		PathID:         objectInfo.PathID,
		ByteStart:      objectInfo.ByteStart,
		ByteSize:       objectInfo.ByteSize,
		ClassID:        objectInfo.ClassID,
		SerializedType: objectInfo.SerializedType,
		Platform:       assetFile.TargetPlatform,
		Version:        assetFile.Version,
		FormatVersion:  int(assetFile.Header.Version),
	}
}

type ResourceReader struct {
	NeedSearch bool
	Path       string
	AssetFile  *SerializedFile
	Offset     int64
	Size       int64
	Reader     *BinaryReader
}

func NewResourceReaderWithAssetFile(path string, assetsFile *SerializedFile, offset int64, size int64) *ResourceReader {
	r := &ResourceReader{}
	r.NeedSearch = true
	r.Path = path
	r.AssetFile = assetsFile
	r.Offset = offset
	r.Size = size

	return r
}

func NewResourceReader(reader *BinaryReader, offset int64, size int64) *ResourceReader {
	r := &ResourceReader{}
	r.Reader = reader
	r.Offset = offset
	r.Size = size

	return r
}

func (r *ResourceReader) GetReader() *BinaryReader {
	if r.NeedSearch {
		resFileName := filepath.Base(r.Path)
		reader, ok := r.AssetFile.AssetsManager.ResourceFileReaders[resFileName]
		if ok {
			r.NeedSearch = false
			r.Reader = reader
			return reader
		}

		parent := filepath.Dir(r.AssetFile.OriginalPath)
		resFilePath := filepath.Join(parent, resFileName)
		_, err := os.Stat(resFilePath)
		if err != nil && os.IsNotExist(err) {
			files, err := filepath.Glob(filepath.Join(parent, "**", resFileName))
			if err != nil && len(files) > 0 {
				resFilePath = files[0]
			}
		}

		_, err = os.Stat(resFilePath)
		if err == nil {
			r.NeedSearch = false
			content, err := os.ReadFile(resFilePath)
			if err != nil {
				return nil
			}
			r.Reader = NewBinaryReaderFromBytes(content, true)
			r.AssetFile.AssetsManager.ResourceFileReaders[resFileName] = reader
			return reader
		}
		return nil
	} else {
		return r.Reader
	}
}

func (r *ResourceReader) GetData() []byte {
	reader := r.GetReader()
	if reader == nil {
		return nil
	}
	reader.SeekTo(r.Offset)
	return reader.Bytes(int(r.Size))
}

type IObject interface {
	GetObject() *Object
}

type Object struct {
	AssetFile       *SerializedFile
	Reader          *ObjectReader
	PathID          int64
	Version         Version
	Platform        int
	ClassID         ClassID
	SerializedType  *SerializedType
	ByteSize        uint32
	ObjectHideFlags uint32
}

func (o *Object) GetObject() *Object {
	return o
}

func NewObject(reader *ObjectReader) *Object {
	reader.SeekTo(int64(reader.ByteStart))

	o := &Object{
		AssetFile:      reader.AssetFile,
		Reader:         reader,
		ClassID:        reader.ClassID,
		PathID:         reader.PathID,
		Version:        reader.Version,
		Platform:       reader.Platform,
		SerializedType: reader.SerializedType,
		ByteSize:       reader.ByteSize,
	}

	if reader.Platform == -2 {
		o.ObjectHideFlags = reader.U32()
	}

	return o
}

type EditorExtension struct {
	*Object
	PrefabParentObject *PPtr
	PrefabInternal     *PPtr
}

func NewEditorExtension(reader *ObjectReader) *EditorExtension {
	e := &EditorExtension{
		Object: NewObject(reader),
	}

	if e.Platform == -2 {
		e.PrefabParentObject = NewPPtr(reader)
		e.PrefabInternal = NewPPtr(reader)
	}

	return e
}

type NamedObject struct {
	*EditorExtension
	Name string
}

func NewNamedObject(reader *ObjectReader) *NamedObject {
	n := &NamedObject{
		EditorExtension: NewEditorExtension(reader),
	}

	n.Name = reader.AlignedString()

	return n
}

type PPtr struct {
	FileID    int
	PathID    int64
	AssetFile *SerializedFile
	Index     int
}

func NewPPtr(reader *ObjectReader) *PPtr {
	result := &PPtr{}
	result.FileID = int(reader.S32())
	if reader.FormatVersion < 14 {
		result.PathID = int64(reader.S32())
	} else {
		result.PathID = reader.S64()
	}
	result.AssetFile = reader.AssetFile
	result.Index = -2
	return result
}

func (p *PPtr) GetAssetsFile() *SerializedFile {
	if p.FileID == 0 {
		return p.AssetFile
	}

	if p.FileID > 0 && p.FileID-1 < len(p.AssetFile.Externals) {
		manager := p.AssetFile.AssetsManager
		files := manager.AssetFiles
		fileIndexCache := manager.AssetFileIndexCache

		if p.Index == -2 {
			external := p.AssetFile.Externals[p.FileID-1]
			path := external.PathName

			found := false
			for pth := range fileIndexCache {
				if pth == path {
					found = true
					break
				}
			}

			if !found {
				p.Index = fileIndexCache[path]
			} else {
				for idx, file := range files {
					if strings.EqualFold(path, file.Path) {
						fileIndexCache[path] = idx
						p.Index = idx
						break
					}
				}
			}

			if p.Index >= 0 {
				return files[p.Index]
			}
		}
	}

	return nil
}

func (p *PPtr) Get() IObject {
	source := p.GetAssetsFile()
	if source == nil {
		return nil
	}

	if v, ok := source.ObjectMap[p.PathID]; ok {
		return v
	}

	return nil
}

func (p *PPtr) IsNull() bool {
	return p.PathID == 0 || p.FileID < 0
}

type StreamFile struct {
	Path   string
	Stream []byte
}

type FileType byte

const (
	FileTypeBundleFile FileType = iota
	FileTypeWebFile
	FileTypeGZipFile
	FileTypeBrotliFile
	FileTypeAssetsFile
	FileTypeZipFile
	FileTypeResourceFile
)

var (
	gzipMagic       []byte = []byte{0x1f, 0x8b}
	brotliMagic     []byte = []byte("brotli")
	zipMagic        []byte = []byte{'P', 'K', 0x03, 0x04}
	zipSpannedMagic []byte = []byte{'P', 'K', 0x07, 0x08}
)

type FileReader struct {
	*BinaryReader
	Path     string
	FileType FileType
}

func NewFileReader(stream []byte, path string) *FileReader {
	reader := &FileReader{
		BinaryReader: NewBinaryReaderFromBytes(stream, true),
	}
	reader.Path = path
	reader.FileType = reader.CheckFileType()
	return reader
}

func (r *FileReader) CheckFileType() FileType {
	signature := r.CharsWithMaxSize(16)
	r.SeekTo(0)

	if bytes.Equal(signature, []byte("UnityWeb")) || bytes.Equal(signature, []byte("UnityRaw")) || bytes.Equal(signature, []byte("UnityArchive")) || bytes.Equal(signature, []byte("UnityFS")) {
		return FileTypeBundleFile
	} else if bytes.Equal(signature, []byte("UnityWebData1.0")) {
		return FileTypeWebFile
	} else {
		magic := r.Bytes(2)
		if bytes.Equal(magic, gzipMagic) {
			return FileTypeGZipFile
		}

		err := r.SeekTo(0x20)
		if err == nil {
			magic = r.Bytes(6)
			if bytes.Equal(magic, brotliMagic) {
				return FileTypeBrotliFile
			}
		}

		r.SeekTo(0)
		if r.IsSerializedFile() {
			return FileTypeAssetsFile
		}

		r.SeekTo(0)
		magic = r.Bytes(4)
		if bytes.Equal(magic, zipMagic) || bytes.Equal(magic, zipSpannedMagic) {
			return FileTypeZipFile
		}

		return FileTypeResourceFile
	}
}

func (r *FileReader) IsSerializedFile() bool {
	length := r.Len()

	if length < 20 {
		return false
	}

	r.Skip(4)
	fileSize := uint64(r.U32())
	version := r.U32()
	dataOffset := uint64(r.U32())

	r.Skip(4)
	if version >= 22 {
		r.Skip(4)
		fileSize = r.U64()
		dataOffset = r.U64()
	}

	if fileSize != uint64(length) || dataOffset > fileSize {
		return false
	}

	return true
}
