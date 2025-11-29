package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/dsrosen6/hyprlaptop/internal/cfg"
	"github.com/dsrosen6/hyprlaptop/internal/cli"
	"github.com/dsrosen6/hyprlaptop/internal/hypr"
)

func main() {
	panic(run())
}

func run() error {
	opts, err := cli.Parse()
	if err != nil {
		return fmt.Errorf("parsing cli: %w", err)
	}

	c, err := cfg.ReadConfig(opts.ConfigFile)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}
	slog.Info("config loaded", "primary_monitor", c.PrimaryMonitorName)

	if os.Getenv("DEBUG") == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	conn, err := hypr.NewConn()
	if err != nil {
		return err
	}

	if err := conn.Listen(); err != nil {
		return err
	}

	return nil
}
