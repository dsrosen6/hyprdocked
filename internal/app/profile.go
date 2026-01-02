package app

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/dsrosen6/hyprlaptop/internal/config"
	"github.com/dsrosen6/hyprlaptop/internal/power"
)

type Profile struct {
	Name       string
	Monitors   map[string]MonitorIdentifiers
	LidState   *power.LidState
	PowerState *power.PowerState
}

func (a *App) LogProfiles() {
	for _, p := range a.Profiles {
		slog.Info("profile loaded", "details", p.logString())
	}
}

func profilesFromConfig(cfp []config.Profile) []Profile {
	profiles := make([]Profile, 0, len(cfp))
	for _, p := range cfp {
		profiles = append(profiles, profileFromConfig(p))
	}

	return profiles
}

func profileFromConfig(cfp config.Profile) Profile {
	p := &Profile{}

	if cfp.Name != "" {
		p.Name = cfp.Name
	}

	if cfp.Monitors != nil {
		p.Monitors = make(map[string]MonitorIdentifiers, len(cfp.Monitors))
		for k, v := range cfp.Monitors {
			p.Monitors[k] = monitorFromConfig(v)
		}
	}

	if cfp.LidState != nil {
		ls := power.ParseLidState(*cfp.LidState)
		if ls != power.LidStateUnknown {
			p.LidState = &ls
		}
	}

	if cfp.PowerState != nil {
		ps := power.ParsePowerState(*cfp.PowerState)
		if ps != power.PowerStateUnknown {
			p.PowerState = &ps
		}
	}

	return *p
}

func (p Profile) logString() string {
	var elements []string
	if p.Name != "" {
		elements = append(elements, fmt.Sprintf("name:%s", p.Name))
	}

	for k, v := range p.Monitors {
		s := fmt.Sprintf("monitor:{%s}", v.logString(k))
		elements = append(elements, s)
	}

	if p.LidState != nil {
		ls := p.LidState.String()
		elements = append(elements, fmt.Sprintf("lid:%s", ls))
	}

	if p.PowerState != nil {
		ps := p.PowerState.String()
		elements = append(elements, fmt.Sprintf("power:%s", ps))
	}

	return strings.Join(elements, ",")
}
