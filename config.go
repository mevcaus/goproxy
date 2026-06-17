package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Port     int      `json:"port"`
	Backends []string `json:"backends"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if len(cfg.Backends) == 0 {
		return nil, fmt.Errorf("config must define at least one backend")
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}

	return &cfg, nil
}
