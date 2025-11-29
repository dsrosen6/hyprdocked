// Package cfg handles reading and setting config variables.
package cfg

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dsrosen6/hyprlaptop/internal/models"
)

type Config struct {
	PrimaryMonitorName string           `json:"primary_monitor_name"`
	Monitors           []models.Monitor `json:"monitors"`
}

var defaultCfg = &Config{
	PrimaryMonitorName: "",
	Monitors:           []models.Monitor{},
}

func ReadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		slog.Info("no config file found; creating default")
		if err := createDefaultFile(path); err != nil {
			return nil, fmt.Errorf("creating default config file: %w", err)
		}
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(file, cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling json: %w", err)
	}

	return cfg, nil
}

func createDefaultFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("checking and/or creating config directory: %w", err)
	}

	str, err := json.MarshalIndent(defaultCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling json: %w", err)
	}

	if err := os.WriteFile(path, str, 0o644); err != nil {
		return fmt.Errorf("writing json to file: %w", err)
	}

	return nil
}
