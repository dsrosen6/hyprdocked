package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/dsrosen6/hyprdocked/internal/power"
	"github.com/godbus/dbus/v5"
)

type (
	listener struct {
		hctlSocketConn *hyprSocketConn
		lidHandler     *power.LidHandler
	}

	listenerEvent struct {
		Type    eventType
		Details string
	}

	listenerParams struct {
		hyprSockConn *hyprSocketConn
		lidHandler   *power.LidHandler
		dbusConn     *dbus.Conn
	}

	eventType string
)

// We are only actively filtering for the v2 monitor events as to not double up (since hyprland
// sends both a "v1" (monitoradded or monitorremoved) but it's expected that v2 is deprecated and just
// replaces the original, so this will probably change.
var displayEvents = map[string]eventType{
	"monitoraddedv2":   displayAddEvent,
	"monitorremovedv2": displayRemoveEvent,
}

const (
	displayInitialEvent eventType = "DISPLAY_INITIAL"
	displayAddEvent     eventType = "DISPLAY_ADDED"
	displayRemoveEvent  eventType = "DISPLAY_REMOVED"
	displayUnknownEvent eventType = "DISLAY_UNKNOWN_EVENT"
	lidSwitchEvent      eventType = "LID_SWITCH"
	// powerChangedEvent   eventType = "POWER_CHANGED"
	idleCmdEvent   eventType = "IDLE_CMD"
	resumeCmdEvent eventType = "RESUME_CMD"
	cmdSockName              = "hyprdocked.sock"
)

func newListener(p listenerParams) (*listener, error) {
	return &listener{
		hctlSocketConn: p.hyprSockConn,
		lidHandler:     p.lidHandler,
	}, nil
}

// listenAndHandle starts the hyprdocked listener, which handles hyprctl display add/remove events
// and events from the hyprdocked CLI.
func (a *App) listenAndHandle(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	events := make(chan listenerEvent, 16)
	errc := make(chan error, 1)

	go func() {
		slog.Info("listening for updates")
		if err := a.listener.listen(ctx, events); err != nil {
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

			slog.Debug("received event from listener", "type", ev.Type, "details", ev.Details)
			switch ev.Type {
			case displayInitialEvent, displayAddEvent,
				displayRemoveEvent, displayUnknownEvent:
				m, err := a.hctl.listDisplays()
				if err != nil {
					slog.Error("listing current displays", "error", err)
					continue
				}
				if !reflect.DeepEqual(a.allDisplays, m) {
					a.allDisplays = m
					slog.Debug("displays state updated", "state", a.allDisplays)
				}

			case lidSwitchEvent:
				ls, err := a.listener.lidHandler.GetCurrentState(ctx)
				if err != nil {
					slog.Error("fetching lid state", "error", err)
					continue
				}

				if a.lidState != ls {
					a.lidState = ls
					slog.Debug("lid state updated", "state", a.lidState)
				}

			case idleCmdEvent:
				slog.Info("idle command received")
				a.mode = modeSuspending

			case resumeCmdEvent:
				slog.Info("resume command received")
				a.mode = modeNormal
				continue
			}

			if !a.ready() {
				slog.Debug("not ready; awaiting initial values")
				continue
			}

			if a.updating {
				slog.Debug("skipping: mid update")
				continue
			}

			if err := a.runUpdater(); err != nil {
				slog.Error("running updater", "error", err)
			}

		case err := <-errc:
			return fmt.Errorf("listener failed: %w", err)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (l *listener) listen(ctx context.Context, events chan<- listenerEvent) error {
	errc := make(chan error, 1)
	go func() {
		slog.Debug("listening for hyprland events")
		if err := l.listenHyprctl(ctx, events); err != nil {
			errc <- fmt.Errorf("hyprland listener: %w", err)
		}
	}()

	go func() {
		slog.Debug("listening for lid events")
		if err := l.listenLidEvents(ctx, events); err != nil {
			errc <- fmt.Errorf("lid listener: %w", err)
		}
	}()

	go func() {
		slog.Debug("listening for command events")
		if err := l.listenCommandEvents(ctx, events); err != nil {
			errc <- fmt.Errorf("command listener: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errc:
		return err
	}
}

// listenHyprctl listens for hyprctl events and sends an event if it is a display add or removal.
func (l *listener) listenHyprctl(ctx context.Context, events chan<- listenerEvent) error {
	var lastEvent listenerEvent
	scn := bufio.NewScanner(l.hctlSocketConn)
	for scn.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line := scn.Text()

			ev, err := parseDisplayEvent(line)
			if err != nil {
				slog.Error("parse error", "err", err)
				continue
			}

			if ev.Type == displayUnknownEvent {
				continue
			}

			// store and check for last event so it doesn't attempt to send an unnecessary event if received
			if reflect.DeepEqual(lastEvent, ev) {
				slog.Debug("hyprctl listener: new event matches last event, no action needed")
				continue
			}

			lastEvent = ev
			events <- ev
		}
	}

	if err := scn.Err(); err != nil {
		return fmt.Errorf("error scanning: %w", err)
	}

	return nil
}

func (l *listener) listenLidEvents(ctx context.Context, events chan<- listenerEvent) error {
	go func() {
		if err := l.lidHandler.ListenForChanges(ctx); err != nil && err != context.Canceled {
			slog.Error("lid listener stopped", "error", err)
		}
	}()

	for range l.lidHandler.Events {
		select {
		case events <- listenerEvent{Type: lidSwitchEvent}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (l *listener) listenCommandEvents(ctx context.Context, events chan<- listenerEvent) error {
	sock := filepath.Join(os.TempDir(), cmdSockName)

	// remove existing file if it already exists
	_ = os.Remove(sock)

	ln, err := net.Listen("unix", sock)
	if err != nil {
		return fmt.Errorf("command listener: listening to unix socket: %w", err)
	}

	defer func() {
		if err := ln.Close(); err != nil {
			slog.Error("command listener: closing hyprdocked socket", "error", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := ln.Accept()
			if err != nil {
				continue
			}

			go func() {
				defer func() {
					if err := conn.Close(); err != nil {
						slog.Error("command listener: closing socket conn", "error", err)
					} else {
						slog.Debug("command listener: socket conn closed")
					}
				}()

				buf, _ := io.ReadAll(conn)
				msg := strings.TrimSpace(string(buf))

				switch msg {
				case string(resumeCmdEvent):
					events <- listenerEvent{Type: resumeCmdEvent}
				case string(idleCmdEvent):
					events <- listenerEvent{Type: idleCmdEvent}
				default:
					slog.Warn("command listener: got unknown command", "command", msg)
				}
			}()
		}
	}
}

// parseDisplayEvent splits the event string and returns what type of event it is.
func parseDisplayEvent(line string) (listenerEvent, error) {
	parts := strings.SplitN(line, ">>", 2)
	if len(parts) != 2 {
		return listenerEvent{}, fmt.Errorf("invalid event: %q", line)
	}

	ev := &listenerEvent{
		Type: displayUnknownEvent,
	}

	if et, ok := displayEvents[parts[0]]; ok {
		ev.Type = et
		ev.Details = parts[1]
	}

	return *ev, nil
}
