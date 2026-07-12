package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Port        int    `json:"port"`
	StoragePath string `json:"storagePath"`
	Token       string `json:"token"` // bearer token clients must present
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zipp-nest", "config.json")
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if os.IsNotExist(err) {
		return defaultConfig()
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// configs from before auth get a token on first load
	if cfg.Token == "" {
		token, err := newToken()
		if err != nil {
			return nil, err
		}
		cfg.Token = token
		if err := cfg.save(); err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}

func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func defaultConfig() (*Config, error) {
	home, _ := os.UserHomeDir()
	token, err := newToken()
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		Port:        9090,
		StoragePath: filepath.Join(home, ".zipp-nest", "backups"),
		Token:       token,
	}
	if err := cfg.save(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) save() error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
