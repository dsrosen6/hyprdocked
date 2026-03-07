package app

import (
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Debug           bool       `mapstructure:"debug"`
	Laptop          string     `mapstructure:"laptop"`
	SuspendIdle     bool       `mapstructure:"suspend-idle"`
	SuspendClosed   bool       `mapstructure:"suspend-closed"`
	PostUpdateHooks []PostHook `mapstructure:"post-hooks"`
	SequentialHooks bool       `mapstructure:"sequential-hooks"`
}

type PostHook struct {
	Command        string `mapstructure:"command"`
	OnStatusChange bool   `mapstructure:"on-status-change"`
}

// onConfigChange handles live updates when a config file change is detected.
func (a *App) onConfigChange(e fsnotify.Event) {
	if time.Since(a.lastConfigChange) < 100*time.Millisecond {
		return
	}
	a.lastConfigChange = time.Now()

	var newCfg Config
	if err := viper.Unmarshal(&newCfg); err != nil {
		slog.Error("reloading config", "error", err)
		return
	}

	select {
	case a.listener.configCh <- newCfg:
		slog.Info("config reloaded", "config", a.Config)
	default:
	}
}
