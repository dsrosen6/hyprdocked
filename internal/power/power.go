package power

import (
	"context"
	"fmt"
	"slices"

	"github.com/godbus/dbus/v5"
)

const (
	upowerDest     = "org.freedesktop.UPower"
	upowerPath     = "/org/freedesktop/UPower"
	upowerMatchIfc = "org.freedesktop.DBus.Properties"
	upowerMatchMbr = "PropertiesChanged"
	upowerMethod   = "org.freedesktop.DBus.Properties.Get"
	onBatProperty  = "OnBattery"
)

type (
	Handler struct {
		conn    *dbus.Conn
		Events  chan struct{}
		signals chan *dbus.Signal
	}

	State string
)

const (
	StateUnknown   State = "unknown"
	StateOnBattery State = "battery"
	StateOnAC      State = "ac"
)

func NewHandler(conn *dbus.Conn) *Handler {
	return &Handler{
		conn:    conn,
		Events:  make(chan struct{}, 10),
		signals: make(chan *dbus.Signal, 10),
	}
}

func (p *Handler) ListenForChanges(ctx context.Context) error {
	defer close(p.Events)
	defer p.conn.RemoveSignal(p.signals)
	if err := p.startDbus(ctx); err != nil {
		return err
	}

	for {
		select {
		case sig, ok := <-p.signals:
			if !ok {
				return fmt.Errorf("signals channel closed")
			}

			if !p.shouldHandleSignal(sig) {
				continue
			}

			select {
			case p.Events <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Handler) startDbus(ctx context.Context) error {
	if err := p.conn.AddMatchSignalContext(
		ctx, dbus.WithMatchInterface(upowerMatchIfc), dbus.WithMatchMember(upowerMatchMbr),
		dbus.WithMatchObjectPath(dbus.ObjectPath(upowerPath)),
	); err != nil {
		return fmt.Errorf("failed to add dbus match rule: %w", err)
	}

	p.conn.Signal(p.signals)
	return nil
}

func (p *Handler) GetCurrentState(ctx context.Context) (State, error) {
	obj := p.conn.Object(upowerDest, upowerPath)
	var result dbus.Variant
	if err := obj.CallWithContext(ctx, upowerMethod, 0, upowerDest, onBatProperty).Store(&result); err != nil {
		return StateUnknown, err
	}

	if onBat, ok := result.Value().(bool); ok {
		if onBat {
			return StateOnBattery, nil
		}
		return StateOnAC, nil
	}

	return StateUnknown, fmt.Errorf("unexpected type for OnBattery")
}

func (p *Handler) shouldHandleSignal(sig *dbus.Signal) bool {
	if sig.Name != "org.freedesktop.DBus.Properties.PropertiesChanged" {
		return false
	}

	if len(sig.Body) < 2 {
		return false
	}

	if changedProps, ok := sig.Body[1].(map[string]dbus.Variant); ok {
		if _, exists := changedProps[onBatProperty]; exists {
			return true
		}
	}

	if len(sig.Body) >= 3 {
		if invalidated, ok := sig.Body[2].([]string); ok {
			if slices.Contains(invalidated, onBatProperty) {
				return true
			}
		}
	}

	return false
}
