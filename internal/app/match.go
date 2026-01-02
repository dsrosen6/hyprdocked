package app

import (
	"log/slog"

	"github.com/dsrosen6/hyprlaptop/internal/hypr"
)

func (a *App) FindMatchingProfiles() *Profile {
	var pr *Profile
	for _, p := range a.Profiles {
		if a.State.matchesProfile(p) {
			pr = &p
		}
	}

	return pr
}

func (s *State) matchesProfile(p Profile) bool {
	if p.LidState != nil {
		match := *p.LidState == s.LidState
		slog.Debug(
			"lid state comparison",
			"profile_name", p.Name,
			"profile_lid", p.LidState.String(),
			"state_lid", s.LidState.String(),
			"matches", match,
		)
		if !match {
			slog.Debug("lid state mismatch, rejecting profile", "profile", p.Name)
			return false
		}
	}

	if p.PowerState != nil {
		match := *p.PowerState == s.PowerState
		slog.Debug(
			"power state comparison",
			"profile_name", p.Name,
			"profile_power", p.PowerState.String(),
			"state_power", s.PowerState.String(),
			"matches", match,
		)
		if !match {
			slog.Debug("power state mismatch, rejecting profile", "profile", p.Name)
			return false
		}
	}

	matchedM := 0
	for _, m := range p.Monitors {
		matched := false
		for _, sm := range s.Monitors {
			if m.matchesHyprMonitor(sm) {
				matched = true
				break
			}
		}

		if matched {
			matchedM++
		}
	}

	return matchedM == len(p.Monitors)
}

func (m MonitorIdentifiers) matchesHyprMonitor(hm hypr.Monitor) bool {
	if m.Name == nil && m.Description == nil && m.Make == nil && m.Model == nil {
		return false
	}

	if m.Name != nil {
		if *m.Name != hm.Name {
			return false
		}
	}

	if m.Description != nil {
		if *m.Description != hm.Description {
			return false
		}
	}

	if m.Make != nil {
		if *m.Make != hm.Make {
			return false
		}
	}

	if m.Model != nil {
		if *m.Model != hm.Model {
			return false
		}
	}

	return true
}
