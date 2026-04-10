// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type DeviceConfig struct {
	Serial string `json:"-"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Config struct {
	Path    string                   `json:"-"`
	Devices map[string]*DeviceConfig `json:"devices"`
}

func (c *Config) askFor(serial string) *DeviceConfig {
	dc := &DeviceConfig{}
	fmt.Printf("Please provide info for device [%s]\n", serial)
	for dc.Width <= 0 {
		fmt.Print("Device Width (an integer > 0): ")
		fmt.Scanln(&dc.Width)
	}

	for dc.Height <= 0 {
		fmt.Print("Device Height (an integer > 0): ")
		fmt.Scanln(&dc.Height)
	}

	dc.Serial = serial
	return dc
}

func (c *Config) Get(serial string) *DeviceConfig {
	if c.Devices == nil {
		c.Devices = map[string]*DeviceConfig{}
	}

	if dc, ok := c.Devices[serial]; ok {
		dc.Serial = serial
		return dc
	} else {
		dc = c.askFor(serial)
		c.Devices[serial] = dc
		c.Save()
		return dc
	}
}

func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if _, err := os.Create(path); err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	json.Unmarshal(data, &c)
	c.Path = path
	return c, nil
}

func (c *Config) Save() error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(c.Path, data, 0o666)
}
