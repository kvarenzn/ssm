// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

import (
	"strconv"
	"strings"

	"github.com/kvarenzn/ssm/optional"
)

type LocalSerializedObjectIdentifier struct {
	LocalSerializedFileIndex int32
	LocalIdentifierInFile    int64
}

type FileIdentifier struct {
	GUID     [16]byte
	Type     optional.Optional[int32]
	PathName string
}

type SerializedFileHeader struct {
	MetadataSize  uint32
	FileSize      uint64
	Version       uint32
	DataOffset    uint64
	BigEndian     bool
	ReservedBytes []byte
}

type SerializedFile struct {
	AssetsManager   *AssetsManager
	Reader          *FileReader
	OriginalPath    string
	Path            string
	Header          *SerializedFileHeader
	BigEndian       bool
	UnityVersion    string
	Version         []int
	TargetPlatform  int
	EnableTypeTree  bool
	Types           []*SerializedType
	BigIDEnabled    bool
	ObjectInfos     []*ObjectInfo
	ScriptTypes     []*LocalSerializedObjectIdentifier
	Externals       []*FileIdentifier
	RefTypes        []*SerializedType
	UserInformation string
	Objects         []IObject
	ObjectMap       map[int64]IObject
	Parent          *BundleFile
}

func NewSerializedFile(reader *FileReader, assetsManager *AssetsManager, parent *BundleFile) (*SerializedFile, error) {
	file := &SerializedFile{
		AssetsManager: assetsManager,
		Reader:        reader,
		Path:          reader.Path,
		Parent:        parent,
		Header: &SerializedFileHeader{
			MetadataSize: reader.U32(),
			FileSize:     uint64(reader.U32()),
			Version:      reader.U32(),
			DataOffset:   uint64(reader.U32()),
		},
	}

	// header
	if file.Header.Version >= 9 {
		file.Header.BigEndian = reader.Bool()
		file.Header.ReservedBytes = reader.Bytes(3)
		file.BigEndian = file.Header.BigEndian
	} else {
		reader.SeekTo(int64(file.Header.FileSize) - int64(file.Header.MetadataSize))
		file.BigEndian = reader.Bool()
	}

	if file.Header.Version >= 22 {
		file.Header.MetadataSize = reader.U32()
		file.Header.FileSize = reader.U64()
		file.Header.DataOffset = reader.U64()
		reader.Skip(8)
	}

	// metadata
	if !file.BigEndian {
		reader.SetBigEndian(false)
	}

	if file.Header.Version >= 7 {
		file.UnityVersion = reader.CString()
		file.SetVersion(file.UnityVersion)
	}

	if file.Header.Version >= 8 {
		file.TargetPlatform = int(reader.S32())
	}

	file.EnableTypeTree = true
	if file.Header.Version >= 13 {
		file.EnableTypeTree = reader.Bool()
	}

	// types
	typesCount := reader.S32()
	file.Types = []*SerializedType{}
	for range int(typesCount) {
		file.Types = append(file.Types, file.readSerializedType(reader.BinaryReader, false))
	}

	file.BigIDEnabled = false
	if file.Header.Version >= 7 && file.Header.Version < 14 {
		val := reader.S32()
		if val != 0 {
			file.BigIDEnabled = true
		}
	}

	// objects
	objectsCount := int(reader.S32())
	file.ObjectInfos = []*ObjectInfo{}
	for range objectsCount {
		var pathID int64
		if file.BigIDEnabled {
			pathID = reader.S64()
		} else if file.Header.Version < 14 {
			pathID = int64(reader.S32())
		} else {
			reader.Align(4)
			pathID = reader.S64()
		}

		var byteStart int64
		if file.Header.Version >= 22 {
			byteStart = reader.S64()
		} else {
			byteStart = int64(reader.U32())
		}

		byteStart += int64(file.Header.DataOffset)
		byteSize := reader.U32()
		typeID := reader.S32()

		var classID ClassID
		var serializedType *SerializedType

		if file.Header.Version < 16 {
			classID = ClassID(reader.U16())
			for _, t := range file.Types {
				if t.ClassID == ClassID(typeID) {
					serializedType = t
					break
				}
			}
		} else {
			serializedType = file.Types[typeID]
			classID = serializedType.ClassID
		}

		isDestroyed := optional.None[uint16]()
		if file.Header.Version < 11 {
			isDestroyed = optional.Some(reader.U16())
		}

		if file.Header.Version >= 11 && file.Header.Version < 17 {
			scriptTypeIndex := reader.U16()
			if serializedType != nil {
				serializedType.ScriptTypeIndex = optional.Some(scriptTypeIndex)
			}
		}

		stripped := optional.None[bool]()
		if file.Header.Version == 15 || file.Header.Version == 16 {
			stripped = optional.Some(reader.Bool())
		}

		file.ObjectInfos = append(file.ObjectInfos, &ObjectInfo{
			ByteStart:      byteStart,
			ByteSize:       byteSize,
			TypeID:         int(typeID),
			ClassID:        classID,
			PathID:         pathID,
			IsDestroyed:    isDestroyed,
			Stripped:       stripped,
			SerializedType: serializedType,
		})
	}

	// scripts
	if file.Header.Version >= 11 {
		scriptsCount := reader.S32()

		for range int(scriptsCount) {
			localSerializedFileIndex := reader.S32()

			var localIdentifierInFile int64
			if file.Header.Version < 14 {
				localIdentifierInFile = int64(reader.S32())
			} else {
				reader.Align(4)
				localIdentifierInFile = reader.S64()
			}

			file.ScriptTypes = append(file.ScriptTypes, &LocalSerializedObjectIdentifier{
				LocalSerializedFileIndex: localSerializedFileIndex,
				LocalIdentifierInFile:    localIdentifierInFile,
			})
		}
	}

	// externals
	externalCount := reader.S32()
	for range int(externalCount) {
		if file.Header.Version >= 6 {
			reader.CString()
		}

		var guid [16]byte
		var typ optional.Optional[int32]
		if file.Header.Version >= 5 {
			guid = [16]byte(reader.Bytes(16))
			typ = optional.Some(reader.S32())
		}

		pathName := reader.CString()
		file.Externals = append(file.Externals, &FileIdentifier{
			GUID:     guid,
			Type:     typ,
			PathName: pathName,
		})
	}

	// ref types
	if file.Header.Version >= 20 {
		length := int(reader.S32())
		for range length {
			file.RefTypes = append(file.RefTypes, file.readSerializedType(reader.BinaryReader, true))
		}
	}

	if file.Header.Version >= 5 {
		file.UserInformation = reader.CString()
	}

	file.ObjectMap = map[int64]IObject{}

	return file, nil
}

