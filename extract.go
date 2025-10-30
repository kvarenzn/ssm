// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kvarenzn/ssm/log"
	"github.com/kvarenzn/ssm/term"
	"github.com/kvarenzn/ssm/uni"
)

type AssetFileMeta struct {
	Hash      string               `json:"hash"`
	Corrupted bool                 `json:"corrupted"`
	Files     []*FileExtractResult `json:"files"`
}

type FileExtractResult struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

type AssetFilesDatabase map[string]*AssetFileMeta

func checkPathAndCreateParentDirectory(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		if !stat.IsDir() {
			log.Debugf("`%s` already exists, skip", path)
			return true
		} else {
			log.Dief("`%s` is a directory. This is weird", path)
			return false
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		log.Dief("Cannot stat `%s`: %s", path, err)
		return false
	}

	parent := filepath.Dir(path)
	pstat, err := os.Stat(parent)
	if errors.Is(err, fs.ErrNotExist) {
		log.Debugf("Parent folder `%s` not found", parent)
		if err = os.MkdirAll(parent, 0o755); err != nil {
			log.Dief("Failed to create parent folder `%s`: %s", parent, err)
		}
	} else if pstat == nil {
		log.Dief("Cannot stat `%s`: %s", parent, err)
	} else if !pstat.IsDir() {
		log.Dief("`%s` is not a directory. This is weird", parent)
	}

	return false
}

func Extract(baseDir string, pathFilter func(string) bool) (AssetFilesDatabase, error) {
	if pathFilter == nil {
		pathFilter = func(s string) bool {
			return true
		}
	}
	db := AssetFilesDatabase{}
	manager := uni.NewAssetsManager()
	bundles, err := filepath.Glob(filepath.Join(baseDir, strings.Repeat("?", 64)))
	if err != nil {
		return nil, err
	}

	for i, bundle := range bundles {
		filename := filepath.Base(bundle)
		log.Infof("[%d/%d] %s", i, len(bundles), bundle)
		term.MoveUpAndReset(1)

		file, err := os.Open(bundle)
		if err != nil {
			return nil, err
		}

		data, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		meta := &AssetFileMeta{
			Hash: fmt.Sprintf("%x", md5.Sum(data)),
		}

		err = manager.LoadDataFromHandler(data, file.Name())
		file.Close()
		if err != nil {
			meta.Corrupted = true
			db[filename] = meta
			continue
		}

		for _, file := range manager.AssetFiles {
			for _, info := range file.ObjectInfos {
				reader := uni.NewObjectReader(file.Reader.BinaryReader, file, info)
				var obj uni.IObject
				switch reader.ClassID {
				case uni.ClassIDTextAsset:
					obj = uni.NewTextAsset(reader)
				case uni.ClassIDAssetBundle:
					obj = uni.NewAssetBundle(reader)
				case uni.ClassIDTexture2D:
					obj = uni.NewTexture2D(reader)
				}

				if obj != nil {
					file.AddObject(obj)
				}
			}
		}

		files := []*FileExtractResult{}
		for _, file := range manager.AssetFiles {
			for _, obj := range file.Objects {
				switch o := obj.(type) {
				case *uni.AssetBundle:
					for _, pair := range o.Container {
						if !pathFilter(pair.Key) {
							continue
						}

						item := pair.Value.Asset.Get()
						switch it := item.(type) {
						case *uni.TextAsset:
							key := filepath.Join(".", pair.Key)
							if checkPathAndCreateParentDirectory(key) {
								continue
							}

							result := &FileExtractResult{Name: key}
							files = append(files, result)

							if err = os.WriteFile(key, it.Content, 0o644); err != nil {
								result.Error = fmt.Sprintf("Write failed: %s", err)
							}
						case *uni.Texture2D:
							key := filepath.Join(".", pair.Key)
							if checkPathAndCreateParentDirectory(key) {
								continue
							}

							result := &FileExtractResult{Name: key}
							files = append(files, result)

							image, err := DecodeTexture2D(it)
							if err != nil {
								result.Error = fmt.Sprintf("Decode failed: %s", err)
								continue
							}

							f, err := os.Create(key)
							if err != nil {
								result.Error = fmt.Sprintf("Create failed %s", err)
								continue
							}

							if err := png.Encode(f, image); err != nil {
								f.Close()
								result.Error = fmt.Sprintf("Encode failed %s", err)
							}
							f.Close()
						}
					}
				}
			}
		}

		manager.ClearCache()
		meta.Files = files
		db[filename] = meta
	}

	return db, nil
}
