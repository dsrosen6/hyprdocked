package app

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dsrosen6/hyprdocked/internal/hypr"
	"github.com/dsrosen6/hyprdocked/internal/power"
	"github.com/godbus/dbus/v5"
	"github.com/spf13/viper"
)

type App struct {
	Config            Config
	hctl              *hypr.Client
	listener          *listener
	updating          bool
	configReloadTimer *time.Timer
	*state
}

type RunParams struct {
	LaptopMonitorName string
	SuspendOnIdle     bool
	SuspendOnClosed   bool
}

func newApp(cfg Config, hc *hypr.Client, l *listener, s *state) *App {
	return &App{
		Config:   cfg,
		hctl:     hc,
		listener: l,
		state:    s,
	}
}

func RunListener(c Config) error {
	hypr.WaitForEnvs()
	hyprClient, err := hypr.NewClient()
	if err != nil {
		return fmt.Errorf("creating hyprctl client: %w", err)
	}

	// Run an initial reload in case laptop display is already disabled. Assuming the laptop
	// display is correctly set to initially enable in the hyprland config, this will re-enable
	// it so hyprdocked can properly identify it.
	slog.Info("running hyprctl reload")
	if err := hyprClient.Reload(); err != nil {
		return fmt.Errorf("running hyprctl reload: %w", err)
	}

	var (
		hyprSock *hypr.SocketConn
		dbusConn *dbus.Conn
	)

	defer func() {
		if hyprSock != nil {
			if err := hyprSock.Close(); err != nil {
				slog.Error("closing hypr socket connection", "error", err)
			}
		}

		if dbusConn != nil {
			if err := dbusConn.Close(); err != nil {
				slog.Error("closing dbus connection", "error", err)
			}
		}
	}()
	hyprSock, err = hypr.NewSocketConn()
	if err != nil {
		return fmt.Errorf("creating hyprland socket connection: %w", err)
	}

	dbusConn, err = dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("creating dbus connection: %w", err)
	}

	lh := power.NewLidHandler(dbusConn)
	lp := listenerParams{
		hyprSockConn: hyprSock,
		lidHandler:   lh,
		dbusConn:     dbusConn,
	}

	l, err := newListener(lp)
	if err != nil {
		return fmt.Errorf("creating listener: %w", err)
	}

	sp := initialStateParams{
		laptopMonitorName: c.laptopMonitorName(),
		hyprClient:        hyprClient,
		lidHandler:        lh,
	}

	s, err := getInitialState(context.Background(), sp)
	if err != nil {
		return fmt.Errorf("getting initial state: %w", err)
	}

	a := newApp(c, hyprClient, l, s)
	slog.Info("app initialized",
		"laptop_monitor_name", a.laptopDisplay.Name,
		"status", a.statusString(),
		"suspend_idle", a.Config.SuspendIdle,
		"suspend_closed", a.Config.SuspendClosed,
	)
	a.laptopMonitor() // logs resolved settings at debug level

	// initial updater run before starting listener
	changed, _ := a.runUpdater()
	a.runPostHooks(changed)

	viper.OnConfigChange(a.onConfigChange)
	viper.WatchConfig()

	return a.listenAndHandle(context.Background())
}

func (a *App) laptopMonitor() hypr.Monitor {
	return a.resolveMonitor(a.laptopDisplay)
}

func (a *App) applyMonitorConfigs() {
	for _, m := range a.allDisplays {
		if m.Name == a.laptopDisplay.Name {
			continue // managed by updater
		}
		if a.Config.monitorSettingsFor(m.Name) == nil {
			continue
		}
		if err := a.hctl.EnableOrUpdateMonitor(a.resolveMonitor(m)); err != nil {
			slog.Error("applying monitor config", "monitor", m.Name, "error", err)
		}
	}
}

func (a *App) resolveMonitor(m hypr.Monitor) hypr.Monitor {
	cfg := a.Config.monitorSettingsFor(m.Name)
	if cfg == nil {
		slog.Debug("[MONITOR]no config overrides; using runtime values", "monitor", m.Name)
		return m
	}
	resolved := hypr.Monitor{
		Name:        m.Name,
		Width:       cmp.Or(cfg.Width, m.Width),
		Height:      cmp.Or(cfg.Height, m.Height),
		RefreshRate: cmp.Or(cfg.RefreshRate, m.RefreshRate),
		X:           cmp.Or(cfg.X, m.X),
		Y:           cmp.Or(cfg.Y, m.Y),
		Scale:       cmp.Or(cfg.Scale, m.Scale),
	}
	slog.Debug("[MONITOR]resolved monitor settings",
		"monitor", m.Name,
		"width", monitorFieldSrc(cfg.Width, m.Width),
		"height", monitorFieldSrc(cfg.Height, m.Height),
		"refresh_rate", monitorFieldSrc(cfg.RefreshRate, m.RefreshRate),
		"x", monitorFieldSrc(cfg.X, m.X),
		"y", monitorFieldSrc(cfg.Y, m.Y),
		"scale", monitorFieldSrc(cfg.Scale, m.Scale),
	)
	return resolved
}

func monitorFieldSrc[T comparable](cfgVal, runtimeVal T) string {
	var zero T
	if cfgVal != zero {
		return fmt.Sprintf("%v (config)", cfgVal)
	}
	return fmt.Sprintf("%v (runtime)", runtimeVal)
}
