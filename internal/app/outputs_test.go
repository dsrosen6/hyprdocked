package app

import (
	"fmt"
	"testing"

	"github.com/dsrosen6/hyprlaptop/internal/config"
	"github.com/dsrosen6/hyprlaptop/internal/hypr"
)

func TestRun(t *testing.T) {
	a, err := newTestApp(t)
	if err != nil {
		t.Fatalf("creating app: %v", err)
	}

	if err := a.Run(); err != nil {
		t.Fatal(err)
	}
}

func newTestApp(t *testing.T) (*App, error) {
	t.Helper()
	cfg, err := config.InitConfig("")
	if err != nil {
		return nil, fmt.Errorf("initializing config: %w", err)
	}

	hc, err := hypr.NewHyprctlClient()
	if err != nil {
		return nil, fmt.Errorf("creating hyprctl client: %w", err)
	}

	return NewApp(cfg, hc), nil
}
