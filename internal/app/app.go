// Package app handles application logic.
package app

import (
	"log/slog"
	"os"

	"github.com/dsrosen6/hyprlaptop/internal/config"
	"github.com/dsrosen6/hyprlaptop/internal/hypr"
)

type App struct {
	hctl *hypr.HyprctlClient
	cfg  *config.Config
}

func NewApp(cfg *config.Config, hc *hypr.HyprctlClient) *App {
	if os.Getenv("DEBUG") == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	return &App{
		hctl: hc,
		cfg:  cfg,
	}
}
