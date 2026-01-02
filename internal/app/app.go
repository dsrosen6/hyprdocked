package app

import (
	"log/slog"

	"github.com/dsrosen6/hyprlaptop/internal/config"
	"github.com/dsrosen6/hyprlaptop/internal/hypr"
)

type App struct {
	Hctl     *hypr.HyprctlClient
	Cfg      *config.Config
	Profiles []Profile
	State    *State
}

func NewApp(cfg *config.Config, hc *hypr.HyprctlClient) *App {
	return &App{
		Hctl:     hc,
		Cfg:      cfg,
		Profiles: profilesFromConfig(cfg.Profiles),
		State:    &State{},
	}
}

func (a *App) RunUpdater() error {
	slog.Info("checking for matches", "state", a.State.logString())
	m := a.FindMatchingProfiles()
	if m != nil {
		slog.Info("found match", "profile_name", m.Name)
	}
	return nil
}
