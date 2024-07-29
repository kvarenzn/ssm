package main

import (
	"encoding/json"
	"os"
)

type SongInfo struct {
	Tag         string                     `json:"tag"`
	BandID      int                        `json:"bandId"`
	JacketImage []string                   `json:"jacketImage"`
	MusicTitle  [5]string                  `json:"musicTitle"`
	PublishedAt [5]string                  `json:"publishedAt"`
	ClosedAt    [5]string                  `json:"closedAt"`
	Difficulty  map[int]SongDifficultyInfo `json:"difficulty"`
	MusicVideos map[string]MusicVideoInfo  `json:"musicVideos"`
}

type SongDifficultyInfo struct {
	PlayLevel   int       `json:"playLevel"`
	PublishedAt [5]string `json:"publishedAt"`
}

type MusicVideoInfo struct {
	StartAt [5]string `json:"startAt"`
}

type AllSong5Json map[int]SongInfo

type BandInfo struct {
	BandName [5]string `json:"bandName"`
}

type AllBand1Json map[int]BandInfo

type SongInfoData struct {
	SongInfos   map[int]SongInfo
	BandInfos   map[int]BandInfo
	SongNameMap map[string]int
}

func Load(songJsonPath string, bandJsonPath string) (*SongInfoData, error) {
	data, err := os.ReadFile(songJsonPath)
	if err != nil {
		return nil, err
	}

	var songInfo AllSong5Json
	if err = json.Unmarshal(data, &songInfo); err != nil {
		return nil, err
	}

	var bandInfo AllBand1Json
	if err = json.Unmarshal(data, &bandInfo); err != nil {
		return nil, err
	}

	songNameMap := map[string]int{}
	for song, data := range songInfo {
		for _, name := range data.MusicTitle {
			songNameMap[name] = song
		}
	}

	return &SongInfoData{
		SongInfos:   songInfo,
		BandInfos:   bandInfo,
		SongNameMap: songNameMap,
	}, nil
}
