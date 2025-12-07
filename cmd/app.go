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
	saveDiplaysCmd = flag.NewFlagSet("save-displays", flag.ExitOnError)
	mtrName        = saveDiplaysCmd.String("laptop", "", "name of laptop display")
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
	case "save-displays", "sd":
		return handleSaveDisplays(a, args)
	case "listen":
		return handleListen(a)
	default:
		return errors.New("invalid command")
	}
}

func handleSaveDisplays(a *app.App, args []string) error {
	expectedArgs := 1
	gotArgs := len(args) - 1
	if gotArgs != expectedArgs {
		return fmt.Errorf("expected %d arguments, got %d", expectedArgs, gotArgs)
	}

	if err := saveDiplaysCmd.Parse(args[1:]); err != nil {
		return fmt.Errorf("parsing arguments: %w", err)
	}

	if err := a.SaveCurrentDisplays(*mtrName); err != nil {
		return fmt.Errorf("setting laptop display: %w", err)
	}

	fmt.Printf("Laptop display '%s' saved to config.\n", a.Cfg.LaptopDisplay.Name)
	externals := a.Cfg.ExternalDisplays
	switch len(externals) {
	case 0:
		fmt.Println("No external display detected.")
	default:
		fmt.Println("Saved external display(s):")
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
