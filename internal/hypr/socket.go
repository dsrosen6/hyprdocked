package hypr

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
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
