package main

import (
	"log/slog"
	"time"
)

func (a *app) runUpdater() error {
	a.updating = true
	defer func() {
		a.lastUpdateEnd = time.Now()
		a.updating = false
	}()

	if a.mode == modeSuspending {
		slog.Info("[UPDATER]suspend command received; enabling laptop display")
		return a.hctl.enableOrUpdateDisplay(a.laptopDisplay)
	}

	s := a.getStatus()
	switch s {
	case statusDockedOpened, statusOnlyLaptopOpened, statusOnlyLaptopClosed:
		slog.Info("[UPDATER]enabling laptop display if not already enabled", "status", s.string())
		return a.hctl.enableOrUpdateDisplay(a.laptopDisplay)
	case statusDockedClosed:
		slog.Info("[UPDATER]disabling laptop display if not already disabled", "status", s.string())
		return a.hctl.disableDisplay(a.laptopDisplay)
	default:
		slog.Info("[UPDATER]unknown status; doing nothing", "status", s.string())
	}

	return nil
}
