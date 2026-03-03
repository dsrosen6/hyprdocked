package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	serviceName = "hyprdocked.service"

	serviceContent = `[Unit]
Description=Hyprdocked Listener
After=wayland-session@Hyprland.target

[Service]
ExecStart=%s listen
Restart=on-failure
RestartSec=2

[Install]
WantedBy=wayland-session@Hyprland.target
`
)

func Install(customBinary string) error {
	var execPath string
	if customBinary != "" {
		execPath = customBinary
		if _, err := os.Stat(execPath); os.IsNotExist(err) {
			return fmt.Errorf("custom binary path %s does not exist", execPath)
		}
	} else {
		var err error
		execPath, err = os.Executable()
		if err != nil {
			return fmt.Errorf("getting executable path: %w", err)
		}
	}

	serviceDir, err := userServiceDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		return fmt.Errorf("creating service directory: %w", err)
	}

	servicePath := filepath.Join(serviceDir, serviceName)
	content := fmt.Sprintf(serviceContent, execPath)
	if err := os.WriteFile(servicePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing service file: %w", err)
	}

	fmt.Println("service file written to", servicePath)

	if err := systemctlUser("daemon-reload"); err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}

	if err := systemctlUser("enable", serviceName); err != nil {
		return fmt.Errorf("enabling service: %w", err)
	}

	if err := systemctlUser("start", serviceName); err != nil {
		return fmt.Errorf("starting service: %w", err)
	}

	return nil
}

func Restart() error {
	return systemctlUser("restart", serviceName)
}

func Uninstall() error {
	_ = systemctlUser("stop", serviceName)
	_ = systemctlUser("disable", serviceName)

	serviceDir, err := userServiceDir()
	if err != nil {
		return err
	}

	servicePath := filepath.Join(serviceDir, serviceName)
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing service file: %w", err)
	}

	fmt.Println("service file removed from", servicePath)

	if err := systemctlUser("daemon-reload"); err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}

	return nil
}

func ShowLogs(stream bool) error {
	args := []string{"-u", serviceName}
	if stream {
		// tail logs
		args = append(args, "-f")
	} else {
		// show in pager starting at end
		args = append(args, "-e")
	}

	return journalctlUser(args...)
}

func userServiceDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	return filepath.Join(home, ".config", "systemd", "user"), nil
}

func systemctlUser(args ...string) error {
	cmd := exec.Command("systemctl", append([]string{"--user"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func journalctlUser(args ...string) error {
	cmd := exec.Command("journalctl", append([]string{"--user"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
