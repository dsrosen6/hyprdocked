package main

import (
	"errors"
	"fmt"
)

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
