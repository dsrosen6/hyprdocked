package hypr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	binaryName       = "hyprctl"
	unknownReqOutput = "unknown request"
	runtimeEnv       = "XDG_RUNTIME_DIR"
	sigEnv           = "HYPRLAND_INSTANCE_SIGNATURE"
	sockName         = ".socket2.sock"
)

var (
	errUnknownRequest = errors.New(unknownReqOutput)
	errMissingEnvs    = errors.New("missing hyprland envs")
)

type (
	Client struct {
		binaryPath string
	}

	Monitor struct {
		Name        string  `json:"name,omitempty"`
		Description string  `json:"description,omitempty"`
		Width       int64   `json:"width,omitempty"`
		Height      int64   `json:"height,omitempty"`
		RefreshRate float64 `json:"refreshRate,omitempty"`
		X           int64   `json:"x,omitempty"`
		Y           int64   `json:"y,omitempty"`
		Scale       float64 `json:"scale,omitempty"`
	}

	SocketConn struct {
		*net.UnixConn
	}
)

func NewClient() (*Client, error) {
	bp, err := exec.LookPath(binaryName)
	if err != nil {
		return nil, fmt.Errorf("finding full hyprctl binary path: %w", err)
	}

	return &Client{binaryPath: bp}, nil
}

func WaitForEnvs() {
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

func NewSocketConn() (*SocketConn, error) {
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

	return &SocketConn{conn}, nil
}

func (h *Client) RunCmdUnmarshal(args []string, v any) error {
	a := append([]string{"-j"}, args...)
	out, err := h.RunCmd(a)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(out, v); err != nil {
		return fmt.Errorf("unmarshaling json: %w", err)
	}

	return nil
}

func (h *Client) RunCmd(args []string) ([]byte, error) {
	cmd := exec.Command(h.binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("running command: %w", err)
	}

	out := stdout.Bytes()
	errStr := strings.TrimSpace(stderr.String())
	if errStr != "" {
		return nil, errors.New(errStr)
	}

	return out, checkForErr(string(out))
}

func (h *Client) Reload() error {
	if _, err := h.RunCmd([]string{"reload"}); err != nil {
		return err
	}
	return nil
}

func (h *Client) ListMonitors() ([]Monitor, error) {
	var displays []Monitor
	if err := h.RunCmdUnmarshal([]string{"monitors"}, &displays); err != nil {
		return nil, err
	}

	return displays, nil
}

func (h *Client) EnableOrUpdateMonitor(m Monitor) error {
	args := []string{"keyword", "monitor", MonitorToConfigString(m)}
	if _, err := h.RunCmd(args); err != nil {
		return err
	}

	return nil
}

func (h *Client) DisableMonitor(m Monitor) error {
	args := []string{"keyword", "monitor", m.Name + ",", "disable"}
	if _, err := h.RunCmd(args); err != nil {
		return err
	}

	return nil
}

func MonitorToConfigString(m Monitor) string {
	res := fmt.Sprintf("%dx%d", m.Width, m.Height)
	res = fmt.Sprintf("%s@%f", res, m.RefreshRate)
	xy := fmt.Sprintf("%dx%d", m.X, m.Y)
	scale := fmt.Sprintf("%f", m.Scale)
	return fmt.Sprintf("%s,%s,%s,%s", m.Name, res, xy, scale)
}

func checkForErr(out string) error {
	out = strings.TrimSpace(out)
	switch out {
	case unknownReqOutput:
		return errUnknownRequest
	default:
		return nil
	}
}
