// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package db

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type MusicDatabase interface {
	Title(id int, format string) string
	Jacket(id int) (string, string)
}

func httpGetJson(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := json.Indent(&out, body, "", "    "); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func loadFromFileOrUrlAndSave(path string, url string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			data, err = httpGetJson(url)
			if err != nil {
				return nil, err
			}

			err = os.WriteFile(path, data, 0o644)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return data, nil
}
