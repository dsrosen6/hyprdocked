package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	hypr "github.com/dsrosen6/hyprland-go"
)

const (
	runtimeEnv = "XDG_RUNTIME_DIR"
	sigEnv     = "HYPRLAND_INSTANCE_SIGNATURE"
	sockName   = ".socket2.sock"
)

var errMissingEnvs = errors.New("missing hyprland envs")

type (
	hyprSocketConn struct {
		*net.UnixConn
	}
)

func waitForHyprEnvs() {
	ready := func() bool {
		runtime := os.Getenv(runtimeEnv)
		sig := os.Getenv(sigEnv)
		return runtime != "" && sig != ""
	}

	for !ready() {
		slog.Info("hyprland env not yet loaded; waiting 1s")
		time.Sleep(1 * time.Second)
	}
	slog.Info("hyprland envs loaded")
}

func disableMonitor(hc *hypr.Client, m hypr.Monitor) error {
	return hc.RunKeywordCmd("monitor", fmt.Sprintf("%s,disable", m.Name))
}

func enableOrUpdateDisplay(hc *hypr.Client, m hypr.Monitor) error {
	return hc.RunKeywordCmd("monitor", displayToConfigString(m))
}

func newHyprSocketConn() (*hyprSocketConn, error) {
	runtime := os.Getenv(runtimeEnv)
	sig := os.Getenv(sigEnv)
	if runtime == "" || sig == "" {
		return nil, errMissingEnvs
	}

	sock := filepath.Join(runtime, "hypr", sig, sockName)
	addr := &net.UnixAddr{
		Name: sock,
		Net:  "unix",
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("connecting to socket: %w", err)
	}

	return &hyprSocketConn{conn}, nil
}

func displayToConfigString(m hypr.Monitor) string {
	res := fmt.Sprintf("%dx%d", m.Width, m.Height)
	res = fmt.Sprintf("%s@%f", res, m.RefreshRate)
	xy := fmt.Sprintf("%dx%d", m.X, m.Y)
	scale := fmt.Sprintf("%f", m.Scale)
	return fmt.Sprintf("%s,%s,%s,%s", m.Name, res, xy, scale)
}
