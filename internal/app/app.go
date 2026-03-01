package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dsrosen6/hyprdocked/internal/power"
	"github.com/godbus/dbus/v5"
)

type App struct {
	hctl            *hyprClient
	listener        *listener
	updating        bool
	lastUpdateEnd   time.Time
	suspendOnIdle   bool
	suspendOnClosed bool
	*state
}

type ListenerParams struct {
	LaptopMonitorName string
	SuspendOnIdle     bool
	SuspendOnClosed   bool
}

func newApp(hc *hyprClient, l *listener, s *state, suspendIdle, suspendClosed bool) *App {
	return &App{
		hctl:            hc,
		listener:        l,
		state:           s,
		suspendOnIdle:   suspendIdle,
		suspendOnClosed: suspendClosed,
	}
}

func RunListener(p ListenerParams) error {
	waitForHyprEnvs()
	if p.LaptopMonitorName == "" {
		return errors.New("laptop monitor name cannot be empty")
	}

	hyprClient, err := newHyprctlClient()
	if err != nil {
		return fmt.Errorf("creating hyprctl client: %w", err)
	}

	// Run an initial reload in case laptop display is already disabled. Assuming the laptop
	// display is correctly set to initially enable in the hyprland config, this will re-enable
	// it so hyprdocked can properly identify it.
	slog.Info("running hyprctl reload")
	if err := hyprClient.reload(); err != nil {
		return fmt.Errorf("running hyprctl reload: %w", err)
	}

	var (
		hyprSock *hyprSocketConn
		dbusConn *dbus.Conn
	)

	defer func() {
		if hyprSock != nil {
			if err := hyprSock.Close(); err != nil {
				slog.Error("closing hypr socket connection", "error", err)
			}
		}

		if dbusConn != nil {
			if err := dbusConn.Close(); err != nil {
				slog.Error("closing dbus connection", "error", err)
			}
		}
	}()
	hyprSock, err = newHyprSocketConn()
	if err != nil {
		return fmt.Errorf("creating hyprland socket connection: %w", err)
	}

	dbusConn, err = dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("creating dbus connection: %w", err)
	}

	lh := power.NewLidHandler(dbusConn)
	lp := listenerParams{
		hyprSockConn: hyprSock,
		lidHandler:   lh,
		dbusConn:     dbusConn,
	}

	l, err := newListener(lp)
	if err != nil {
		return fmt.Errorf("creating listener: %w", err)
	}

	sp := initialStateParams{
		laptopMonitorName: p.LaptopMonitorName,
		hyprClient:        hyprClient,
		lidHandler:        lh,
	}

	s, err := getInitialState(context.Background(), sp)
	if err != nil {
		return fmt.Errorf("getting initial state: %w", err)
	}

	app := newApp(hyprClient, l, s, p.SuspendOnIdle, p.SuspendOnClosed)
	slog.Info("app initialized",
		"laptop_monitor_name", app.laptopDisplay.Name,
		"status", app.statusString(),
		"suspend_idle", app.suspendOnIdle,
		"suspend_closed", app.suspendOnClosed,
	)

	// initial updater run before starting listener
	_ = app.runUpdater()

	return app.listenAndHandle(context.Background())
}
