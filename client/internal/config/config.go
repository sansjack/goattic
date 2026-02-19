package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const appDir = "goattic"

type Config struct {
	APIURL string `json:"apiUrl"`
	APIKey string `json:"apiKey"`
}

func dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate config dir: %w", err)
	}
	return filepath.Join(base, appDir), nil
}

func Load() (*Config, error) {
	d, err := dir()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filepath.Join(d, "config.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no config found — run: goattic configure")
		}
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	d, err := dir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(d, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(d, "config.json"), data, 0600)
}
