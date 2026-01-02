package power

import (
	"context"
	"fmt"
	"slices"

	"github.com/godbus/dbus/v5"
)

type (
	PowerListener struct {
		conn    *dbus.Conn
		events  chan PowerEvent
		signals chan *dbus.Signal
	}

	PowerEvent struct {
		State PowerState
	}

	PowerState int
)

const (
	uPowerOnBatProperty = "OnBattery"

	PowerStateUnknownStr   = "unknown"
	PowerStateOnBatteryStr = "battery"
	PowerStateOnACStr      = "ac"
)

const (
	PowerStateUnknown PowerState = iota
	PowerStateOnBattery
	PowerStateOnAC
)

func NewPowerListener(conn *dbus.Conn) *PowerListener {
	return &PowerListener{
		conn:    conn,
		events:  make(chan PowerEvent, 10),
		signals: make(chan *dbus.Signal, 10),
	}
}

func (p *PowerListener) Events() <-chan PowerEvent {
	return p.events
}

func (p *PowerListener) Run(ctx context.Context) error {
	defer close(p.events)
	defer p.conn.RemoveSignal(p.signals)
	if err := p.Start(ctx); err != nil {
		return err
	}

	initialState, err := p.GetCurrentState(ctx)
	if err != nil {
		return fmt.Errorf("getting initial lid state: %w", err)
	}

	select {
	case p.events <- PowerEvent{State: initialState}:
	case <-ctx.Done():
		return ctx.Err()
	}

	lastState := initialState

	for {
		select {
		case sig, ok := <-p.signals:
			if !ok {
				return fmt.Errorf("signals channel closed")
			}

			if !p.shouldHandleSignal(sig) {
				continue
			}

			currentState, err := p.GetCurrentState(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current lid state: %w", err)
			}

			if currentState != lastState {
				select {
				case p.events <- PowerEvent{State: currentState}:
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

func (p *PowerListener) Start(ctx context.Context) error {
	if err := p.conn.AddMatchSignalContext(
		ctx, dbus.WithMatchInterface(uPowerMatchIfc), dbus.WithMatchMember(uPowerMatchMbr),
		dbus.WithMatchObjectPath(dbus.ObjectPath(uPowerPath)),
	); err != nil {
		return fmt.Errorf("failed to add dbus match rule: %w", err)
	}

	p.conn.Signal(p.signals)
	return nil
}

func (p *PowerListener) GetCurrentState(ctx context.Context) (PowerState, error) {
	obj := p.conn.Object(uPowerDest, uPowerPath)
	var result dbus.Variant
	if err := obj.CallWithContext(ctx, uPowerMethod, 0, uPowerDest, uPowerOnBatProperty).Store(&result); err != nil {
		return PowerStateUnknown, err
	}

	if onBat, ok := result.Value().(bool); ok {
		if onBat {
			return PowerStateOnBattery, nil
		}
		return PowerStateOnAC, nil
	}

	return PowerStateUnknown, fmt.Errorf("unexpected type for OnBattery")
}

func (p *PowerListener) shouldHandleSignal(sig *dbus.Signal) bool {
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

func (ps PowerState) String() string {
	switch ps {
	case PowerStateOnAC:
		return PowerStateOnACStr
	case PowerStateOnBattery:
		return PowerStateOnBatteryStr
	default:
		return PowerStateUnknownStr
	}
}

func ParsePowerState(s string) PowerState {
	switch s {
	case PowerStateOnACStr:
		return PowerStateOnAC
	case PowerStateOnBatteryStr:
		return PowerStateOnBattery
	default:
		return PowerStateUnknown
	}
}
