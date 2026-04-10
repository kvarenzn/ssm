// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package db

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type sekaiMusicInfo struct {
	ID                                int      `json:"id"`
	Seq                               int      `json:"seq"`
	ReleaseConditionID                int      `json:"releaseConditionId"`
	Categories                        []string `json:"categories"`
	Title                             string   `json:"title"`
	Pronunciation                     string   `json:"pronunciation"`
	CreatorArtistID                   int      `json:"creatorArtistId"`
	Lyricist                          string   `json:"lyricist"`
	Composer                          string   `json:"composer"`
	Arranger                          string   `json:"arranger"`
	DancerCount                       int      `json:"dancerCount"`
	SelfDancerPosition                int      `json:"selfDancerPosition"`
	AssetBundleName                   string   `json:"assetbundleName"`
	LiveTalkBackgroundAssetBundleName string   `json:"liveTalkBackgroundAssetbundleName"`
	PublishedAt                       uint64   `json:"publishedAt"`
	ReleasedAt                        uint64   `json:"releasedAt"`
	LiveStageID                       int      `json:"liveStageId"`
	FillerSec                         float64  `json:"fillerSec"`
	MusicCollaborationID              int      `json:"musicCollaborationId"`
	IsNewlyWrittenMusic               bool     `json:"isNewlyWrittenMusic"`
	IsFullLength                      bool     `json:"isFullLength"`
}

type sekaiMusics struct {
	Info []*sekaiMusicInfo
	Map  map[int]*sekaiMusicInfo
}

func (s *sekaiMusics) Title(id int, format string) string {
	info, ok := s.Map[id]
	if !ok {
		return ""
	}

	format = strings.ReplaceAll(format, "${title}", info.Title)
	format = strings.ReplaceAll(format, "${artist}", fmt.Sprintf("l. %s / c. %s / a. %s", info.Lyricist, info.Composer, info.Arranger))
	return format
}

func (s *sekaiMusics) Jacket(id int) (string, string) {
	const base = "./assets/sekai/assetbundle/resources/startapp"

	// look for thumbnail
	thumbnail := filepath.Join(base, "thumbnail", "music_jacket", fmt.Sprintf("jacket_s_%03d.png", id))

	jackets, err := filepath.Glob(filepath.Join(base, "music", "jacket", fmt.Sprintf("jacket_s_%03d", id), "*.png"))
	if err != nil || len(jackets) == 0 {
		return "", ""
	}

	return thumbnail, jackets[0]
}

func NewSekaiDB() (MusicDatabase, error) {
	data, err := loadFromFileOrUrlAndSave("./musics.json", "https://sekai-world.github.io/sekai-master-db-diff/musics.json")
	if err != nil {
		return nil, err
	}

	var info []*sekaiMusicInfo
	if err = json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	result := &sekaiMusics{
		Info: info,
		Map:  map[int]*sekaiMusicInfo{},
	}

	for _, i := range info {
		result.Map[i.ID] = i
	}

	return result, nil
}
