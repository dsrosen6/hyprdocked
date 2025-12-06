package lid

import (
	"fmt"
	"os"
	"strings"
)

const (
	lidStateFile = "/proc/acpi/button/lid/LID/state"
)

type State string

const (
	StateUnknown State = "Unknown"
	StateOpen    State = "Open"
	StateClosed  State = "Closed"
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
