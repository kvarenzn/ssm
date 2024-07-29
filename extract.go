package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kvarenzn/ssm/optional"
	"github.com/pierrec/lz4"
)

var COMMON_STRINGS map[int]string = map[int]string{
	0:    "AABB",
	5:    "AnimationClip",
	19:   "AnimationCurve",
	34:   "AnimationState",
	49:   "Array",
	55:   "Base",
	60:   "BitField",
	69:   "bitset",
	76:   "bool",
	81:   "char",
	86:   "ColorRGBA",
	96:   "Component",
	106:  "data",
	111:  "deque",
	117:  "double",
	124:  "dynamic_array",
	138:  "FastPropertyName",
	155:  "first",
	161:  "float",
	167:  "Font",
	172:  "GameObject",
	183:  "Generic Mono",
	196:  "GradientNEW",
	208:  "GUID",
	213:  "GUIStyle",
	222:  "int",
	226:  "list",
	231:  "long long",
	241:  "map",
	245:  "Matrix4x4f",
	256:  "MdFour",
	263:  "MonoBehaviour",
	277:  "MonoScript",
	288:  "m_ByteSize",
	299:  "m_Curve",
	307:  "m_EditorClassIdentifier",
	331:  "m_EditorHideFlags",
	349:  "m_Enabled",
	359:  "m_ExtensionPtr",
	374:  "m_GameObject",
	387:  "m_Index",
	395:  "m_IsArray",
	405:  "m_IsStatic",
	416:  "m_MetaFlag",
	427:  "m_Name",
	434:  "m_ObjectHideFlags",
	452:  "m_PrefabInternal",
	469:  "m_PrefabParentObject",
	490:  "m_Script",
	499:  "m_StaticEditorFlags",
	519:  "m_Type",
	526:  "m_Version",
	536:  "Object",
	543:  "pair",
	548:  "PPtr<Component>",
	564:  "PPtr<GameObject>",
	581:  "PPtr<Material>",
	596:  "PPtr<MonoBehaviour>",
	616:  "PPtr<MonoScript>",
	633:  "PPtr<Object>",
	646:  "PPtr<Prefab>",
	659:  "PPtr<Sprite>",
	672:  "PPtr<TextAsset>",
	688:  "PPtr<Texture>",
	702:  "PPtr<Texture2D>",
	718:  "PPtr<Transform>",
	734:  "Prefab",
	741:  "Quaternionf",
	753:  "Rectf",
	759:  "RectInt",
	767:  "RectOffset",
	778:  "second",
	785:  "set",
	789:  "short",
	795:  "size",
	800:  "SInt16",
	807:  "SInt32",
	814:  "SInt64",
	821:  "SInt8",
	827:  "staticvector",
	840:  "string",
	847:  "TextAsset",
	857:  "TextMesh",
	866:  "Texture",
	874:  "Texture2D",
	884:  "Transform",
	894:  "TypelessData",
	907:  "UInt16",
	914:  "UInt32",
	921:  "UInt64",
	928:  "UInt8",
	934:  "unsigned int",
	947:  "unsigned long long",
	966:  "unsigned short",
	981:  "vector",
	988:  "Vector2f",
	997:  "Vector3f",
	1006: "Vector4f",
	1015: "m_ScriptingClassIdentifier",
	1042: "Gradient",
	1051: "Type*",
	1057: "int2_storage",
	1070: "int3_storage",
	1083: "BoundsInt",
	1093: "m_CorrespondingSourceObject",
	1121: "m_PrefabInstance",
	1138: "m_PrefabAsset",
	1152: "FileSize",
	1161: "Hash128",
}

type TypeTreeNode struct {
	Type        string
	Name        string
	ByteSize    int
	Index       optional.Optional[int]
	TypeFlags   int
	Version     int
	MetaFlag    optional.Optional[int]
	Level       int
	RefTypeHash optional.Optional[uint64]
}

type TypeTree struct {
	Nodes        []TypeTreeNode
	StringBuffer []byte
}

type ClassID int

