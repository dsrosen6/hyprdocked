package app

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/dsrosen6/hyprlaptop/internal/config"
	"github.com/dsrosen6/hyprlaptop/internal/hypr"
)

type App struct {
	Hctl *hypr.HyprctlClient
	Cfg  *config.Config
}

func NewApp(cfg *config.Config, hc *hypr.HyprctlClient) *App {
	if os.Getenv("DEBUG") == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	return &App{
		Hctl: hc,
		Cfg:  cfg,
	}
}

func (a *App) SaveCurrentMonitors(laptop string) error {
	monitors, err := a.Hctl.ListMonitors()
	if err != nil {
		return fmt.Errorf("listing monitors: %w", err)
	}

	var lm *hypr.Monitor
	if laptop == "" {
		for _, m := range monitors {
			if strings.Contains(m.Name, "eDP") {
				lm = &m
			}
		}
	} else {
		l, ok := monitors[laptop]
		if ok {
			lm = &l
		}
	}

	if lm == nil {
		return fmt.Errorf("monitor '%s' not found", laptop)
	}

	externals := map[string]hypr.Monitor{}
	for _, m := range monitors {
		if m.Name != lm.Name {
			externals[m.Name] = m
		}
	}

	a.Cfg.LaptopMonitor = *lm
	a.Cfg.ExternalMonitors = externals

	if err := a.Cfg.Write(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
