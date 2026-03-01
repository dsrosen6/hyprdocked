package app

import (
	"log/slog"
	"os/exec"
	"time"
)

func (a *App) runUpdater() error {
	a.updating = true
	defer func() {
		a.lastUpdateEnd = time.Now()
		a.updating = false
	}()

	if a.mode == modeSuspending {
		slog.Info("[UPDATER]enabling laptop display before suspend")
		if err := a.hctl.enableOrUpdateDisplay(a.laptopDisplay); err != nil {
			slog.Error("issue enabling laptop display for suspend command; continuing with suspend", "error", err)
		}
		return systemctlSuspend()
	}

	s := a.getStatus()
	lg := slog.Default().With(
		slog.String("mode", a.mode.string()),
		slog.String("status", s.string()),
	)

	switch s {
	case statusDockedOpened, statusOnlyLaptopOpened:
		switch a.laptopIsEnabled() {
		case true:
			lg.Debug("[UPDATER]laptop display already enabled; no action needed")
		case false:
			lg.Info("[UPDATER]enabling laptop display")
			return a.hctl.enableOrUpdateDisplay(a.laptopDisplay)
		}

	case statusOnlyLaptopClosed:
		switch a.laptopIsEnabled() {
		case true:
			lg.Debug("[UPDATER]laptop display already enabled; no display action needed")
		case false:
			lg.Info("[UPDATER]enabling laptop display")
			if err := a.hctl.enableOrUpdateDisplay(a.laptopDisplay); err != nil {
				lg.Error("[UPDATER]issue enabling laptop display", "error", err)
			}
		}

		if a.suspendOnClosed {
			lg.Info("[UPDATER]suspending machine")
			return systemctlSuspend()
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

func systemctlSuspend() error {
	cmd := exec.Command("systemctl", "suspend")
	return cmd.Run()
}
