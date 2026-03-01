package app

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

func SendPingCmd() error {
	return sendCmd(string(pingCmdEvent), "")
}

func SendIdleCmd(source string) error {
	return sendCmd(string(idleCmdEvent), source)
}

func SendResumeCmd(source string) error {
	return sendCmd(string(resumeCmdEvent), source)
}

func sendCmd(cmd, source string) error {
	msg := cmd
	if source != "" {
		msg = cmd + " " + source
	}

	sock := filepath.Join(os.TempDir(), cmdSockName)
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return fmt.Errorf("command listener not running")
	}

	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error closing socket connection: %v", err)
		}
	}()

	if _, err := conn.Write([]byte(msg)); err != nil {
		return fmt.Errorf("writing message '%s' to socket: %w", msg, err)
	}

	// Half-close the write side so the listener's io.ReadAll returns and it can process the command.
	if err := conn.(*net.UnixConn).CloseWrite(); err != nil {
		return fmt.Errorf("closing write side of socket: %w", err)
	}

	// Block until the listener responds, signalling that processing is complete.
	resp, err := io.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if string(resp) != "OK" {
		return fmt.Errorf("listener returned error: %s", string(resp))
	}

	return nil
}
