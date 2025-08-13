package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const musicJacketPath = "./assets/star/forassetbundle/startapp/musicjacket/"

const (
	SongJSONPath = "./all.5.json"
	BandJSONPath = "./all.1.json"
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

type SongsData struct {
	Songs       All5JSON
	Bands       All1JSON
	SongNameMap map[string]int
}

func pick(names []string) string {
	name := names[preferLocale]
	for i := 0; name == "" && i < 5; i++ {
		name = names[i]
	}

	return name
}

func (s *SongsData) Title(id int, format string) string {
	info, ok := s.Songs[id]
	if !ok {
		return ""
	}

	band, ok := s.Bands[info.BandID]
	if !ok {
		return ""
	}

	title := pick(info.MusicTitle)
	artist := pick(band.BandName)

	return strings.ReplaceAll(strings.ReplaceAll(format, "%title", title), "%artist", artist)
}

func (s *SongsData) Jacket(id int) string {
	info, ok := s.Songs[id]
	if !ok {
		return ""
	}

	imgName := info.JacketImage[0]
	results, err := filepath.Glob(filepath.Join(musicJacketPath, fmt.Sprintf("musicjacket*/%s/", imgName)))
	if err != nil || len(results) == 0 {
		return ""
	}

	return results[0]
}

func fetchAll1Json() ([]byte, error) {
	resp, err := http.Get("https://bestdori.com/api/bands/all.1.json")
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, err
}

func fetchAll5Json() ([]byte, error) {
	resp, err := http.Get("https://bestdori.com/api/songs/all.5.json")
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func LoadSongsData() (*SongsData, error) {
	data, err := os.ReadFile(SongJSONPath)
	if err != nil {
		if os.IsNotExist(err) {
			data, err = fetchAll5Json()
			if err != nil {
				return nil, err
			}

			err = os.WriteFile(SongJSONPath, data, 0o644)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var songInfo All5JSON
	if err = json.Unmarshal(data, &songInfo); err != nil {
		return nil, err
	}

	data, err = os.ReadFile(BandJSONPath)
	if err != nil {
		if os.IsNotExist(err) {
			data, err = fetchAll1Json()
			if err != nil {
				return nil, err
			}

			err = os.WriteFile(BandJSONPath, data, 0o644)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
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

	return &SongsData{
		Songs:       songInfo,
		Bands:       bandInfo,
		SongNameMap: songNameMap,
	}, nil
}