const (
	ClassIDUnknown                 ClassID = -1
	ClassIDObject                  ClassID = 0
	ClassIDGameObject              ClassID = 1
	ClassIDTransform               ClassID = 4
	ClassIDCamera                  ClassID = 20
	ClassIDMaterial                ClassID = 21
	ClassIDMeshRenderer            ClassID = 23
	ClassIDTexture2D               ClassID = 28
	ClassIDMeshFilter              ClassID = 33
	ClassIDMesh                    ClassID = 43
	ClassIDShader                  ClassID = 48
	ClassIDTextAsset               ClassID = 49
	ClassIDBoxCollider2D           ClassID = 61
	ClassIDBoxCollider             ClassID = 65
	ClassIDComputeShader           ClassID = 72
	ClassIDAnimationClip           ClassID = 74
	ClassIDAudioSource             ClassID = 82
	ClassIDAudioClip               ClassID = 83
	ClassIDRenderTexture           ClassID = 84
	ClassIDAvatar                  ClassID = 90
	ClassIDAnimatorController      ClassID = 91
	ClassIDAnimator                ClassID = 95
	ClassIDRenderSettings          ClassID = 104
	ClassIDLIGHT                   ClassID = 108
	ClassIDMonoBehaviour           ClassID = 114
	ClassIDMonoScript              ClassID = 115
	ClassIDFont                    ClassID = 128
	ClassIDSphereCollider          ClassID = 135
	ClassIDSkinnedMeshRenderer     ClassID = 137
	ClassIDAssetBundle             ClassID = 142
	ClassIDPreloadData             ClassID = 150
	ClassIDLightmapSettings        ClassID = 157
	ClassIDNavMeshSettings         ClassID = 196
	ClassIDParticleSystem          ClassID = 198
	ClassIDParticleSystemRenderer  ClassID = 199
	ClassIDShaderVariantCollection ClassID = 200
	ClassIDSpriteRenderer          ClassID = 212
	ClassIDSprite                  ClassID = 213
	ClassIDCanvasRenderer          ClassID = 222
	ClassIDCanvas                  ClassID = 223
	ClassIDRectTransform           ClassID = 224
	ClassIDCanvasGroup             ClassID = 225
	ClassIDPlayableDirector        ClassID = 320
)

type SerializedType struct {
	ClassID          ClassID
	IsStrippedType   optional.Optional[bool]
	ScriptTypeIndex  optional.Optional[uint16]
	TypeTree         optional.Optional[TypeTree]
	ScriptID         []byte
	OldTypeHash      []byte
	TypeDependencies []int
	ClassName        string
	Namespace        string
	AsmName          string
}

