// Package cmd is the entry point for hyprlaptop
package cmd

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/dsrosen6/hyprlaptop/internal/app"
	"github.com/dsrosen6/hyprlaptop/internal/config"
	"github.com/dsrosen6/hyprlaptop/internal/hypr"
)

var (
	saveMtrsCmd = flag.NewFlagSet("save-monitors", flag.ExitOnError)
	mtrName     = saveMtrsCmd.String("laptop", "", "name of laptop monitor")
)

func Run() error {
	if err := parseFlags(); err != nil {
		return fmt.Errorf("parsing cli flags: %w", err)
	}

	cfg, err := config.ReadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	hc, err := hypr.NewHyprctlClient()
	if err != nil {
		return fmt.Errorf("creating hyprctl client: %w", err)
	}

	a := app.NewApp(cfg, hc)

	return nil
}

func handleCommands(a *app.App) error {
	args := os.Args
	if len(args) < 2 {
		return errors.New("no subcommand provided")
	}

	switch args[1] {
	case "save-monitors", "sm":
		// do thing
		return nil
	case "listen":
		// do thing
		return nil
	default:
		return errors.New("invalid command")
	}
}

func handleSaveMonitors(a *app.App, args []string) error {
}

func (a *App) HandleSaveMonitors(args []string) error {
	expectedArgs := 1
	gotArgs := len(args) - 1
	if gotArgs != expectedArgs {
		return fmt.Errorf("expected %d arguments, got %d", expectedArgs, gotArgs)
	}

	if err := saveMtrsCmd.Parse(args[1:]); err != nil {
		return fmt.Errorf("parsing arguments: %w", err)
	}

	if err := a.SaveCurrentMonitors(*mtrName); err != nil {
		return fmt.Errorf("setting laptop monitor: %w", err)
	}

	fmt.Printf("Laptop monitor '%s' saved to config.\n", a.config.LaptopMonitor.Name)
	externals := a.config.ExternalMonitors
	switch len(externals) {
	case 0:
		fmt.Println("No external monitors detected.")
	default:
		fmt.Println("Saved external monitor(s):")
		for _, e := range externals {
			fmt.Printf("	%s\n", e.Name)
		}
	}

	return nil
}

func (a *App) HandleListen() error {
	slog.Info("initializing socket connection")
	sc, err := hypr.NewSocketConn()
	if err != nil {
		return err
	}
	defer func() {
		if err := sc.Close(); err != nil {
			slog.Error("error closing socket connection", "error", err)
			return
		}
		slog.Info("socket connection closed")
	}()

	slog.Info("listening for hyprland events")
	if err := sc.ListenForEvents(); err != nil {
		return err
	}

	return nil
}

func (a *App) SaveCurrentMonitors(laptop string) error {
	monitors, err := a.hctl.ListMonitors()
	if err != nil {
		return fmt.Errorf("listing monitors: %w", err)
	}

	var lm *hypr.Monitor
	if laptop == "" {
		for _, m := range monitors {
			if strings.Contains(m.Name, "eDP") {
				lm = &m
			}
		}
	} else {
		l, ok := monitors[laptop]
		if ok {
			lm = &l
		}
	}

	if lm == nil {
		return fmt.Errorf("monitor '%s' not found", laptop)
	}

	externals := map[string]hypr.Monitor{}
	for _, m := range monitors {
		if m.Name != lm.Name {
			externals[m.Name] = m
		}
	}

	a.config.LaptopMonitor = *lm
	a.config.ExternalMonitors = externals

	if err := a.config.Write(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
