package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
)

type (
	// state contains all of the entities that can frequently change.
	state struct {
		lidState      lidState   // current state of laptop lid
		powerState    powerState // current power state (battery/ac)
		mode          mode
		allDisplays   []display // current displays, returned by hyprctl monitors
		laptopDisplay display
	}

	// mode is the operating mode of the app.
	mode int
)

var commonLaptopDisplays = []string{
	"edp1",
}

const laptopNameEnv = "LAPTOP_DISPLAY_NAME"
const (
	modeNormal mode = iota
	modeSuspending
)

func (s *state) ready() bool {
	if s == nil {
		slog.Error("state ready check", "error", "state nil")
		return false
	}

	var notReady []string
	if len(s.allDisplays) == 0 {
		notReady = append(notReady, "allDisplays")
	}

	if !displayReady(s.laptopDisplay) {
		notReady = append(notReady, "laptopDisplay")
	}

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

func displayReady(m display) bool {
	return m.Name != ""
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

	am, err := hc.listDisplays()
	if err != nil {
		return nil, fmt.Errorf("listing displays: %w", err)
	}

	lm, err := identifyLaptopDisplay(am)
	if err != nil {
		return nil, fmt.Errorf("identifying laptop display: %w", err)
	}
	slog.Info("identified laptop display", "name", lm.Name, "desc", lm.Description)

	return &state{
		lidState:      ls,
		powerState:    ps,
		allDisplays:   am,
		laptopDisplay: lm,
	}, nil
}

func identifyLaptopDisplay(displays []display) (display, error) {
	env := os.Getenv(laptopNameEnv)
	for _, m := range displays {
		trimmed := trimmedDisplayName(m.Name)
		if slices.Contains(commonLaptopDisplays, trimmed) {
			return m, nil
		} else if env != "" && trimmed == env {
			return m, nil
		}
	}

	return display{}, errors.New("could not identify a laptop display")
}

func trimmedDisplayName(name string) string {
	name = strings.ToLower(name)
	return strings.ReplaceAll(name, "-", "")
}
