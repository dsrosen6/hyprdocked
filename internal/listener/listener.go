package listener

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dsrosen6/hyprlaptop/internal/hypr"
)

type Listener struct {
	hctlSocketConn *hypr.SocketConn
}

func NewListener(hctlSocketConn *hypr.SocketConn) *Listener {
	return &Listener{
		hctlSocketConn: hctlSocketConn,
	}
}

func ListenForEvents(ctx context.Context, events chan<- Event) error {
	sc, err := hypr.NewSocketConn()
	if err != nil {
		return fmt.Errorf("creating hyprland socket connection: %w", err)
	}

	defer func() {
		if err := sc.Close(); err != nil {
			slog.Error("closing hyprland socket connection", "error", err)
		}
	}()

	l := NewListener(sc)
	return l.listenForEvents(ctx, events)
}

func (l *Listener) listenForEvents(ctx context.Context, events chan<- Event) error {
	errc := make(chan error, 1)
	defer func() {
		if err := l.hctlSocketConn.Close(); err != nil {
			slog.Error("closing hypr socket connection", "error", err)
		}
	}()

	go func() {
		if err := l.ListenHyprctl(ctx, events); err != nil {
			errc <- fmt.Errorf("hyprland listener: %w", err)
		}
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
