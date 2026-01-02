package app

import (
	"fmt"
	"strings"

	"github.com/dsrosen6/hyprlaptop/internal/config"
	"github.com/dsrosen6/hyprlaptop/internal/hypr"
)

type MonitorIdentifiers struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Make        *string `json:"make,omitempty"`
	Model       *string `json:"model,omitempty"`
}

func monitorFromConfig(cm config.MonitorIdentifiers) MonitorIdentifiers {
	return MonitorIdentifiers{
		Name:        cm.Name,
		Description: cm.Description,
		Make:        cm.Make,
		Model:       cm.Model,
	}
}

func monitorsFromHyprMap(hm hypr.MonitorMap) []MonitorIdentifiers {
	var monitors []MonitorIdentifiers
	for _, h := range hm {
		monitors = append(monitors, monitorFromHyprMonitor(h))
	}

	return monitors
}

func monitorFromHyprMonitor(h hypr.Monitor) MonitorIdentifiers {
	return MonitorIdentifiers{
		Name:        &h.Name,
		Description: &h.Description,
		Make:        &h.Make,
		Model:       &h.Model,
	}
}

func (m MonitorIdentifiers) logString(alias string) string {
	var elements []string

	if alias != "" {
		elements = append(elements, fmt.Sprintf("alias=%s", alias))
	}

	if m.Name != nil {
		elements = append(elements, fmt.Sprintf("name=%s", *m.Name))
	}

	if m.Description != nil {
		elements = append(elements, fmt.Sprintf("desc=%s", *m.Description))
	}

	if m.Make != nil {
		elements = append(elements, fmt.Sprintf("make=%s", *m.Make))
	}

	if m.Model != nil {
		elements = append(elements, fmt.Sprintf("model=%s", *m.Model))
	}

	return strings.Join(elements, ",")
}
