package listener

import (
	"context"
	"log/slog"

	"github.com/dsrosen6/hyprlaptop/internal/power"
)

func (l *Listener) listenForLidEvents(ctx context.Context, events chan<- Event) error {
	lidListener := power.NewLidListener(l.dbusConn)

	go func() {
		if err := lidListener.Run(ctx); err != nil && err != context.Canceled {
			slog.Error("lid listener stopped", "error", err)
		}
	}()

	for lidEvent := range lidListener.Events() {
		select {
		case events <- Event{Type: LidSwitchEvent, Details: lidEvent.State.String()}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (l *Listener) listenForPowerEvents(ctx context.Context, events chan<- Event) error {
	powerListener := power.NewPowerListener(l.dbusConn)

	go func() {
		if err := powerListener.Run(ctx); err != nil && err != context.Canceled {
			slog.Error("power listener stopped", "error", err)
		}
	}()

	for powerEvent := range powerListener.Events() {
		select {
		case events <- Event{Type: PowerChangedEvent, Details: powerEvent.State.String()}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
