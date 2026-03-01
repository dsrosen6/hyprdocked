package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/dsrosen6/hyprdocked/internal/power"
)

type (
	// state contains all of the entities that can frequently change.
	state struct {
		lidState      power.LidState // current state of laptop lid
		mode          mode
		allDisplays   []display // current displays, returned by hyprctl monitors
		laptopDisplay display
	}

	initialStateParams struct {
		laptopMonitorName string
		hyprClient        *hyprClient
		lidHandler        *power.LidHandler
	}

	// mode is the operating mode of the app.
	mode int
)

var commonLaptopDisplays = []string{
	"edp1",
}

const (
	modeNormal mode = iota
	modeIdle
)

func (s *state) ready() bool {
	if s == nil {
		slog.Error("state ready check", "error", "state nil")
		return false
	}

	var notReady []string
	if !displayReady(s.laptopDisplay) {
		notReady = append(notReady, "laptopDisplay")
	}

	if s.lidState == power.LidStateUnknown {
		notReady = append(notReady, "lid")
	}

	if len(notReady) > 0 {
		nr := strings.Join(notReady, ",")
		slog.Info("ready check: one or more states not ready", "states", nr)
		return false
	}

	return true
}

func (s *state) laptopIsEnabled() bool {
	for _, m := range s.allDisplays {
		if m.Name == s.laptopDisplay.Name {
			return true
		}
	}

	return false
}

func (m mode) string() string {
	switch m {
	case modeNormal:
		return "normal"
	case modeIdle:
		return "suspending"
	default:
		return "unknown"
	}
}

func displayReady(m display) bool {
	return m.Name != ""
}

func getInitialState(ctx context.Context, sp initialStateParams) (*state, error) {
	ls, err := sp.lidHandler.GetCurrentState(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting lid status: %w", err)
	}

	ds, err := sp.hyprClient.listDisplays()
	if err != nil {
		return nil, fmt.Errorf("listing displays: %w", err)
	}

	lm, err := identifyLaptopDisplay(sp.laptopMonitorName, ds)
	if err != nil {
		return nil, fmt.Errorf("identifying laptop display: %w", err)
	}
	slog.Info("identified laptop display", "name", lm.Name, "desc", lm.Description)

	return &state{
		lidState:      ls,
		allDisplays:   ds,
		laptopDisplay: lm,
	}, nil
}

func identifyLaptopDisplay(cfgName string, displays []display) (display, error) {
	for _, m := range displays {
		trimmed := trimmedDisplayName(m.Name)
		if slices.Contains(commonLaptopDisplays, trimmed) {
			return m, nil
		} else if cfgName != "" && trimmed == cfgName {
			return m, nil
		}
	}

	return display{}, errors.New("could not identify a laptop display")
}

func trimmedDisplayName(name string) string {
	name = strings.ToLower(name)
	return strings.ReplaceAll(name, "-", "")
}
