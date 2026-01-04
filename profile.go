package main

import (
	"log/slog"
)

type (
	profile struct {
		Name              string         `json:"name"`
		Conditions        conditions     `json:"conditions"`
		MonitorStates     []monitorState `json:"monitor_states"`
		ExactMatch        bool           `json:"exact_match"`
		DisableUndeclared bool           `json:"disable_undeclared_monitors"`
		valid             bool
	}

	conditions struct {
		LidState        *lidState   `json:"lid_state"`
		PowerState      *powerState `json:"power_state"`
		EnabledMonitors []string    `json:"enabled_monitors"`
	}

	monitorState struct {
		Label   string  `json:"label"`
		Disable bool    `json:"disabled"`
		Preset  *string `json:"preset"`
	}
)

func (p *profile) matchesState(lookup labelLookup, state *state) bool {
	if p.Conditions.LidState != nil {
		if *p.Conditions.LidState != state.LidState {
			return false
		}
	}

	if p.Conditions.PowerState != nil {
		if *p.Conditions.PowerState != state.PowerState {
			return false
		}
	}

	for _, label := range p.Conditions.EnabledMonitors {
		lm, ok := lookup[label]
		if !ok || !lm.CurrentlyEnabled {
			return false
		}
	}

	return true
}

func (p *profile) validate(monitors monitorConfigMap) {
	valid := true
	pLog := slog.Default().With(slog.String("profile_name", p.Name))

	if p.Conditions.LidState != nil {
		parsed := parseLidState(string(*p.Conditions.LidState))
		if parsed == lidStateUnknown {
			valid = false
			pLog.Warn("invalid condition: lid state")
		}
	}

	if p.Conditions.PowerState != nil {
		parsed := parsePowerState(string(*p.Conditions.PowerState))
		if parsed == powerStateUnknown {
			pLog.Warn("invalid condition: power state")
		}
	}

	for _, m := range p.Conditions.EnabledMonitors {
		if !validMonitorLabel(monitors, m) {
			valid = false
			pLog.Warn("invalid condition: enabled monitor", "label", m)
		}
	}

	for _, s := range p.MonitorStates {
		if !validMonitorLabel(monitors, s.Label) {
			valid = false
			pLog.Warn("invalid monitor state", "label", s.Label, "reason", "label not found")
			continue
		}

		if s.Preset != nil {
			if s.Disable {
				valid = false
				pLog.Warn("invalid monitor state", "label", s.Label, "reason", "conflict: disabled set to true, but preset declared")
				continue
			}

			if !validMonitorPreset(monitors[s.Label].Presets, *s.Preset) {
				valid = false
				pLog.Warn("invalid monitor state", "label", s.Label, "reason", "preset not found", "preset", *s.Preset)
			}
		}
	}

	p.valid = valid
}

func getMatchingProfile(pr []*profile, lookup labelLookup, state *state) *profile {
	var matched *profile
	for _, p := range pr {
		if p.matchesState(lookup, state) {
			if !p.valid {
				slog.Warn("conditions met for profile, but the profile is invalid; skipping", "profile", p.Name)
				continue
			}
			matched = p
		}
	}

	return matched
}

func validateProfiles(pr []*profile, m monitorConfigMap) {
	for _, p := range pr {
		p.validate(m)
	}
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
