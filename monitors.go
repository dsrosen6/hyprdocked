package main

import (
	"errors"
	"fmt"
)

// Monitor matches the output of 'hyprctl monitors', and is also used for config.
type Monitor struct {
	Name        string  `json:"name"`
	Width       int64   `json:"width"`
	Height      int64   `json:"height"`
	RefreshRate float64 `json:"refreshRate"`
	X           int64   `json:"x"`
	Y           int64   `json:"y"`
	Scale       float64 `json:"scale"`
}

func (a *app) saveCurrentMonitors(laptopMtr string) error {
	if laptopMtr == "" {
		return errors.New("no laptop monitor name provided")
	}

	monitors, err := a.hc.listMonitors()
	if err != nil {
		return fmt.Errorf("fetching all current monitors via hyprctl: %w", err)
	}

	lm, valid := monitors[laptopMtr]
	if !valid {
		return fmt.Errorf("setting laptop monitor: monitor '%s' not found", laptopMtr)
	}

	externals := map[string]Monitor{}
	for _, m := range monitors {
		if m.Name != lm.Name {
			externals[m.Name] = m
		}
	}

	a.cfg.LaptopMonitor = lm
	a.cfg.ExternalMonitors = externals

	if err := a.cfg.write(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

func (a *app) setLaptopMonitor(name string) error {
	if name == "" {
		return errors.New("no monitor name provided")
	}

	monitors, err := a.hc.listMonitors()
	if err != nil {
		return fmt.Errorf("fetching monitors from hyprctl: %w", err)
	}

	m, valid := monitors[name]
	if !valid {
		return fmt.Errorf("monitor '%s' not found", name)
	}

	a.cfg.LaptopMonitor = m
	return a.cfg.write()
}

func (h *hyprctlClient) listMonitors() (map[string]Monitor, error) {
	var monitors []Monitor
	if err := h.runCommandWithUnmarshal([]string{"monitors"}, &monitors); err != nil {
		return nil, err
	}

	mm := make(map[string]Monitor, len(monitors))
	for _, m := range monitors {
		mm[m.Name] = m
	}

	return mm, nil
}

func (h *hyprctlClient) enableMonitor(m Monitor) error {
	args := []string{"keyword", "monitor", monitorToConfigString(m)}
	if _, err := h.runCommand(args); err != nil {
		return err
	}

	return nil
}

func (h *hyprctlClient) disableMonitor(m Monitor) error {
	args := []string{"keyword", "monitor", m.Name, "disable"}
	if _, err := h.runCommand(args); err != nil {
		return err
	}

	return nil
}

func monitorToConfigString(m Monitor) string {
	res := fmt.Sprintf("%dx%d", m.Width, m.Height)
	res = fmt.Sprintf("%s@%f", res, m.RefreshRate)
	xy := fmt.Sprintf("%dx%d", m.X, m.Y)
	scale := fmt.Sprintf("%f", m.Scale)
	return fmt.Sprintf("%s,%s,%s,%s", m.Name, res, xy, scale)
}
