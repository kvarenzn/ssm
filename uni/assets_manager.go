// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
)

type AssetsManager struct {
	AssetFiles          []*SerializedFile
	AssetFileHashes     []string
	ResourceFileReaders map[string]*BinaryReader
	AssetFileIndexCache map[string]int
}

func NewAssetsManager() *AssetsManager {
	return &AssetsManager{
		AssetFiles:          []*SerializedFile{},
		AssetFileHashes:     []string{},
		ResourceFileReaders: map[string]*BinaryReader{},
		AssetFileIndexCache: map[string]int{},
	}
}

func (m *AssetsManager) ClearCache() {
	m.AssetFiles = nil
	m.AssetFileHashes = nil
	m.ResourceFileReaders = map[string]*BinaryReader{}
	m.AssetFileIndexCache = map[string]int{}
}

func (m *AssetsManager) LoadFileFromHandler(file *os.File) error {
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return m.LoadDataFromHandler(data, file.Name())
}

func (m *AssetsManager) LoadDataFromHandler(data []byte, path string) error {
	reader := NewFileReader(data, path)
	if reader.FileType == FileTypeBundleFile {
		if err := m.LoadBundle(reader, ""); err != nil {
			return fmt.Errorf("Failed to load bundle: %w", err)
		}
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
	if slices.Contains(m.AssetFileHashes, name) {
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
