package hypr

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	runtimeEnv = "XDG_RUNTIME_DIR"
	sigEnv     = "HYPRLAND_INSTANCE_SIGNATURE"
	sockName   = ".socket2.sock"
)

var ErrMissingEnvs = errors.New("missing hyprland envs")

type SocketConn struct {
	*net.UnixConn
}

type baseEvent struct {
	name    string
	payload string
}

func NewSocketConn() (*SocketConn, error) {
	runtime := os.Getenv(runtimeEnv)
	sig := os.Getenv(sigEnv)
	if runtime == "" || sig == "" {
		return nil, ErrMissingEnvs
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

func (h *SocketConn) ListenForEvents() error {
	scn := bufio.NewScanner(h)
	for scn.Scan() {
		line := scn.Text()
		if err := handleLine(line); err != nil {
			fmt.Printf("Error handline line %s: %v\n", line, err)
		}
	}

	if err := scn.Err(); err != nil {
		return fmt.Errorf("error scanning: %w", err)
	}

	return nil
}

func handleLine(line string) error {
	event, err := parseBaseEvent(line)
	if err != nil {
		return fmt.Errorf("parsing event: %w", err)
	}
	slog.Debug("event received", "name", event.name, "payload", event.payload)
	switch event.name {
	case "monitoraddedv1":
		n, err := extractMonitorName(event.payload)
		if err != nil {
			logExtractErr(err)
		}
		slog.Debug("got monitor added event", "monitor_name", n)
	case "monitorremovedv1":
		n, err := extractMonitorName(event.payload)
		if err != nil {
			logExtractErr(err)
		}
		slog.Debug("got monitor removed event", "monitor_name", n)
	}

	return nil
}

func logExtractErr(err error) {
	slog.Error("extracting monitor name from event", "error", err)
}

func extractMonitorName(payload string) (string, error) {
	parts := strings.Split(payload, ",")
	if len(parts) != 3 {
		return "", fmt.Errorf("bad monitorv2 event: %q", payload)
	}

	return parts[1], nil
}

func parseBaseEvent(line string) (*baseEvent, error) {
	parts := strings.SplitN(line, ">>", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid event: %q", line)
	}

	return &baseEvent{
		name:    parts[0],
		payload: parts[1],
	}, nil
}
