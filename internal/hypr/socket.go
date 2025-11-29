package hypr

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	runtimeEnv = "XDG_RUNTIME_DIR"
	sigEnv     = "HYPRLAND_INSTANCE_SIGNATURE"
)

var ErrMissingEnvs = errors.New("missing hyprland envs")

type Conn struct {
	*net.UnixConn
}

func NewConn() (*Conn, error) {
	runtime := os.Getenv(runtimeEnv)
	sig := os.Getenv(sigEnv)
	if runtime == "" || sig == "" {
		return nil, ErrMissingEnvs
	}

	sock := filepath.Join(runtime, "hypr", sig, ".socket2.sock")
	addr := &net.UnixAddr{
		Name: sock,
		Net:  "unix",
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("connecting to socket: %w", err)
	}

	return &Conn{conn}, nil
}

func (c *Conn) Listen() error {
	sc := bufio.NewScanner(c)
	for sc.Scan() {
		line := sc.Text()
		handleLine(line)
	}

	if err := sc.Err(); err != nil {
		return fmt.Errorf("error scanning: %w", err)
	}

	return nil
}

func handleLine(line string) {
	switch {
	case strings.HasPrefix(line, "openwindow"):
		fmt.Println("ayyyyyy da window opened")
	}
}
