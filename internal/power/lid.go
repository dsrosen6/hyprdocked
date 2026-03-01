package power

import (
	"context"
	"fmt"
	"slices"

	"github.com/godbus/dbus/v5"
)

const (
	lidProperty = "LidIsClosed"
)

type (
	LidHandler struct {
		conn    *dbus.Conn
		Events  chan struct{}
		signals chan *dbus.Signal
	}

	LidState string
)

const (
	LidStateUnknown LidState = "unknown"
	LidStateOpened  LidState = "opened"
	LidStateClosed  LidState = "closed"
)

func NewLidHandler(conn *dbus.Conn) *LidHandler {
	return &LidHandler{
		conn:    conn,
		Events:  make(chan struct{}, 10),
		signals: make(chan *dbus.Signal, 10),
	}
}

func (l *LidHandler) ListenForChanges(ctx context.Context) error {
	defer close(l.Events)
	defer l.conn.RemoveSignal(l.signals)
	if err := l.startDbus(ctx); err != nil {
		return err
	}

	for {
		select {
		case sig, ok := <-l.signals:
			if !ok {
				return fmt.Errorf("signals channel closed")
			}

			if !l.shouldHandleSignal(sig) {
				continue
			}

			select {
			case l.Events <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (l *LidHandler) startDbus(ctx context.Context) error {
	if err := l.conn.AddMatchSignalContext(
		ctx, dbus.WithMatchInterface(upowerMatchIfc), dbus.WithMatchMember(upowerMatchMbr),
		dbus.WithMatchObjectPath(dbus.ObjectPath(upowerPath)),
	); err != nil {
		return fmt.Errorf("failed to add dbus match rule: %w", err)
	}

	l.conn.Signal(l.signals)
	return nil
}

func (l *LidHandler) GetCurrentState(ctx context.Context) (LidState, error) {
	obj := l.conn.Object(upowerDest, upowerPath)
	var result dbus.Variant
	if err := obj.CallWithContext(ctx, upowerMethod, 0, upowerDest, lidProperty).Store(&result); err != nil {
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

func (l *LidHandler) shouldHandleSignal(sig *dbus.Signal) bool {
	if sig.Name != "org.freedesktop.DBus.Properties.PropertiesChanged" {
		return false
	}

	if len(sig.Body) < 2 {
		return false
	}

	if changedProps, ok := sig.Body[1].(map[string]dbus.Variant); ok {
		if _, exists := changedProps[lidProperty]; exists {
			return true
		}
	}

	if len(sig.Body) >= 3 {
		if invalidated, ok := sig.Body[2].([]string); ok {
			if slices.Contains(invalidated, lidProperty) {
				return true
			}
		}
	}

	return false
}
