package listener

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dsrosen6/hyprlaptop/internal/hypr"
	"github.com/godbus/dbus/v5"
)

type Listener struct {
	hctlSocketConn *hypr.SocketConn
	dbusConn       *dbus.Conn
	cfgPath        string
}

func NewListener(hctlSocketConn *hypr.SocketConn, dbusConn *dbus.Conn, cfgPath string) *Listener {
	return &Listener{
		hctlSocketConn: hctlSocketConn,
		dbusConn:       dbusConn,
		cfgPath:        cfgPath,
	}
}

func ListenForEvents(ctx context.Context, cfgPath string, events chan<- Event) error {
	sc, err := hypr.NewSocketConn()
	if err != nil {
		return fmt.Errorf("creating hyprland socket connection: %w", err)
	}

	dc, err := dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("creating dbus connection: %w", err)
	}

	defer func() {
		if err := sc.Close(); err != nil {
			slog.Error("closing hyprland socket connection", "error", err)
		}

		if err := dc.Close(); err != nil {
			slog.Error("closing dbus connection", "error", err)
		}
	}()

	l := NewListener(sc, dc, cfgPath)
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
		slog.Info("listening for hyprland events")
		if err := l.ListenHyprctl(ctx, events); err != nil {
			errc <- fmt.Errorf("hyprland listener: %w", err)
		}
	}()

	go func() {
		slog.Info("listening for config changes")
		if err := l.listenForConfigChanges(ctx, events); err != nil {
			errc <- fmt.Errorf("config listener: %w", err)
		}
	}()

	go func() {
		slog.Info("listening for commands")
		if err := l.commandListener(ctx, events); err != nil {
			errc <- fmt.Errorf("command listener: %w", err)
		}
	}()

	go func() {
		slog.Info("listening for lid events")
		if err := l.listenForLidEvents(ctx, events); err != nil {
			errc <- fmt.Errorf("lid listener: %w", err)
		}
	}()

	go func() {
		slog.Info("listening for power events")
		if err := l.listenForPowerEvents(ctx, events); err != nil {
			errc <- fmt.Errorf("power listener: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errc:
		return err
	}
}
