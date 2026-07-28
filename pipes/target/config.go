package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type PipeConfig struct {
	Pipes       int   `json:"pipes"`
	Fps         int   `json:"fps"`
	Steady      int   `json:"steady"`
	Limit       int   `json:"limit"`
	RandomStart bool  `json:"random_start"`
	Bold        bool  `json:"bold"`
	Color       bool  `json:"color"`
	KeepStyle   bool  `json:"keep_style"`
	Colors      []int `json:"colors"`
	PipeTypes   []int `json:"pipe_types"`
}

func DefaultConfig() PipeConfig {
	return PipeConfig{
		Pipes:       1,
		Fps:         75,
		Steady:      13,
		Limit:       2000,
		RandomStart: false,
		Bold:        true,
		Color:       true,
		KeepStyle:   false,
		Colors:      []int{1, 2, 3, 4, 5, 6, 7, 0},
		PipeTypes:   []int{0},
	}
}

func getConfigDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "pipes-py")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pipes-py")
}

func LoadConfig() PipeConfig {
	cfgFile := filepath.Join(getConfigDir(), "config.json")
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return DefaultConfig()
	}

	config := DefaultConfig()
	_ = json.Unmarshal(data, &config)
	return config
}

func SaveConfig(config PipeConfig) {
	dir := getConfigDir()
	os.MkdirAll(dir, 0755)
	cfgFile := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err == nil {
		os.WriteFile(cfgFile, data, 0644)
	}
}
