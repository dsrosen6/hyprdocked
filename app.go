package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/godbus/dbus/v5"
)

type app struct {
	hctl          *hyprClient
	cfg           *config
	listener      *listener
	currentState  *state
	updating      bool
	lastUpdateEnd time.Time
}

func newApp(cfg *config, hc *hyprClient, l *listener, s *state) *app {
	return &app{
		hctl:         hc,
		cfg:          cfg,
		listener:     l,
		currentState: s,
	}
}

func run() error {
	if os.Getenv("DEBUG") == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "suspend":
			if err := sendSuspendCmd(); err != nil {
				return fmt.Errorf("sending suspend command: %w", err)
			}
			time.Sleep(2 * time.Second) // give listener time before idle agent actually suspends
			return nil
		case "wake":
			if err := sendWakeCmd(); err != nil {
				return fmt.Errorf("sending wake command: %w", err)
			}
			return nil
		default:
			fmt.Printf("unknown command: %s\n", os.Args[1])
		}
	}

	cfg, err := initConfig("")
	if err != nil {
		return fmt.Errorf("initializing config: %w", err)
	}

	hyprClient, err := newHyprctlClient()
	if err != nil {
		return fmt.Errorf("creating hyprctl client: %w", err)
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

	lh := newLidHandler(dbusConn)
	ph := newPowerHandler(dbusConn)
	p := listenerParams{
		hyprSockConn: hyprSock,
		lidHandler:   lh,
		powerHandler: ph,
		dbusConn:     dbusConn,
		cfgPath:      cfg.path,
	}

	l, err := newListener(p)
	if err != nil {
		return fmt.Errorf("creating listener: %w", err)
	}

	s, err := getInitialState(context.Background(), hyprClient, lh, ph)
	if err != nil {
		return fmt.Errorf("getting initial state: %w", err)
	}

	app := newApp(cfg, hyprClient, l, s)
	// initial updater run before starting listener
	_ = app.runUpdater()

	return app.listenAndHandle(context.Background())
}
