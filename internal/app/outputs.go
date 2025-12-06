package app

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/dsrosen6/hyprlaptop/internal/hypr"
	"github.com/dsrosen6/hyprlaptop/internal/lid"
)

type (
	outputs struct {
		laptopName string
		monitors   map[string]hypr.Monitor
		lidState   lid.State
	}

	outputsStatus string
)

const (
	statusUnknown           outputsStatus = "Unknown"
	statusFullShut          outputsStatus = "FullShut"
	statusOnlyLaptop        outputsStatus = "OnlyLaptop"
	statusExternalClamshell outputsStatus = "ExternalClamshell"
	statusExternalOpen      outputsStatus = "ExternalOpen"
)

func (a *App) Run() error {
	o, err := a.getOutputs()
	if err != nil {
		return fmt.Errorf("getting output info: %w", err)
	}

	s := o.statusShouldBe()
	slog.Info(fmt.Sprintf("status should be: %s", s))
	return nil
}

func (a *App) getOutputs() (*outputs, error) {
	current, err := a.Hctl.ListMonitors()
	if err != nil {
		return nil, fmt.Errorf("listing current monitors: %w", err)
	}

	var names []string
	for _, m := range current {
		names = append(names, m.Name)
	}
	slog.Info("monitors detected", "names", strings.Join(names, ","))

	ls, err := lid.GetState()
	if err != nil {
		return nil, fmt.Errorf("getting lid status: %w", err)
	}
	slog.Info(fmt.Sprintf("lid state: %s", ls))

	return &outputs{
		laptopName: a.Cfg.LaptopMonitor.Name,
		monitors:   current,
		lidState:   ls,
	}, nil
}

// statusShouldBe checks the state of monitors and lid status, and returns the status
// that hyprlaptop should be switched to (if it isn't already)
func (o *outputs) statusShouldBe() outputsStatus {
	// check if laptop is the only monitor
	if _, ok := o.monitors[o.laptopName]; ok && len(o.monitors) == 1 {
		return onlyLaptopStates(o.lidState)
	}

	return withExternalStates(o.lidState)
}

func onlyLaptopStates(ls lid.State) outputsStatus {
	switch ls {
	case lid.StateOpen:
		return statusOnlyLaptop
	case lid.StateClosed:
		return statusFullShut
	case lid.StateUnknown:
		return statusUnknown
	default:
		return statusUnknown
	}
}

func withExternalStates(ls lid.State) outputsStatus {
	switch ls {
	case lid.StateOpen:
		return statusExternalOpen
	case lid.StateClosed:
		return statusExternalClamshell
	case lid.StateUnknown:
		return statusUnknown
	default:
		return statusUnknown
	}
}
