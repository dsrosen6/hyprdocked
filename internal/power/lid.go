package power

import (
	"context"
	"fmt"
	"slices"

	"github.com/godbus/dbus/v5"
)

type (
	LidListener struct {
		conn    *dbus.Conn
		events  chan LidEvent
		signals chan *dbus.Signal
	}

	LidEvent struct {
		State LidState
	}

	LidState int
)

const (
	uPowerDest     = "org.freedesktop.UPower"
	uPowerPath     = "/org/freedesktop/UPower"
	uPowerMatchIfc = "org.freedesktop.DBus.Properties"
	uPowerMatchMbr = "PropertiesChanged"
	uPowerMethod   = "org.freedesktop.DBus.Properties.Get"
	uPowerProperty = "LidIsClosed"

	LidStateUnknown LidState = iota
	LidStateOpened
	LidStateClosed

	LidStateUnknownStr = "unknown"
	LidStateOpenedStr  = "opened"
	LidStateClosedStr  = "closed"
)

func NewLidListener(conn *dbus.Conn) *LidListener {
	return &LidListener{
		conn:    conn,
		events:  make(chan LidEvent, 10),
		signals: make(chan *dbus.Signal, 10),
	}
}

func (l *LidListener) Events() <-chan LidEvent {
	return l.events
}

func (l *LidListener) Run(ctx context.Context) error {
	defer close(l.events)
	defer l.conn.RemoveSignal(l.signals)
	if err := l.Start(ctx); err != nil {
		return err
	}

	lastState := LidStateUnknown

	for {
		select {
		case sig, ok := <-l.signals:
			if !ok {
				return fmt.Errorf("signals channel closed")
			}

			if !l.shouldHandleSignal(sig) {
				continue
			}

			currentState, err := l.GetCurrentState(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current lid state: %w", err)
			}

			if currentState != lastState {
				select {
				case l.events <- LidEvent{State: currentState}:
					lastState = currentState
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		// TODO: handle
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (l *LidListener) Start(ctx context.Context) error {
	if err := l.conn.AddMatchSignalContext(
		ctx, dbus.WithMatchInterface(uPowerMatchIfc), dbus.WithMatchMember(uPowerMatchMbr),
		dbus.WithMatchObjectPath(dbus.ObjectPath(uPowerPath)),
	); err != nil {
		return fmt.Errorf("failed to add dbus match rule: %w", err)
	}

	l.conn.Signal(l.signals)
	return nil
}

func (l *LidListener) GetCurrentState(ctx context.Context) (LidState, error) {
	obj := l.conn.Object(uPowerDest, uPowerPath)
	var result dbus.Variant
	if err := obj.CallWithContext(ctx, uPowerMethod, 0, uPowerDest, uPowerProperty).Store(&result); err != nil {
		return LidStateUnknown, err
	}

	if closed, ok := result.Value().(bool); ok {
		if closed {
			return LidStateClosed, nil
		}
		return LidStateOpened, nil
	}

	return LidStateUnknown, fmt.Errorf("unexpected type for LidIsClosed")
}

func (l *LidListener) shouldHandleSignal(sig *dbus.Signal) bool {
	if sig.Name != "org.freedesktop.DBus.Properties.PropertiesChanged" {
		return false
	}

	if len(sig.Body) < 2 {
		return false
	}

	if changedProps, ok := sig.Body[1].(map[string]dbus.Variant); ok {
		if _, exists := changedProps[uPowerProperty]; exists {
			return true
		}
	}

	if len(sig.Body) >= 3 {
		if invalidated, ok := sig.Body[2].([]string); ok {
			if slices.Contains(invalidated, uPowerProperty) {
				return true
			}
		}
	}

	return false
}

func (ls LidState) String() string {
	switch ls {
	case LidStateOpened:
		return LidStateOpenedStr
	case LidStateClosed:
		return LidStateClosedStr
	default:
		return LidStateUnknownStr
	}
}

func ParseLidState(s string) LidState {
	switch s {
	case LidStateOpenedStr:
		return LidStateOpened
	case LidStateClosedStr:
		return LidStateClosed
	default:
		return LidStateUnknown
	}
}