func (f *SerializedFile) readSerializedType(reader *BinaryReader, isRefType bool) *SerializedType {
	t := &SerializedType{
		ClassID: ClassID(reader.S32()),
	}

	version := f.Header.Version
	if version >= 16 {
		t.IsStrippedType = optional.Some(reader.Bool())
	}

	if version >= 17 {
		t.ScriptTypeIndex = optional.Some(reader.U16())
	}

	if version >= 13 {
		if isRefType && t.ScriptTypeIndex.IsSome() {
			t.ScriptID = reader.Bytes(16)
		} else if version < 16 && t.ClassID < 0 || version >= 16 && t.ClassID == ClassIDMonoBehaviour {
			t.ScriptID = reader.Bytes(16)
		}

		t.OldTypeHash = reader.Bytes(16)
	}

	if f.EnableTypeTree {
		nodes := []*TypeTreeNode{}
		if version >= 12 || version == 10 {
			nodesCount := reader.S32()
			stringBufferSize := int(reader.S32())

			offsets := [][2]uint32{}
			for range nodesCount {
				node := &TypeTreeNode{}
				node.Version = int(reader.U16())
				node.Level = int(reader.U8())
				node.TypeFlags = int(reader.U8())
				typeStrOffset := reader.U32()
				nameStrOffset := reader.U32()
				node.ByteSize = int(reader.S32())
				node.Index = int(reader.S32())
				node.MetaFlag = optional.Some(int(reader.S32()))

				if version >= 19 {
					node.RefTypeHash = optional.Some(reader.U64())
				}

				offsets = append(offsets, [2]uint32{typeStrOffset, nameStrOffset})
				nodes = append(nodes, node)
			}

			stringBuffer := reader.Bytes(stringBufferSize)
			bufReader := NewBinaryReaderFromBytes(stringBuffer, true)

			readString := func(offset uint32) string {
				if offset&0x80000000 == 0 {
					bufReader.SeekTo(int64(offset))
					return bufReader.CString()
				}

				return COMMON_STRINGS[int(offset&0x7fffffff)]
			}

			for i, info := range offsets {
				nodes[i].Type = readString(info[0])
				nodes[i].Name = readString(info[1])
			}
		} else {
			var readTypeTree func(int)
			readTypeTree = func(level int) {
				typeString := reader.CString()
				name := reader.CString()
				byteSize := reader.S32()

				if version == 2 {
					reader.Skip(4)
				}

				var index int32
				if version == 3 {
					index = reader.S32()
				}

				typeFlags := reader.S32()
				ver := reader.S32()

				var metaFlag optional.Optional[int32]
				if version != 3 {
					metaFlag = optional.Some(reader.S32())
				}

				nodes = append(nodes, &TypeTreeNode{
					Level:       level,
					Type:        typeString,
					Name:        name,
					ByteSize:    int(byteSize),
					Index:       int(index),
					TypeFlags:   int(typeFlags),
					Version:     int(ver),
					MetaFlag:    optional.Some(int(metaFlag.Unwrap())),
					RefTypeHash: nil,
				})

				levels := int(reader.S32())
				for range levels {
					readTypeTree(level + 1)
				}
			}

			readTypeTree(0)
		}

		t.Type = &TypeTree{
			Nodes: nodes,
		}

		if version >= 21 {
			if isRefType {
				t.ClassName = reader.CString()
				t.Namespace = reader.CString()
				t.AsmName = reader.CString()
			} else {
				for range reader.U32() {
					t.TypeDependencies = append(t.TypeDependencies, reader.S32())
				}
			}
		}

	}
	return t
}

func (f *SerializedFile) AddObject(object IObject) {
	f.Objects = append(f.Objects, object)
	if o := object.GetObject(); o != nil {
		f.ObjectMap[o.PathID] = object
	} else {
		panic("???")
	}
}

func (f *SerializedFile) SetVersion(unityVersion string) {
	if unityVersion == "0.0.0" {
		f.Version = fallbackVersion
		return
	}
	f.Version = Version{}
	for str := range strings.SplitSeq(unityVersion, ".") {
		sp, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			f.Version = append(f.Version, int(sp))
		}
	}
}
