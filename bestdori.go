package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type SongInfo struct {
	Tag         string                      `json:"tag"`
	BandID      int                         `json:"bandId"`
	JacketImage []string                    `json:"jacketImage"`
	MusicTitle  [5]string                   `json:"musicTitle"`
	PublishedAt [5]string                   `json:"publishedAt"`
	ClosedAt    [5]string                   `json:"closedAt"`
	Difficulty  map[int]*SongDifficultyInfo `json:"difficulty"`
	MusicVideos map[string]*MusicVideoInfo  `json:"musicVideos"`
}

type SongDifficultyInfo struct {
	PlayLevel   int       `json:"playLevel"`
	PublishedAt [5]string `json:"publishedAt"`
}

type MusicVideoInfo struct {
	StartAt [5]string `json:"startAt"`
}

type All5Json map[int]*SongInfo

type BandInfo struct {
	BandName [5]string `json:"bandName"`
}

type All1Json map[int]*BandInfo

type SongsData struct {
	Songs       All5Json
	Bands       All1Json
	SongNameMap map[string]int
}

func FetchAll1Json() (All1Json, error) {
	var result All1Json
	resp, err := http.Get("https://bestdori.com/api/bands/all.1.json")
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func FetchAll5Json() (All5Json, error) {
	var result All5Json
	resp, err := http.Get("https://bestdori.com/api/songs/all.5.json")
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func LoadSongsData(songJsonPath string, bandJsonPath string) (*SongsData, error) {
	data, err := os.ReadFile(songJsonPath)
	if err != nil {
		return nil, err
	}

	var songInfo All5Json
	if err = json.Unmarshal(data, &songInfo); err != nil {
		return nil, err
	}

	var bandInfo All1Json
	if err = json.Unmarshal(data, &bandInfo); err != nil {
		return nil, err
	}

	songNameMap := map[string]int{}
	for song, data := range songInfo {
		for _, name := range data.MusicTitle {
			songNameMap[name] = song
		}
	}

	return &SongsData{
		Songs:       songInfo,
		Bands:       bandInfo,
		SongNameMap: songNameMap,
	}, nil
}
