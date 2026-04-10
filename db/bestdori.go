// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package db

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kvarenzn/ssm/locale"
)

type SongInfo struct {
	Tag         string                      `json:"tag"`
	BandID      int                         `json:"bandId"`
	JacketImage []string                    `json:"jacketImage"`
	MusicTitle  []string                    `json:"musicTitle"`
	PublishedAt []string                    `json:"publishedAt"`
	ClosedAt    []string                    `json:"closedAt"`
	Difficulty  map[int]*SongDifficultyInfo `json:"difficulty"`
	MusicVideos map[string]*MusicVideoInfo  `json:"musicVideos"`
}

type SongDifficultyInfo struct {
	PlayLevel   int      `json:"playLevel"`
	PublishedAt []string `json:"publishedAt"`
}

type MusicVideoInfo struct {
	StartAt []string `json:"startAt"`
}

type All5JSON map[int]*SongInfo

type BandInfo struct {
	BandName []string `json:"bandName"`
}

type All1JSON map[int]*BandInfo

type BestdoriSongs struct {
	preferLocale int
	Songs        All5JSON
	Bands        All1JSON
	SongNameMap  map[string]int
}

func (s *BestdoriSongs) pick(names []string) string {
	name := names[s.preferLocale]
	for i := 0; name == "" && i < 5; i++ {
		name = names[i]
	}

	return name
}

func (s *BestdoriSongs) Title(id int, format string) string {
	info, ok := s.Songs[id]
	if !ok {
		return ""
	}

	band, ok := s.Bands[info.BandID]
	if !ok {
		return ""
	}

	title := s.pick(info.MusicTitle)
	artist := s.pick(band.BandName)

	return strings.ReplaceAll(strings.ReplaceAll(format, "${title}", title), "${artist}", artist)
}

func (s *BestdoriSongs) Jacket(id int) (string, string) {
	const bangMusicJacketPath = "./assets/star/forassetbundle/startapp/musicjacket/"

	info, ok := s.Songs[id]
	if !ok {
		return "", ""
	}

	imgName := info.JacketImage[0]
	results, err := filepath.Glob(filepath.Join(bangMusicJacketPath, fmt.Sprintf("musicjacket*/%s/", imgName)))
	if err != nil || len(results) == 0 {
		return "", ""
	}

	path := results[0]
	return filepath.Join(path, "thumb.png"), filepath.Join(path, "jacket.png")
}

func NewBestdoriDB() (MusicDatabase, error) {
	var preferLocale int
	switch locale.LanguageString[:2] {
	case "zh":
		if locale.LanguageString == "zh_TW" || locale.LanguageString == "zh-TW" {
			preferLocale = 2 // LANG: zh_TW
		} else {
			preferLocale = 3 // LANG: zh_CN
		}
	case "ko":
		preferLocale = 4 // LANG: ko_KR
	case "en":
		preferLocale = 1 // LANG: en_US
	default:
		preferLocale = 0 // LANG: ja_JP
	}

	data, err := loadFromFileOrUrlAndSave("./all.5.json", "https://bestdori.com/api/songs/all.5.json")
	if err != nil {
		return nil, err
	}

	var songInfo All5JSON
	if err = json.Unmarshal(data, &songInfo); err != nil {
		return nil, err
	}

	data, err = loadFromFileOrUrlAndSave("./all.1.json", "https://bestdori.com/api/bands/all.1.json")
	if err != nil {
		return nil, err
	}

	var bandInfo All1JSON
	if err = json.Unmarshal(data, &bandInfo); err != nil {
		return nil, err
	}

	songNameMap := map[string]int{}
	for song, data := range songInfo {
		for _, name := range data.MusicTitle {
			songNameMap[name] = song
		}
	}

	return &BestdoriSongs{
		preferLocale: preferLocale,
		Songs:        songInfo,
		Bands:        bandInfo,
		SongNameMap:  songNameMap,
	}, nil
}
