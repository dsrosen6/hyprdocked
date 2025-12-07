package app

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/dsrosen6/hyprlaptop/internal/hypr"
	"github.com/dsrosen6/hyprlaptop/internal/lid"
)

type (
	getOutputResult struct {
		laptopName string
		displays   hypr.MonitorMap
		lidState   lid.State
	}

	outputsStatus string
)

const (
	statusUnknown outputsStatus = "UNKNOWN"
	statusOLLC    outputsStatus = "ONLY_LAPTOP_LID_CLOSED"
	statusOLLO    outputsStatus = "ONLY_LAPTOP_LID_OPEN"
	statusWELC    outputsStatus = "WITH_EXTERNAL_LID_CLOSED"
	statusWELO    outputsStatus = "WITH_EXTERNAL_LID_OPEN"
)

func (a *App) Run() error {
	o, err := a.getOutputs()
	if err != nil {
		return fmt.Errorf("getting output info: %w", err)
	}

	s := o.statusShouldBe()
	slog.Info(fmt.Sprintf("status should be: %s", s))

	payloads := a.createPayloads(o, s)
	needUpdate := false
	for _, p := range payloads {
		slog.Info("display detected", logMonitorAttr(p))
		if p.update {
			needUpdate = true
		}
	}

	if !needUpdate {
		slog.Info("no updates needed")
		return nil
	}

	if err := a.updateMonitors(payloads); err != nil {
		return fmt.Errorf("updating monitors: %w", err)
	}

	return nil
}

func (a *App) getOutputs() (*getOutputResult, error) {
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

	return &getOutputResult{
		laptopName: a.Cfg.LaptopDisplay.Name,
		displays:   current,
		lidState:   ls,
	}, nil
}

func (a *App) updateMonitors(monitors []displayPayload) error {
	wg := new(sync.WaitGroup)
	errc := make(chan error, len(monitors))

	for _, p := range monitors {
		if !p.update {
			continue
		}

		wg.Add(1)
		go func(p displayPayload) {
			defer wg.Done()
			m := p.out
			if p.enable {
				if err := a.Hctl.EnableOrUpdateMonitor(m); err != nil {
					errc <- fmt.Errorf("enabling or updating monitor %s: %w", m.Name, err)
				}
				slog.Info("display enabled", "name", m.Name)
			} else {
				if err := a.Hctl.DisableMonitor(m); err != nil {
					errc <- fmt.Errorf("disabling monitor %s: %w", m.Name, err)
				}
				slog.Info("display disabled", "name", m.Name)
			}
		}(p)
	}

	go func() {
		wg.Wait()
		close(errc)
	}()

	hasErr := false
	for err := range errc {
		hasErr = true
		slog.Error(err.Error())
	}

	if hasErr {
		return errors.New("failed to update one or more external monitors; see logs")
	}

	return nil
}

func (a *App) isLaptopDisplay(m hypr.Monitor) bool {
	return m.Name == a.Cfg.LaptopDisplay.Name
}

func (a *App) getDisplayFromConfig(m hypr.Monitor) (hypr.Monitor, bool) {
	if a.isLaptopDisplay(m) {
		return a.Cfg.LaptopDisplay, true
	}

	if c, ok := a.Cfg.ExternalDisplays[m.Name]; ok {
		return c, true
	}

	return hypr.Monitor{}, false
}

// statusShouldBe checks the state of monitors and lid status, and returns the status
// that hyprlaptop should be switched to (if it isn't already)
func (o *getOutputResult) statusShouldBe() outputsStatus {
	// check if laptop is the only monitor
	if _, ok := o.displays[o.laptopName]; ok && len(o.displays) == 1 {
		return onlyLaptopStates(o.lidState)
	}

	return withExternalStates(o.lidState)
}

func onlyLaptopStates(ls lid.State) outputsStatus {
	switch ls {
	case lid.StateOpen:
		return statusOLLO
	case lid.StateClosed:
		return statusOLLC
	case lid.StateUnknown:
		return statusUnknown
	default:
		return statusUnknown
	}
}

func withExternalStates(ls lid.State) outputsStatus {
	switch ls {
	case lid.StateOpen:
		return statusWELO
	case lid.StateClosed:
		return statusWELC
	case lid.StateUnknown:
		return statusUnknown
	default:
		return statusUnknown
	}
}

func logMonitorAttr(p displayPayload) slog.Attr {
	return slog.Group(
		p.out.Name,
		slog.Bool("from_config", p.fromConfig),
		slog.Bool("update_needed", p.update),
		slog.Int64("width", p.out.Width),
		slog.Int64("height", p.out.Height),
		slog.Float64("refresh_rate", p.out.RefreshRate),
		slog.Int64("x", p.out.X),
		slog.Int64("y", p.out.Y),
		slog.Float64("scale", p.out.Scale),
	)
}
