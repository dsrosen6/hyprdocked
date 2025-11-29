package main

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

type options struct {
	configFile string
}

func parseFlags() (*options, error) {
	flag.StringVar(&cfgFile, "c", "", "specify a config file")

	if cfgFile == "" {
		cd, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("getting user config directory: %w", err)
		}
		cfgFile = filepath.Join(cd, defaultDir, defaultFile)
	}
	return &options{
		configFile: cfgFile,
	}, nil
}
