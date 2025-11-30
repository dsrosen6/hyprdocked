// Package lid handles laptop lid detection.
package lid

import (
	"fmt"
	"os"
	"strings"
)

const (
	lidStateFile = "/proc/acpi/button/lid/LID/state"
)

type State int

const (
	StateUnknown State = iota
	StateOpen
	StateClosed
)

func GetState() (State, error) {
	b, err := os.ReadFile(lidStateFile)
	if err != nil {
		return StateUnknown, fmt.Errorf("reading lid state file: %w", err)
	}
	s := strings.ToLower(string(b))

	switch {
	case strings.Contains(s, "open"):
		return StateOpen, nil
	case strings.Contains(s, "closed"):
		return StateClosed, nil
	default:
		return StateUnknown, nil
	}
}

func (s State) string() string {
	switch s {
	case StateOpen:
		return "open"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}
