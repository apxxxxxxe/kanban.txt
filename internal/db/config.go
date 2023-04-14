package db

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/apxxxxxxe/kanban.txt/pkg/util"
)

type Config struct {
	Color *ColorConfig `json:"color"`
}

type ColorConfig struct {
	EnablePaint  bool `json:"enablePaint"`
	MaxHue       int  `json:"maxHue"`
	MinHue       int  `json:"minHue"`
	MaxSaturatio int  `json:"maxSaturatio"`
	MinSaturatio int  `json:"minSaturatio"`
	MaxLightness int  `json:"maxLightness"`
	MinLightness int  `json:"minLightness"`
}

const (
	defaultEnablePaint  = true
	defaultMaxHue       = 360
	defaultMinHue       = 0
	defaultMaxSaturatio = 100
	defaultMinSaturatio = 30
	defaultMaxLightness = 100
	defaultMinLightness = 60
)

func LoadOrNewConfig() *Config {
	config, err := loadConfig(ConfigPath)
	if err != nil {
		config = newConfig()
		SaveConfig(config)
	}
	return config
}

func SaveConfig(config *Config) error {
	return saveConfig(config, ConfigPath)
}

// H is 0 to 360,
// S is 0 to 100,
// L is 50 to 100
func newConfig() *Config {
	config := &Config{
		Color: &ColorConfig{
			EnablePaint:  defaultEnablePaint,
			MaxHue:       defaultMaxHue,
			MinHue:       defaultMinHue,
			MaxSaturatio: defaultMaxSaturatio,
			MinSaturatio: defaultMinSaturatio,
			MaxLightness: defaultMaxLightness,
			MinLightness: defaultMinLightness,
		},
	}
	return config
}

func loadConfig(dataPath string) (*Config, error) {
	b, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	return &config, err
}

func saveConfig(config *Config, dataPath string) error {
	if !util.IsDir(dataPath) {
		if err := os.MkdirAll(filepath.Dir(dataPath), 0755); err != nil {
			return err
		}
	}

	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}

	if err := util.SaveBytes(b, dataPath); err != nil {
		return err
	}
	return nil
}
