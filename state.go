package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type (
	// state contains all of the entities that can frequently change.
	state struct {
		lidState          lidState   // current state of laptop lid
		powerState        powerState // current power state (battery/ac)
		mode              mode
		monitors          []monitor // current monitors, returned by hyprctl monitors
		suspendedMonitors []monitor
	}

	// mode is the operating mode of the app.
	mode int
)

const (
	modeNormal mode = iota
	modeSuspending
	modeWaking
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

func getInitialState(ctx context.Context, hc *hyprClient, lh *lidHandler, ph *powerHandler) (*state, error) {
	ls, err := lh.getCurrentState(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting lid status: %w", err)
	}

	ps, err := ph.getCurrentState(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting power status: %w", err)
	}

	m, err := hc.listMonitors()
	if err != nil {
		return nil, fmt.Errorf("listing monitors: %w", err)
	}

	return &state{
		lidState:   ls,
		powerState: ps,
		monitors:   m,
	}, nil
}
