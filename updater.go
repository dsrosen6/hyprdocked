package main

import (
	"fmt"
	"log/slog"
	"time"
)

func (a *app) runUpdater() error {
	a.updating = true
	defer func() {
		a.lastUpdateEnd = time.Now()
		a.updating = false
	}()

	st := a.getStatus()

	var p *monitorUpdateParams
	switch a.currentState.mode {
	case modeSuspending:
		p = a.suspendingUpdateParams()
	case modeWaking:
		p = a.wakingUpdateParams()
	default:
		p = a.createUpdateParams(st)
	}
	logUpdateParams(*p)

	if len(p.enableOrUpdate) == 0 && len(p.disable) == 0 {
		slog.Info("no changes needed", "status", st.string())
		return nil
	}

	slog.Info("applying updates", "status", st.string())
	if err := a.hctl.bulkUpdateMonitors(p); err != nil {
		return fmt.Errorf("bulk updating monitors: %w", err)
	}

	m, err := a.hctl.listMonitors()
	if err != nil {
		return fmt.Errorf("listing monitors post-update: %w", err)
	}
	a.currentState.monitors = m

	return nil
}

func (a *app) createUpdateParams(st status) *monitorUpdateParams {
	logger := slog.Default().With(slog.String("status", st.string()))
	var toUpdate, toDisable, noChanges []monitor
	switch st {
	case statusOnlyLaptopOpened, statusOnlyLaptopClosed:
		// we still want to enable the laptop display even if it's closed. This is so
		// we don't get an "oopsie daisy" from hyprland when waking.
		lg := logger.With(monitorLogGroup("laptop", a.cfg.Laptop))
		m := a.currentState.getMonitorByIdentifiers(a.cfg.Laptop.monitorIdentifiers)
		if m == nil || changesNeeded(a.cfg.Laptop, *m) {
			lg.Debug("updater: laptop monitor updates needed")
			toUpdate = append(toUpdate, a.cfg.Laptop)
		} else {
			lg.Debug("updater: no laptop monitor changes needed")
			noChanges = append(noChanges, a.cfg.Laptop)
		}

	case statusDockedClosed, statusDockedOpened:
		for i, cm := range a.cfg.Monitors {
			n := fmt.Sprintf("external%d", i)
			lg := logger.With(monitorLogGroup(fmt.Sprintf("cfg_%s", n), cm))
			m := a.currentState.getMonitorByIdentifiers(cm.monitorIdentifiers)
			if m == nil {
				continue
			}
			lg = lg.With(monitorLogGroup(fmt.Sprintf("state_%s", n), *m))

			if changesNeeded(cm, *m) {
				lg.Debug("updater: changes needed for monitor")
				toUpdate = append(toUpdate, cm)
				continue
			}
			lg.Debug("updater: no changes needed for monitor")
			noChanges = append(noChanges, cm)
		}

		lg := logger.With(monitorLogGroup("cfg_laptop", a.cfg.Laptop))
		if st == statusDockedClosed {
			if a.currentState.getMonitorByIdentifiers(a.cfg.Laptop.monitorIdentifiers) != nil {
				lg.Debug("updater: laptop monitor needs disabled")
				toDisable = append(toDisable, a.cfg.Laptop)
			}
		} else {
			m := a.currentState.getMonitorByIdentifiers(a.cfg.Laptop.monitorIdentifiers)
			if m == nil || changesNeeded(a.cfg.Laptop, *m) {
				lg.Debug("updater: laptop monitor updates needed")
				toUpdate = append(toUpdate, a.cfg.Laptop)
			} else {
				lg.Debug("updater: no laptop monitor updates needed")
				noChanges = append(noChanges, a.cfg.Laptop)
			}
		}

	}

	return newMonitorUpdateParams(toUpdate, toDisable, noChanges)
}

func (a *app) suspendingUpdateParams() *monitorUpdateParams {
	var toUpdate, toDisable, noChanges []monitor
	// add ALL externals to disable group, and also to suspended monitors in state memory
	for _, m := range a.currentState.monitors {
		if !matchesIdentifiers(m, a.cfg.Laptop.monitorIdentifiers) {
			toDisable = append(toDisable, m)
			a.currentState.suspendedMonitors = append(a.currentState.suspendedMonitors, m)
			slog.Debug("updater: suspending, adding external monitor to disable and suspended lists", monitorLogGroup(m.Name, m))
		}
	}
	toUpdate = append(toUpdate, a.cfg.Laptop)
	return newMonitorUpdateParams(toUpdate, toDisable, noChanges)
}

func (a *app) wakingUpdateParams() *monitorUpdateParams {
	var toUpdate, toDisable, noChanges []monitor
	for _, m := range a.currentState.suspendedMonitors {
		toUpdate = append(toUpdate, m)
		slog.Debug("updater: waking, adding previously suspended monitor to enable list", monitorLogGroup(m.Name, m))
	}

	return newMonitorUpdateParams(toUpdate, toDisable, noChanges)
}

func changesNeeded(cfg, state monitor) bool {
	lg := slog.Default().With(slog.String("monitor_name", cfg.Name))
	changes := false
	if cfg.Width != state.Width {
		changes = true
		lg.Debug("monitor change detected: width", slog.Int64("cfg", cfg.Width), slog.Int64("state", state.Width))
	}

	if cfg.Height != state.Height {
		changes = true
		lg.Debug("monitor change detected: height", slog.Int64("cfg", cfg.Height), slog.Int64("state", state.Height))
	}

	if cfg.RefreshRate != state.RefreshRate {
		changes = true
		lg.Debug("monitor change detected: refresh rate", slog.Float64("cfg", cfg.RefreshRate), slog.Float64("state", state.RefreshRate))
	}

	if cfg.Scale != state.Scale {
		changes = true
		lg.Debug("monitor change detected: scale", slog.Float64("cfg", cfg.Scale), slog.Float64("state", state.Scale))
	}

	cfgPos := fmt.Sprintf("%dx%d", cfg.X, cfg.Y)
	stPos := fmt.Sprintf("%dx%d", state.X, state.Y)
	if cfgPos != stPos {
		changes = true
		lg.Debug("monitor change detected: position", slog.String("cfg", cfgPos), slog.String("state", stPos))
	}

	return changes
}

func logUpdateParams(params monitorUpdateParams) {
	for _, m := range params.enableOrUpdate {
		slog.Debug("will update monitor", "name", m.Name, "desc", m.Description)
	}

	for _, m := range params.disable {
		slog.Debug("will disable monitor", "name", m.Name, "desc", m.Description)
	}

	for _, m := range params.noChanges {
		slog.Debug("no changes needed for monitor", "name", m.Name, "desc", m.Description)
	}
}
