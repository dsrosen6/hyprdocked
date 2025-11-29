// Package cli provides cli commands for hyprlaptop.
package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultDir  = "hyprlaptop"
	defaultFile = "config.json"
)

var cfgFile string

type Options struct {
	ConfigFile string
}

func Parse() (*Options, error) {
	flag.StringVar(&cfgFile, "c", "", "specify a config file")

	if cfgFile == "" {
		cd, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("getting user config directory: %w", err)
		}
		cfgFile = filepath.Join(cd, defaultDir, defaultFile)
	}
	return &Options{
		ConfigFile: cfgFile,
	}, nil
}
