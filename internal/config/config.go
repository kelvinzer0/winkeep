package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	BaseDir    string `json:"-"`
	AutoRestart bool   `json:"auto_restart"`
	LogMaxSize  int64  `json:"log_max_size"`
}

func Dir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, _ := os.UserHomeDir()
		appData = filepath.Join(home, ".config")
	}
	return filepath.Join(appData, "winkeep")
}

func Load() (*Config, error) {
	cfg := &Config{
		BaseDir:     Dir(),
		AutoRestart: false,
		LogMaxSize:  10 * 1024 * 1024,
	}

	cfgPath := filepath.Join(Dir(), "config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, cfg.Save()
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Save() error {
	if err := os.MkdirAll(c.BaseDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.BaseDir, "config.json"), data, 0644)
}
