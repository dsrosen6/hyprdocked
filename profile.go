package main

import (
	"fmt"
	"log/slog"
	"strings"
)

type (
	profile struct {
		Name          string         `json:"name"`
		Conditions    conditions     `json:"conditions"`
		MonitorStates []monitorState `json:"monitor_states"`
		valid         bool
	}

	conditions struct {
		LidState         *lidState   `json:"lid_state"`
		PowerState       *powerState `json:"power_state"`
		EnabledMonitors  []string    `json:"enabled_monitors"`
		DisabledMonitors []string    `json:"disabled_monitors"`
	}

	monitorState struct {
		Label  string `json:"label"`
		Preset string `json:"preset"`
	}
)

func (a *app) validateAllProfiles() {
	for _, p := range a.profiles {
		a.validateProfile(p)
	}
}

func (a *app) validateProfile(p *profile) {
	var invalidReasons []string
	if p.Conditions.LidState != nil {
		parsed := parseLidState(string(*p.Conditions.LidState))
		if parsed == lidStateUnknown {
			invalidReasons = append(invalidReasons, "lid state")
		}
	}

	if p.Conditions.PowerState != nil {
		parsed := parsePowerState(string(*p.Conditions.PowerState))
		if parsed == powerStateUnknown {
			invalidReasons = append(invalidReasons, "power state")
		}
	}

	var invalidEnabled []string
	for _, m := range p.Conditions.EnabledMonitors {
		if !a.validMonitorLabel(m) {
			invalidEnabled = append(invalidEnabled, m)
		}
	}

	var invalidDisabled []string
	for _, m := range p.Conditions.DisabledMonitors {
		if !a.validMonitorLabel(m) {
			invalidDisabled = append(invalidDisabled, m)
		}
	}

	var invalidStateLabels, invalidStatePresets []string
	for _, s := range p.MonitorStates {
		if !a.validMonitorLabel(s.Label) {
			invalidStateLabels = append(invalidStateLabels, s.Label)
			continue
		}

		if !a.validMonitorPreset(s.Label, s.Preset) {
			invalidStatePresets = append(invalidStatePresets, fmt.Sprintf("%s:%s", s.Label, s.Preset))
		}
	}

	if len(invalidEnabled) > 0 {
		s := fmt.Sprintf("enabled monitor conditions: [%s]", strings.Join(invalidEnabled, ","))
		invalidReasons = append(invalidReasons, s)
	}

	if len(invalidDisabled) > 0 {
		s := fmt.Sprintf("disabled monitor conditions: [%s]", strings.Join(invalidDisabled, ","))
		invalidReasons = append(invalidReasons, s)
	}

	if len(invalidStateLabels) > 0 {
		s := fmt.Sprintf("monitor state labels: [%s]", strings.Join(invalidStateLabels, ","))
		invalidReasons = append(invalidReasons, s)
	}

	if len(invalidStatePresets) > 0 {
		s := fmt.Sprintf("monitor preset labels: [%s]", strings.Join(invalidStatePresets, ","))
		invalidReasons = append(invalidReasons, s)
	}

	if len(invalidReasons) > 0 {
		slog.Warn("profile invalid", "name", p.Name, "invalid_reasons", strings.Join(invalidReasons, ", "))
		p.valid = false
		return
	}

	p.valid = true
}

func (a *app) validMonitorLabel(label string) bool {
	return validMonitorLabel(a.monitors, label)
}

func (a *app) validMonitorPreset(monitor, preset string) bool {
	if !a.validMonitorLabel(monitor) {
		return false
	}

	return validMonitorPreset(a.monitors[monitor].Presets, preset)
}

func validMonitorLabel(monitors monitorConfigMap, label string) bool {
	if _, ok := monitors[label]; ok {
		return true
	}
	return false
}

func validMonitorPreset(presets monitorPresetMap, label string) bool {
	if _, ok := presets[label]; ok {
		return true
	}
	return false
}