type ObjectInfo struct {
	ByteStart      int
	ByteSize       int
	TypeID         int
	ClassID        ClassID
	PathID         int64
	IsDestroyed    optional.Optional[uint16]
	Stripped       optional.Optional[bool]
	SerializedType optional.Optional[SerializedType]
}

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
	Header          SerializedFileHeader
	BigEndian       bool
	UnityVersion    string
	Version         []int
	TargetPlatform  int
	EnableTypeTree  bool
	Types           []SerializedType
	BigIDEnabled    bool
	ObjectInfos     []ObjectInfo
	ScriptTypes     []LocalSerializedObjectIdentifier
	Externals       []FileIdentifier
	RefTypes        []SerializedType
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
		Header: SerializedFileHeader{
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
	file.Types = []SerializedType{}
	for i := 0; i < int(typesCount); i++ {
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
	file.ObjectInfos = []ObjectInfo{}
	for i := 0; i < objectsCount; i++ {
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
		var serializedType optional.Optional[SerializedType]

		if file.Header.Version < 16 {
			classID = ClassID(reader.U16())
			for _, t := range file.Types {
				if t.ClassID == ClassID(typeID) {
					serializedType = optional.Some(t)
					break
				}
			}
		} else {
			serializedType := file.Types[typeID]
			classID = serializedType.ClassID
		}

		isDestroyed := optional.None[uint16]()
		if file.Header.Version < 11 {
			isDestroyed = optional.Some(reader.U16())
		}

		if file.Header.Version >= 11 && file.Header.Version < 17 {
			scriptTypeIndex := reader.U16()
			if serializedType.IsSome() {
				serializedType.UnwrapPtr().ScriptTypeIndex = optional.Some(scriptTypeIndex)
			}
		}

		stripped := optional.None[bool]()
		if file.Header.Version == 15 || file.Header.Version == 16 {
			stripped = optional.Some(reader.Bool())
		}

		file.ObjectInfos = append(file.ObjectInfos, ObjectInfo{
			ByteStart:      int(byteStart),
			ByteSize:       int(byteSize),
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

		for i := 0; i < int(scriptsCount); i++ {
			localSerializedFileIndex := reader.S32()

			var localIdentifierInFile int64
			if file.Header.Version < 14 {
				localIdentifierInFile = int64(reader.S32())
			} else {
				reader.Align(4)
				localIdentifierInFile = reader.S64()
			}

			file.ScriptTypes = append(file.ScriptTypes, LocalSerializedObjectIdentifier{
				LocalSerializedFileIndex: localSerializedFileIndex,
				LocalIdentifierInFile:    localIdentifierInFile,
			})
		}
	}

	// externals
	externalCount := reader.S32()
	for i := 0; i < int(externalCount); i++ {
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
		file.Externals = append(file.Externals, FileIdentifier{
			GUID:     guid,
			Type:     typ,
			PathName: pathName,
		})
	}

	// ref types
	if file.Header.Version >= 20 {
		length := reader.S32()
		for i := 0; i < int(length); i++ {
			file.RefTypes = append(file.RefTypes, file.readSerializedType(reader.BinaryReader, true))
		}
	}

	if file.Header.Version >= 5 {
		file.UserInformation = reader.CString()
	}

	file.ObjectMap = map[int64]IObject{}

	return file, nil
}

func (f *SerializedFile) readSerializedType(reader *BinaryReader, isRefType bool) SerializedType {
	t := SerializedType{
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
		if isRefType && t.ScriptTypeIndex.IsSome() && t.ScriptTypeIndex.Unwrap() >= 0 {
			t.ScriptID = reader.Bytes(16)
		} else if version < 16 && t.ClassID < 0 || version >= 16 && t.ClassID == ClassIDMonoBehaviour {
			t.ScriptID = reader.Bytes(16)
		}

		t.OldTypeHash = reader.Bytes(16)
	}

	if f.EnableTypeTree {
		nodes := []TypeTreeNode{}
		if version >= 12 || version == 10 {
			nodesCount := int(reader.S32())
			length := reader.S32()

			offsets := [][2]int{}
			for i := 0; i < nodesCount; i++ {
				ver := reader.U16()
				level := reader.U8()
				typeFlags := reader.U8()
				typeStrOffset := reader.U32()
				nameStrOffset := reader.U32()
				byteSize := reader.S32()
				index := reader.S32()
				metaFlag := reader.S32()

				var refTypeHash optional.Optional[uint64]
				if version >= 19 {
					refTypeHash = optional.Some(reader.U64())
				}

				offsets = append(offsets, [2]int{int(typeStrOffset), int(nameStrOffset)})
				nodes = append(nodes, TypeTreeNode{
					ByteSize:    int(byteSize),
					Index:       optional.Some(int(index)),
					TypeFlags:   int(typeFlags),
					Version:     int(ver),
					MetaFlag:    optional.Some(int(metaFlag)),
					Level:       int(level),
					RefTypeHash: refTypeHash,
				})
			}

			stringBuffer := reader.Bytes(int(length))

			bufReader := NewBinaryReaderFromBytes(stringBuffer, true)

			readString := func(offset int) string {
				if offset&0x80000000 == 0 {
					bufReader.SeekTo(int64(offset))
					return bufReader.CString()
				}

				offset = offset & 0x7fffffff
				return COMMON_STRINGS[offset]
			}

			for i := 0; i < nodesCount; i++ {
				info := offsets[i]
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

				var index optional.Optional[int32]
				if version == 3 {
					index = optional.Some(reader.S32())
				}

				typeFlags := reader.S32()
				ver := reader.S32()

				var metaFlag optional.Optional[int32]
				if version != 3 {
					metaFlag = optional.Some(reader.S32())
				}

				nodes = append(nodes, TypeTreeNode{
					Level:       level,
					Type:        typeString,
					Name:        name,
					ByteSize:    int(byteSize),
					Index:       optional.Some(int(index.Unwrap())),
					TypeFlags:   int(typeFlags),
					Version:     int(ver),
					MetaFlag:    optional.Some(int(metaFlag.Unwrap())),
					RefTypeHash: nil,
				})

				levels := int(reader.S32())
				for i := 0; i < levels; i++ {
					readTypeTree(level + 1)
				}
			}

			readTypeTree(0)
		}

		if version >= 21 {
			if isRefType {
				t.ClassName = reader.CString()
				t.Namespace = reader.CString()
				t.AsmName = reader.CString()
			} else {
				depsCount := int(reader.U32())
				for i := 0; i < depsCount; i++ {
					t.TypeDependencies = append(t.TypeDependencies, int(reader.S32()))
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
		f.Version = Version{2021, 6, 4}
		return
	}
	f.Version = Version{}
	for _, str := range strings.Split(unityVersion, ".") {
		sp, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			f.Version = append(f.Version, int(sp))
		}
	}
}

type ObjectReader struct {
	*BinaryReader
	AssetFile      *SerializedFile
	PathID         int64
	ByteStart      int
	ByteSize       int
	ClassID        ClassID
	SerializedType optional.Optional[SerializedType]
	Version        Version
	Platform       int
	FormatVersion  int
}

func NewObjectReader(reader *BinaryReader, assetFile *SerializedFile, objectInfo ObjectInfo) *ObjectReader {
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
			files, err := filepath.Glob(filepath.Join(parent, fmt.Sprintf("**/%s", resFileName)))
			if err != nil && len(files) > 0 {
				resFilePath = files[0]
			}
		}

		_, err = os.Stat(resFilePath)
		if err == nil {
			r.NeedSearch = false
			content, err := os.ReadFile(resFilePath)
			if err != nil {
				fmt.Println("read resource file failed")
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
	SerializedType  optional.Optional[SerializedType]
	ByteSize        int
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
					if strings.ToLower(path) == strings.ToLower(file.Path) {
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

	for id, v := range source.ObjectMap {
		if id == p.PathID {
			return v
		}
	}

	return nil
}

func (p *PPtr) IsNull() bool {
	return p.PathID == 0 || p.FileID < 0
}

type TextAsset struct {
	*NamedObject
	Content []byte
}

func NewTextAsset(reader *ObjectReader) *TextAsset {
	t := &TextAsset{
		NamedObject: NewNamedObject(reader),
	}

	t.Content = reader.Bytes(int(reader.S32()))

	return t
}

type Version []int

func (v Version) GreaterEqual(target Version) bool {
	if len(target) == 1 {
		target = append(target, 0, 0)
	} else if len(target) == 2 {
		target = append(target, 0)
	}

	return v[0] > target[0] || v[0] == target[0] && v[1] > target[1] || v[0] == target[0] && v[1] == target[1] && v[2] >= target[2]
}

func (v Version) LessEqual(target Version) bool {
	if len(target) == 1 {
		target = append(target, 0, 0)
	} else if len(target) == 2 {
		target = append(target, 0)
	}

	return v[0] < target[0] || v[0] == target[0] && v[1] < target[1] || v[0] == target[0] && v[1] == target[1] && v[2] <= target[2]
}

type Texture struct {
	*NamedObject
	ForcedFallbackFormat   optional.Optional[int32]
	DownscaleFallback      optional.Optional[bool]
	IsAlphaChannelOptional optional.Optional[bool]
}

func NewTexture(reader *ObjectReader) *Texture {
	t := &Texture{
		NamedObject: NewNamedObject(reader),
	}

	if t.Version.GreaterEqual(Version{2017, 3}) {
		t.ForcedFallbackFormat = optional.Some(reader.S32())
		t.DownscaleFallback = optional.Some(reader.Bool())
		if t.Version.GreaterEqual(Version{2020, 2}) {
			t.IsAlphaChannelOptional = optional.Some(reader.Bool())
		}
		reader.Align(4)
	}

	return t
}

type TextureFormat int

const (
	Alpha8 TextureFormat = iota + 1
	ARGB4444
	RGB24
	RGBA32
	ARGB32
	ARGBFloat
	RGB565
	BGR24
	R16
	DXT1
	DXT3
	DXT5
	RGBA4444
	BGRA32
	RHalf
	RGHalf
	RGBAHalf
	RFloat
	RGFloat
	RGBAFloat
	YUY2
	RGB9e5Float
	RGBFloat
	BC6H
	BC7
	BC4
	BC5
	DXT1Crunched
	DXT5Crunched
	PVRTC_RGB2
	PVRTC_RGBA2
	PVRTC_RGB4
	PVRTC_RGBA4
	ETC_RGB4
	ATC_RGB4
	ATC_RGBA8
	EAC_R TextureFormat = iota + 5
	EAC_R_SIGNED
	EAC_RG
	EAC_RG_SIGNED
	ETC2_RGB
	ETC2_RGBA1
	ETC2_RGBA8
	ASTC_RGB_4x4
	ASTC_RGB_5x5
	ASTC_RGB_6x6
	ASTC_RGB_8x8
	ASTC_RGB_10x10
	ASTC_RGB_12x12
	ASTC_RGBA_4x4
	ASTC_RGBA_5x5
	ASTC_RGBA_6x6
	ASTC_RGBA_8x8
	ASTC_RGBA_10x10
	ASTC_RGBA_12x12
	ETC_RGB4_3DS
	ETC_RGBA8_3DS
	RG16
	R8
	ETC_RGB4Crunched
	ETC2_RGBA8Crunched
	ASTC_HDR_4x4
	ASTC_HDR_5x5
	ASTC_HDR_6x6
	ASTC_HDR_8x8
	ASTC_HDR_10x10
	ASTC_HDR_12x12
	RG32
	RGB48
	RGBA64
)

func (f TextureFormat) String() string {
	switch f {
	case Alpha8:
		return "Alpha8"
	case ARGB4444:
		return "ARGB4444"
	case RGB24:
		return "RGB24"
	case RGBA32:
		return "RGBA32"
	case ARGB32:
		return "ARGB32"
	case ARGBFloat:
		return "ARGBFloat"
	case RGB565:
		return "RGB565"
	case BGR24:
		return "BGR24"
	case R16:
		return "R16"
	case DXT1:
		return "DXT1"
	case DXT3:
		return "DXT3"
	case DXT5:
		return "DXT5"
	case RGBA4444:
		return "RGBA4444"
	case BGRA32:
		return "BGRA32"
	case RHalf:
		return "RHalf"
	case RGHalf:
		return "RGHalf"
	case RGBAHalf:
		return "RGBAHalf"
	case RFloat:
		return "RFloat"
	case RGFloat:
		return "RGFloat"
	case RGBAFloat:
		return "RGBAFloat"
	case YUY2:
		return "YUY2"
	case RGB9e5Float:
		return "RGB9e5Float"
	case RGBFloat:
		return "RGBFloat"
	case BC6H:
		return "BC6H"
	case BC7:
		return "BC7"
	case BC4:
		return "BC4"
	case BC5:
		return "BC5"
	case DXT1Crunched:
		return "DXT1Crunched"
	case DXT5Crunched:
		return "DXT5Crunched"
	case PVRTC_RGB2:
		return "PVRTC_RGB2"
	case PVRTC_RGBA2:
		return "PVRTC_RGBA2"
	case PVRTC_RGB4:
		return "PVRTC_RGB4"
	case PVRTC_RGBA4:
		return "PVRTC_RGBA4"
	case ETC_RGB4:
		return "ETC_RGB4"
	case ATC_RGB4:
		return "ATC_RGB4"
	case ATC_RGBA8:
		return "ATC_RGBA8"
	case EAC_R:
		return "EAC_R"
	case EAC_R_SIGNED:
		return "EAC_R_SIGNED"
	case EAC_RG:
		return "EAC_RG"
	case EAC_RG_SIGNED:
		return "EAC_RG_SIGNED"
	case ETC2_RGB:
		return "ETC2_RGB"
	case ETC2_RGBA1:
		return "ETC2_RGBA1"
	case ETC2_RGBA8:
		return "ETC2_RGBA8"
	case ASTC_RGB_4x4:
		return "ASTC_RGB_4x4"
	case ASTC_RGB_5x5:
		return "ASTC_RGB_5x5"
	case ASTC_RGB_6x6:
		return "ASTC_RGB_6x6"
	case ASTC_RGB_8x8:
		return "ASTC_RGB_8x8"
	case ASTC_RGB_10x10:
		return "ASTC_RGB_10x10"
	case ASTC_RGB_12x12:
		return "ASTC_RGB_12x12"
	case ASTC_RGBA_4x4:
		return "ASTC_RGBA_4x4"
	case ASTC_RGBA_5x5:
		return "ASTC_RGBA_5x5"
	case ASTC_RGBA_6x6:
		return "ASTC_RGBA_6x6"
	case ASTC_RGBA_8x8:
		return "ASTC_RGBA_8x8"
	case ASTC_RGBA_10x10:
		return "ASTC_RGBA_10x10"
	case ASTC_RGBA_12x12:
		return "ASTC_RGBA_12x12"
	case ETC_RGB4_3DS:
		return "ETC_RGB4_3DS"
	case ETC_RGBA8_3DS:
		return "ETC_RGBA8_3DS"
	case RG16:
		return "RG16"
	case R8:
		return "R8"
	case ETC_RGB4Crunched:
		return "ETC_RGB4Crunched"
	case ETC2_RGBA8Crunched:
		return "ETC2_RGBA8Crunched"
	case ASTC_HDR_4x4:
		return "ASTC_HDR_4x4"
	case ASTC_HDR_5x5:
		return "ASTC_HDR_5x5"
	case ASTC_HDR_6x6:
		return "ASTC_HDR_6x6"
	case ASTC_HDR_8x8:
		return "ASTC_HDR_8x8"
	case ASTC_HDR_10x10:
		return "ASTC_HDR_10x10"
	case ASTC_HDR_12x12:
		return "ASTC_HDR_12x12"
	case RG32:
		return "RG32"
	case RGB48:
		return "RGB48"
	case RGBA64:
		return "RGBA64"
	}

	return fmt.Sprintf("Unknown Format (%d)", f)
}

type TextureSettings struct {
	FilterMode int32
	Aniso      int32
	MipBias    float32
	WrapMode   int32
}

func NewTextureSettings(reader *ObjectReader) *TextureSettings {
	s := &TextureSettings{}

	s.FilterMode = reader.S32()
	s.Aniso = reader.S32()
	s.MipBias = reader.F32()

	s.WrapMode = reader.S32()
	if reader.Version.GreaterEqual(Version{2017}) {
		reader.S32() // WrapV
		reader.S32() // WrapW
	}

	return s
}

type StreamingInfo struct {
	Offset int64
	Size   uint32
	Path   string
}

func NewStreamingInfo(reader *ObjectReader) *StreamingInfo {
	i := &StreamingInfo{}
	if reader.Version.GreaterEqual(Version{2020, 1}) {
		i.Offset = reader.S64()
	} else {
		i.Offset = int64(reader.U32())
	}

	i.Size = reader.U32()
	i.Path = reader.AlignedString()

	return i
}

type Texture2D struct {
	*Texture
	Width           int32
	Height          int32
	Format          TextureFormat
	Mipmap          bool
	MipCount        int32
	TextureSettings *TextureSettings
	ImageData       *ResourceReader
	Info            *StreamingInfo
}

func NewTexture2D(reader *ObjectReader) *Texture2D {
	t := &Texture2D{
		Texture: NewTexture(reader),
	}
	t.Width = reader.S32()
	t.Height = reader.S32()

	reader.S32() // CompleteImageSize
	if t.Version.GreaterEqual(Version{2020, 1}) {
		reader.S32() // MipsStripped
	}

	t.Format = TextureFormat(reader.S32())
	if t.Version.GreaterEqual(Version{5, 2}) {
		t.MipCount = reader.S32()
	} else {
		t.Mipmap = reader.Bool()
	}

	if t.Version.GreaterEqual(Version{2, 6}) {
		reader.Bool() // IsReadable
	}

	if t.Version.GreaterEqual(Version{2020, 1}) {
		reader.Bool() // IsPreProcessed
	}

	if t.Version.GreaterEqual(Version{2019, 3}) {
		reader.Bool() // IgnoreMasterTextureLimit
	}

	if t.Version.GreaterEqual(Version{3}) && t.Version.LessEqual(Version{5, 4, 999}) {
		reader.Bool() // ReadAllowed
	}

	if t.Version.GreaterEqual(Version{2018, 2}) {
		reader.Bool() // StreamingMipmaps
	}

	reader.Align(4)
	if t.Version.GreaterEqual(Version{2018, 2}) {
		reader.S32() // StreamingMipmapsPriority
	}

	reader.S32() // ImageCount
	reader.S32() // TextureDimension

	t.TextureSettings = NewTextureSettings(reader)

	if t.Version.GreaterEqual(Version{3}) {
		reader.S32() // LightmapFormat
	}

	if t.Version.GreaterEqual(Version{3, 5}) {
		reader.S32() // ColorSpace
	}

	if t.Version.GreaterEqual(Version{2020, 2}) {
		reader.Bytes(int(reader.S32())) // PlatformBlob
		reader.Align(4)
	}

	imageDataSize := reader.S32()
	if imageDataSize == 0 && t.Version.GreaterEqual(Version{5, 3}) {
		t.Info = NewStreamingInfo(reader)
	}

	if t.Info == nil || t.Info.Path == "" {
		t.ImageData = NewResourceReader(reader.BinaryReader, reader.Position(), int64(imageDataSize))
	} else {
		t.ImageData = NewResourceReaderWithAssetFile(t.Info.Path, t.AssetFile, t.Info.Offset, int64(t.Info.Size))
	}

	return t
}

type Vector2f struct {
	X float32
	Y float32
}

type Vector3f struct {
	X float32
	Y float32
	Z float32
}

type Vector4f struct {
	X float32
	Y float32
	Z float32
	W float32
}

type Rectf struct {
	X      float32
	Y      float32
	Width  float32
	Height float32
}

func NewRectf(reader *BinaryReader) *Rectf {
	r := &Rectf{}
	r.X = reader.F32()
	r.Y = reader.F32()
	r.Width = reader.F32()
	r.Height = reader.F32()
	return r
}

type Sprite struct {
	Rect          *Rectf
	Offset        *Vector2f
	Border        *Vector4f
	PixelsToUnits float32
	Pivot         *Vector2f
	Extrude       uint32
	IsPolygon     bool
	AtlasTags     []string
	SpriteAtlas   PPtr
	PhysicsShape  [][]Vector2f
}

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
	Container    map[string]*AssetInfo
}

func NewAssetBundle(reader *ObjectReader) *AssetBundle {
	b := &AssetBundle{
		NamedObject: NewNamedObject(reader),
	}

	length := reader.S32()
	for i := 0; i < int(length); i++ {
		b.PreloadTable = append(b.PreloadTable, NewPPtr(reader))
	}

	b.Container = map[string]*AssetInfo{}
	length = reader.S32()
	for i := 0; i < int(length); i++ {
		key := reader.AlignedString()
		b.Container[key] = NewAssetInfo(reader)
	}

	return b
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
		for i := 0; i < matchLength; i++ {
			result[ptr+i] = result[begin+i]
		}
	}

	return result[:uncompressedSize]
}

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
			return nil, fmt.Errorf("File corrupted")
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
				fmt.Println(err)
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
		for i := 0; i < blocksCount; i++ {
			file.BlocksInfo = append(file.BlocksInfo, BundleFileStorageBlock{
				UncompressedSize: ucReader.U32(),
				CompressedSize:   ucReader.U32(),
				Flags:            ucReader.U16(),
			})
		}

		directoryCount := int(ucReader.S32())
		for i := 0; i < directoryCount; i++ {
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

		sReader := NewBinaryReaderFromBytes(blockStream.Bytes(), true)
		file.Files = []StreamFile{}
		for _, node := range file.DirectoryInfo {
			sReader.SeekTo(int64(node.Offset))
			file.Files = append(file.Files, StreamFile{
				Path:   node.Path,
				Stream: sReader.Bytes(node.Size),
			})
		}
	}

	return file, nil
}

type AssetsManager struct {
	AssetFiles          []*SerializedFile
	AssetFileHashes     []string
	ResourceFileReaders map[string]*BinaryReader
	AssetFileIndexCache map[string]int
}

func NewAssetsManager() AssetsManager {
	return AssetsManager{
		AssetFiles:          []*SerializedFile{},
		AssetFileHashes:     []string{},
		ResourceFileReaders: map[string]*BinaryReader{},
		AssetFileIndexCache: map[string]int{},
	}
}

func (m *AssetsManager) LoadFileFromHandler(file *os.File) error {
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	reader := NewFileReader(data, file.Name())
	if reader.FileType == FileTypeBundleFile {
		m.LoadBundle(reader, "")
	}

	return nil
}

func (m *AssetsManager) LoadBundle(reader *FileReader, originalPath string) error {
	bundleFile, err := NewBundleFile(reader)
	if err != nil {
		return err
	}

	for _, file := range bundleFile.Files {
		subreader := NewFileReader(file.Stream, reader.Path)
		subreader.SeekTo(0)
		if subreader.FileType == FileTypeAssetsFile {
			if originalPath == "" {
				err := m.LoadAssets(subreader, subreader.Path, bundleFile.Header.UnityVersion, bundleFile)
				if err != nil {
					return err
				}
			} else {
				err := m.LoadAssets(subreader, originalPath, bundleFile.Header.UnityVersion, bundleFile)
				if err != nil {
					return err
				}
			}
		} else {
			m.ResourceFileReaders[filepath.Base(file.Path)] = subreader.BinaryReader
		}
	}

	return nil
}

func (m *AssetsManager) LoadAssets(reader *FileReader, originalPath, unityVersion string, bundleFile *BundleFile) error {
	name := path.Base(reader.Path)
	found := false
	for _, hash := range m.AssetFileHashes {
		if hash == name {
			found = true
			break
		}
	}

	if found {
		return nil
	}

	assetFile, err := NewSerializedFile(reader, m, bundleFile)
	if err != nil {
		return err
	}

	assetFile.OriginalPath = originalPath
	if len(unityVersion) <= 0 && assetFile.Header.Version < 7 {
		assetFile.SetVersion(unityVersion)
	}

	m.AssetFiles = append(m.AssetFiles, assetFile)
	m.AssetFileHashes = append(m.AssetFileHashes, filepath.Base(assetFile.Path))

	return nil
}

func Extract(baseDir string) error {
	manager := NewAssetsManager()
	bundles, err := filepath.Glob(filepath.Join(baseDir, strings.Repeat("?", 64)))
	if err != nil {
		return err
	}

	formats := map[TextureFormat]bool{}
	for i, bundle := range bundles {
		fmt.Printf("[%d/%d] %s\n", i, len(bundles), bundle)
		file, err := os.Open(bundle)
		if err != nil {
			return nil
		}

		err = manager.LoadFileFromHandler(file)
		file.Close()
		if err != nil {
			fmt.Println("file corrupted, skip")
			continue
		}

		for _, file := range manager.AssetFiles {
			for _, info := range file.ObjectInfos {
				reader := NewObjectReader(file.Reader.BinaryReader, file, info)
				var obj IObject
				switch reader.ClassID {
				case ClassIDTextAsset:
					obj = NewTextAsset(reader)
				case ClassIDAssetBundle:
					obj = NewAssetBundle(reader)
				case ClassIDTexture2D:
					obj = NewTexture2D(reader)
				}
				if obj != nil {
					file.AddObject(obj)
				}
			}
		}

		for _, file := range manager.AssetFiles {
			for _, obj := range file.Objects {
				switch o := obj.(type) {
				case *AssetBundle:
					for key, info := range o.Container {
						item := info.Asset.Get()
						switch it := item.(type) {
						case *TextAsset:
							key = "./" + key
							stat, err := os.Stat(key)
							if err == nil && !stat.IsDir() {
								continue
							}
							parent := filepath.Dir(key)
							pstat, err := os.Stat(parent)
							if err != nil && os.IsNotExist(err) || pstat == nil || !pstat.IsDir() {
								err = os.MkdirAll(parent, 0o755|os.ModeDir)
								if err != nil {
									fmt.Println(err)
									return err
								}
							}
							os.WriteFile(key, it.Content, 0o644)
						case *Texture2D:
							formats[it.Format] = true
							key = "./" + key
							stat, err := os.Stat(key)
							if err == nil && !stat.IsDir() {
								continue
							}
							parent := filepath.Dir(key)
							pstat, err := os.Stat(parent)
							if err != nil && os.IsNotExist(err) || pstat == nil || !pstat.IsDir() {
								err = os.MkdirAll(parent, 0o755|os.ModeDir)
								if err != nil {
									fmt.Println(err)
									return err
								}
							}
							image, err := DecodeTexture2D(it)
							if err != nil {
								fmt.Println(err)
								continue
							}
							f, err := os.Create(key)
							if err != nil {
								fmt.Println(err)
								continue
							}
							fmt.Println(it.Format, key)
							png.Encode(f, image)
						}
					}
				}
			}
		}

		manager.AssetFiles = nil
		manager.AssetFileHashes = nil
	}

	fmt.Println(formats)

	return nil
}
