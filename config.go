package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type JudgementLineConfig struct {
	X1 int `json:"x1"`
	X2 int `json:"x2"`
	Y  int `json:"y"`
}

type DeviceConfig struct {
	Width  int                 `json:"width"`
	Height int                 `json:"height"`
	Line   JudgementLineConfig `json:"line"`
}

type Config struct {
	Devices map[string]DeviceConfig `json:"devices"`
}

var GlobalConfig Config

func (c *Config) AskForSerial(serial string) DeviceConfig {
	dc := DeviceConfig{}
	fmt.Printf("Please provide info for device [%s]\n", serial)
	for dc.Width <= 0 {
		fmt.Print("Device Width (an integer > 0): ")
		fmt.Scanf("%d", &dc.Width)
	}

	for dc.Height <= 0 {
		fmt.Print("Device Height (an integer > 0): ")
		fmt.Scanf("%d", &dc.Height)
	}

	fmt.Print("Line X1: ")
	fmt.Scanf("%d", &dc.Line.X1)

	fmt.Print("Line X2: ")
	fmt.Scanf("%d", &dc.Line.X2)

	fmt.Print("Line Y: ")
	fmt.Scanf("%d", &dc.Line.Y)

	if c.Devices == nil {
		c.Devices = map[string]DeviceConfig{}
	}

	c.Devices[serial] = dc
	return dc
}

func LoadConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if _, err := os.Create(path); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	json.Unmarshal(data, &GlobalConfig)
	return nil
}

func SaveConfig(path string) error {
	data, err := json.Marshal(GlobalConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o666)
}
