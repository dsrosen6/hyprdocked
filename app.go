package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/huh"
)

type app struct {
	hc      *hyprctlClient
	cfg     *config
	cfgPath string
}

func newApp() (*app, error) {
	if os.Getenv("DEBUG") == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	opts, err := parseFlags()
	if err != nil {
		return nil, fmt.Errorf("parsing cli flags: %w", err)
	}

	cfg, err := readConfig(opts.configFile)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	hc, err := newHctlClient()
	if err != nil {
		return nil, fmt.Errorf("creating hyprctl client: %w", err)
	}

	return &app{
		hc:      hc,
		cfg:     cfg,
		cfgPath: opts.configFile,
	}, nil
}

func (a *app) run() error {
	args := os.Args
	if len(args) < 2 {
		return errors.New("no subcommand provided")
	}
	switch args[1] {
	case "select-monitor":
		m, err := a.hc.pickMonitor()
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}
		a.cfg.LaptopMonitor = *m
		if err := a.cfg.write(a.cfgPath); err != nil {
			return fmt.Errorf("writing to config file: %w", err)
		}
		fmt.Printf("Laptop monitor '%s' saved to config\n", a.cfg.LaptopMonitor.Name)

	case "listen":
		sc, err := newSocketConn()
		if err != nil {
			return fmt.Errorf("initializing socket connection: %w", err)
		}
		defer func() {
			if err := sc.Close(); err != nil {
				slog.Error("closing socket connection", "error", err)
			}
		}()

		slog.Info("listening for hyprland events")
		if err := a.listen(sc); err != nil {
			return err
		}
	default:
		return errors.New("invalid command")
	}
	return nil
}
