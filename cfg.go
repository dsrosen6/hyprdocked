package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type config struct {
	LaptopMonitor    Monitor   `json:"laptop_monitor_name"`
	ExternalMonitors []Monitor `json:"external_monitors"`
}

var defaultCfg = &config{
	LaptopMonitor:    Monitor{},
	ExternalMonitors: []Monitor{},
}

func readConfig(path string) (*config, error) {
	cfg := &config{}
	if _, err := os.Stat(path); err != nil {
		slog.Info("no config file found; creating default")
		cfg = defaultCfg
		if err := cfg.write(path); err != nil {
			return nil, fmt.Errorf("creating default config file: %w", err)
		}
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := json.Unmarshal(file, cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling json: %w", err)
	}

	return cfg, nil
}

func (c *config) validate() error {
	if c.LaptopMonitor.Name == "" {
		return errors.New("laptop monitor name not set")
	}

	return nil
}

func (c *config) write(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("checking and/or creating config directory: %w", err)
	}

	str, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling json: %w", err)
	}

	if err := os.WriteFile(path, str, 0o644); err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}
