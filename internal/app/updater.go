package app

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
	lg := slog.Default().With(
		slog.String("mode", a.mode.string()),
		slog.String("status", s.string()),
	)

	switch s {
	case statusDockedOpened, statusOnlyLaptopOpened, statusOnlyLaptopClosed:
		switch a.laptopIsEnabled() {
		case true:
			lg.Debug("[UPDATER]laptop display already enabled; no action needed")
		case false:
			lg.Info("[UPDATER]enabling laptop display")
			return a.hctl.enableOrUpdateDisplay(a.laptopDisplay)
		}
	case statusDockedClosed:
		switch a.laptopIsEnabled() {
		case true:
			lg.Info("[UPDATER]disabling laptop display")
			return a.hctl.disableDisplay(a.laptopDisplay)
		case false:
			lg.Debug("[UPDATER]laptop display already disabled; no action needed")
		}
	default:
		lg.Info("[UPDATER]unknown status; doing nothing")
	}

	return nil
}
