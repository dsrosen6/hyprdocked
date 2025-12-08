package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dsrosen6/hyprlaptop/internal/listener"
)

func (a *App) Listen(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	events := make(chan listener.Event, 16)
	errc := make(chan error, 1)

	go func() {
		if err := listener.ListenForEvents(ctx, events); err != nil {
			errc <- err
			cancel()
		}
	}()

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return nil // normal shutdown
			}
			slog.Info("received event from listener", "type", ev.Type, "details", ev.Details)
			if err := a.Run(); err != nil {
				slog.Error("error running display updater", "error", err)
			}

		case err := <-errc:
			return fmt.Errorf("listener failed: %w", err)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
