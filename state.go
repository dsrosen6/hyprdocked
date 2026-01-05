package main

import (
	"log/slog"
	"strings"
)

type (
	state struct {
		lidState   lidState   // current state of laptop lid
		powerState powerState // current power state (battery/ac)
		monitors   []monitor  // current monitors, returned by hyprctl monitors
	}
)

func (s *state) ready() bool {
	if s == nil {
		slog.Error("state ready check", "error", "state nil")
		return false
	}

	var notReady []string
	if s.lidState == lidStateUnknown {
		notReady = append(notReady, "lid")
	}

	if s.powerState == powerStateUnknown {
		notReady = append(notReady, "power")
	}

	if len(notReady) > 0 {
		nr := strings.Join(notReady, ",")
		slog.Info("ready check: one or more states not ready", "states", nr)
		return false
	}

	return true
}

func (s *state) getMonitorByIdentifiers(ident monitorIdentifiers) *monitor {
	for _, m := range s.monitors {
		if matchesIdentifiers(m, ident) {
			return &m
		}
	}

	return nil
}
