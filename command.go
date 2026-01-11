package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func sendSuspendCmd() error {
	return sendCmd(string(suspendCmdEvent))
}

func sendWakeCmd() error {
	return sendCmd(string(wakeCmdEvent))
}

func sendCmd(msg string) error {
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

	return nil
}
