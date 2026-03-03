package app

import (
	"log/slog"
	"os/exec"
)

func (a *App) runUpdater() error {
	a.updating = true
	defer func() {
		a.updating = false
	}()

	if a.mode == modeIdle {
		return a.handleIdleCmd()
	}

	s := a.status()
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

		if a.Config.SuspendClosed {
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

func (a *App) handleIdleCmd() error {
	if !a.laptopIsEnabled() {
		slog.Info("[UPDATER/IDLE CMD]enabling laptop display")
		if err := a.hctl.enableOrUpdateDisplay(a.laptopDisplay); err != nil {
			slog.Error("[UPDATER/IDLE CMD]issue enabling laptop display", "error", err)
		}
	} else {
		slog.Info("[UPDATER/IDLE CMD]laptop display already enabled")
	}

	if a.Config.SuspendIdle {
		slog.Info("[UPDATER/IDLE CMD]suspending on idle enabled; suspending")
		return systemctlSuspend()
	}
	slog.Info("[UPDATER/IDLE CMD]suspending on idle disabled; doing nothing")
	return nil
}

func systemctlSuspend() error {
	cmd := exec.Command("systemctl", "suspend")
	return cmd.Run()
}
