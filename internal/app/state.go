package app

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/dsrosen6/hyprlaptop/internal/hypr"
	"github.com/dsrosen6/hyprlaptop/internal/power"
)

type State struct {
	Monitors   hypr.MonitorMap
	LidState   power.LidState
	PowerState power.PowerState
}

func (s *State) Ready() bool {
	if s == nil {
		slog.Error("state ready check", "error", "state nil")
		return false
	}

	var notReady []string
	if s.LidState == power.LidStateUnknown {
		notReady = append(notReady, "lid")
	}

	if s.PowerState == power.PowerStateUnknown {
		notReady = append(notReady, "power")
	}

	if len(s.Monitors) == 0 {
		notReady = append(notReady, "monitors")
	}

	if len(notReady) > 0 {
		nr := strings.Join(notReady, ",")
		slog.Info("ready check: one or more states not ready", "states", nr)
		return false
	}

	return true
}

func (s *State) logString() string {
	elements := []string{
		fmt.Sprintf("lid:%s", s.LidState.String()),
		fmt.Sprintf("power:%s", s.PowerState.String()),
	}

	for _, m := range s.Monitors {
		elements = append(elements, fmt.Sprintf("monitor:%s", m.Name))
	}

	return strings.Join(elements, ",")
}
