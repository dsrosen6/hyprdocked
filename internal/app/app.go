package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

const version = "0.2.0"

type app struct {
	hctl          *hyprClient
	listener      *listener
	updating      bool
	lastUpdateEnd time.Time
	*state
}

func newApp(hc *hyprClient, l *listener, s *state) *app {
	return &app{
		hctl:     hc,
		listener: l,
		state:    s,
	}
}

func Run() error {
	if strings.ToLower(os.Getenv("DEBUG")) == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "version":
			fmt.Println(version)
			return nil

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
			return fmt.Errorf("unknown command: %s", os.Args[1])
		}
	}

	return runListener()
}

func runListener() error {
	waitForHyprEnvs()

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

	lh := newLidHandler(dbusConn)
	ph := newPowerHandler(dbusConn)
	p := listenerParams{
		hyprSockConn: hyprSock,
		lidHandler:   lh,
		powerHandler: ph,
		dbusConn:     dbusConn,
	}

	l, err := newListener(p)
	if err != nil {
		return fmt.Errorf("creating listener: %w", err)
	}

	s, err := getInitialState(context.Background(), hyprClient, lh, ph)
	if err != nil {
		return fmt.Errorf("getting initial state: %w", err)
	}

	app := newApp(hyprClient, l, s)
	// initial updater run before starting listener
	_ = app.runUpdater()

	return app.listenAndHandle(context.Background())
}
