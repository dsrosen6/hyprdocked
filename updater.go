package main

import (
	"fmt"
	"log/slog"
	"reflect"
	"time"
)

func (a *app) runUpdater() error {
	a.updating = true
	defer func() {
		a.lastUpdateEnd = time.Now()
		a.updating = false
	}()

	st := a.getStatus()
	params := a.createUpdateParams(st)
	logUpdateParams(*params)
	if len(params.enableOrUpdate) == 0 && len(params.disable) == 0 {
		slog.Info("no changes needed", "status", st.string())
		return nil
	}

	slog.Info("applying updates", "status", st.string())
	if err := a.hctl.bulkUpdateMonitors(params); err != nil {
		return fmt.Errorf("bulk updating monitors: %w", err)
	}

	return nil
}

func (a *app) createUpdateParams(st status) *monitorUpdateParams {
	var toUpdate, toDisable, noChanges []monitor

	switch st {
	case statusOnlyLaptopOpened:
		m := a.currentState.getMonitorByIdentifiers(a.cfg.Laptop.monitorIdentifiers)
		if m == nil || (m != nil && changesNeeded(a.cfg.Laptop, *m)) {
			toUpdate = append(toUpdate, a.cfg.Laptop)
		} else {
			noChanges = append(noChanges, a.cfg.Laptop)
		}
	case statusDockedClosed, statusDockedOpened:
		for _, cm := range a.cfg.Monitors {
			m := a.currentState.getMonitorByIdentifiers(cm.monitorIdentifiers)
			if m == nil {
				continue
			}

			if changesNeeded(cm, *m) {
				toUpdate = append(toUpdate, cm)
				continue
			}
			noChanges = append(noChanges, cm)
		}

		if st == statusDockedClosed {
			if a.currentState.getMonitorByIdentifiers(a.cfg.Laptop.monitorIdentifiers) != nil {
				toDisable = append(toDisable, a.cfg.Laptop)
			}
		} else {
			m := a.currentState.getMonitorByIdentifiers(a.cfg.Laptop.monitorIdentifiers)
			if m == nil || (m != nil && changesNeeded(a.cfg.Laptop, *m)) {
				toUpdate = append(toUpdate, a.cfg.Laptop)
			} else {
				noChanges = append(noChanges, a.cfg.Laptop)
			}
		}
	}

	return newMonitorUpdateParams(toUpdate, toDisable, noChanges)
}

func changesNeeded(a, b monitor) bool {
	return !reflect.DeepEqual(a.monitorSettings, b.monitorSettings)
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
