package main

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/godbus/dbus/v5"
)

type app struct {
	hctl         *hyprClient
	cfg          *config
	listener     *listener
	currentState *state
}

func newApp(cfg *config, hc *hyprClient, l *listener) *app {
	return &app{
		hctl:         hc,
		cfg:          cfg,
		listener:     l,
		currentState: &state{},
	}
}

func run() error {
	cfg, err := initConfig("")
	if err != nil {
		return fmt.Errorf("initializing config: %w", err)
	}

	h, err := newHyprctlClient()
	if err != nil {
		return fmt.Errorf("creating hyprctl client: %w", err)
	}

	var (
		hs *hyprSocketConn
		dc *dbus.Conn
	)

	defer func() {
		if hs != nil {
			if err := hs.Close(); err != nil {
				slog.Error("closing hypr socket connection", "error", err)
			}
		}

		if dc != nil {
			if err := dc.Close(); err != nil {
				slog.Error("closing dbus connection", "error", err)
			}
		}
	}()

	hs, err = newHyprSocketConn()
	if err != nil {
		return fmt.Errorf("creating hyprland socket connection: %w", err)
	}

	dc, err = dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("creating dbus connection: %w", err)
	}

	l, err := newListener(hs, dc, cfg.path)
	if err != nil {
		return fmt.Errorf("creating listener: %w", err)
	}

	app := newApp(cfg, h, l)
	app.validateAllProfiles()

	return app.listen(context.Background())
}

// listen starts hyprlaptop's listener, which handles hyprctl display add/remove events
// and events from the hyprlaptop CLI.
func (a *app) listen(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	events := make(chan listenerEvent, 16)
	errc := make(chan error, 1)

	go func() {
		if err := a.listener.listen(ctx, events); err != nil {
			errc <- err
			cancel()
		}
	}()

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return nil // normal shutdown
			}

			slog.Info("received event from listener", "type", ev.Type, "details", ev.Details)
			switch ev.Type {
			case displayInitialEvent, displayAddEvent,
				displayRemoveEvent, displayUnknownEvent:
				m, err := a.hctl.listMonitors()
				if err != nil {
					slog.Error("listing current monitors", "error", err)
					continue
				}
				if !reflect.DeepEqual(a.currentState.Monitors, m) {
					a.currentState.Monitors = m
					slog.Info("monitors state updated", "state", a.currentState.Monitors)
				}

			case lidSwitchEvent:
				a.currentState.LidState = parseLidState(ev.Details)
				slog.Info("lid state updated", "state", a.currentState.LidState)

			case powerChangedEvent:
				a.currentState.PowerState = parsePowerState(ev.Details)
				slog.Info("power state updated", "state", a.currentState.PowerState)

			case configUpdatedEvent:
				// Update config values
				err := a.cfg.reload(5)
				if err != nil {
					slog.Error("reloading config", "error", err)
					continue
				}
				slog.Info("profiles reloaded", "count", len(a.cfg.Profiles))
				a.validateAllProfiles()
			}

			if !a.currentState.ready() {
				continue
			}

			if err := a.update(); err != nil {
				slog.Error("running updater", "error", err)
			}

		case err := <-errc:
			return fmt.Errorf("listener failed: %w", err)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (a *app) update() error {
	lookup := a.newLabelLookup()
	matched := a.getMatchingProfile(lookup)
	if matched == nil {
		slog.Info("no match found")
		return nil
	}
	slog.Info("found profile match", "profile", matched.Name)

	params := a.updateParamsFromProfile(matched, lookup)
	for _, u := range params.enable {
		slog.Info("will update monitor", "name", u.Name, "desc", u.Description, "make", u.Make, "model", u.Model)
	}

	for _, d := range params.disable {
		slog.Info("will disable monitor", "name", d.Name, "desc", d.Description, "make", d.Make, "model", d.Model)
	}
	return nil
}

func (a *app) updateParamsFromProfile(p *profile, lookup labelLookup) *monitorUpdateParams {
	var toUpdate, toDisable []monitor
	seenNames := make(map[string]bool)

	for _, s := range p.MonitorStates {
		lm, ok := lookup[s.Label]
		if !ok {
			// only warn if not disabled; if it's marked disabled but missing,
			// mission is already accomplished
			if !s.Disabled {
				slog.Warn("update param builder: no monitor found for label", "label", s.Label)
				continue
			}
			slog.Debug("update param builder: no monitor found for label, but marked for disable, no action needed", "label", s.Label)
			continue
		}

		seenNames[lm.Monitor.Name] = true

		if s.Disabled {
			toDisable = append(toDisable, lm.Monitor)
			continue
		}

		if s.Preset == nil {
			// no preset provided, just assume it stays the same
			toUpdate = append(toUpdate, lm.Monitor)
			continue
		}

		pr, ok := a.cfg.Monitors[s.Label].Presets[*s.Preset]
		if !ok {
			slog.Warn("update param builder: provided preset doesn't exist", "preset", s.Preset)
			continue
		}

		m := monitor{
			monitorIdentifiers: lm.Monitor.monitorIdentifiers,
			monitorSettings:    pr,
		}

		toUpdate = append(toUpdate, m)
	}

	if p.DisableUndeclared {
		for _, m := range a.currentState.Monitors {
			if !seenNames[m.Name] {
				toDisable = append(toDisable, m)
			}
		}
	}

	return newMonitorUpdateParams(toUpdate, toDisable)
}
