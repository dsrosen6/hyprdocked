package app

import (
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const configReloadDelay = 100 * time.Millisecond

type (
	Config struct {
		Debug           bool              `mapstructure:"debug"`
		Monitors        []MonitorSettings `mapstructure:"monitors"`
		SuspendIdle     bool              `mapstructure:"suspend-idle"`
		SuspendClosed   bool              `mapstructure:"suspend-closed"`
		PostUpdateHooks []PostHook        `mapstructure:"post-hooks"`
		SequentialHooks bool              `mapstructure:"sequential-hooks"`
		SettleWindow    int               `mapstructure:"settle-window"`
	}

	MonitorSettings struct {
		Name        string  `mapstructure:"name"`
		Laptop      bool    `mapstructure:"laptop"`
		Width       int64   `mapstructure:"width"`
		Height      int64   `mapstructure:"height"`
		RefreshRate float64 `mapstructure:"refresh-rate"`
		X           int64   `mapstructure:"x"`
		Y           int64   `mapstructure:"y"`
		Scale       float64 `mapstructure:"scale"`
	}
)

type PostHook struct {
	Command        string `mapstructure:"command"`
	OnStatusChange bool   `mapstructure:"on-status-change"`
}

// onConfigChange handles live updates when a config file change is detected.
func (a *App) onConfigChange(e fsnotify.Event) {
	if a.configReloadTimer != nil {
		a.configReloadTimer.Stop()
	}

	a.configReloadTimer = time.AfterFunc(configReloadDelay, func() {
		var newCfg Config
		if err := viper.Unmarshal(&newCfg); err != nil {
			slog.Error("reloading config", "error", err)
			return
		}
		select {
		case a.listener.configCh <- newCfg:
			slog.Info("config reloaded", "config", newCfg)
		default:
		}
	})
}

func (c Config) laptopMonitorName() string {
	for _, m := range c.Monitors {
		if m.Laptop {
			return m.Name
		}
	}
	return ""
}

func (c Config) monitorSettingsFor(name string) *MonitorSettings {
	for i := range c.Monitors {
		if c.Monitors[i].Name == name {
			return &c.Monitors[i]
		}
	}
	return nil
}

