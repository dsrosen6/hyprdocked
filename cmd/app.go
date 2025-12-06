package cmd

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

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

	cfg, err := config.InitConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	hc, err := hypr.NewHyprctlClient()
	if err != nil {
		return fmt.Errorf("creating hyprctl client: %w", err)
	}

	a := app.NewApp(cfg, hc)
	if err := handleCommands(a); err != nil {
		return err
	}

	return nil
}

func handleCommands(a *app.App) error {
	args := os.Args[1:]
	if len(args) == 0 {
		return errors.New("no subcommand provided")
	}

	switch args[0] {
	case "save-monitors", "sm":
		return handleSaveMonitors(a, args)
	case "listen":
		return handleListen(a)
	default:
		return errors.New("invalid command")
	}
}

func handleSaveMonitors(a *app.App, args []string) error {
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

	fmt.Printf("Laptop monitor '%s' saved to config.\n", a.Cfg.LaptopMonitor.Name)
	externals := a.Cfg.ExternalMonitors
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

func handleListen(a *app.App) error {
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
