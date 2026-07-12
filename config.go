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
		cfg.Token = newToken()
		if err := cfg.save(); err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}

func newToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func defaultConfig() (*Config, error) {
	home, _ := os.UserHomeDir()
	cfg := &Config{
		Port:        9090,
		StoragePath: filepath.Join(home, ".zipp-nest", "backups"),
		Token:       newToken(),
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
