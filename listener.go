package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/godbus/dbus/v5"
)

type (
	listener struct {
		hctlSocketConn *hyprSocketConn
		lidHandler     *lidHandler
		powerHandler   *powerHandler
		cfgPath        string
	}

	listenerEvent struct {
		Type    eventType
		Details string
	}

	listenerParams struct {
		hyprSockConn *hyprSocketConn
		lidHandler   *lidHandler
		powerHandler *powerHandler
		dbusConn     *dbus.Conn
		cfgPath      string
	}

	eventType string
)

// We are only actively filtering for the v2 monitor events as to not double up (since hyprland
// sends both a "v1" (monitoradded or monitorremoved) but it's expected that v2 is deprecated and just
// replaces the original, so this will probably change.
var monitorEvents = map[string]eventType{
	"monitoraddedv2":   displayAddEvent,
	"monitorremovedv2": displayRemoveEvent,
}

const (
	configUpdatedEvent  eventType = "CONFIG_UPDATED"
	displayInitialEvent eventType = "DISPLAY_INITIAL"
	displayAddEvent     eventType = "DISPLAY_ADDED"
	displayRemoveEvent  eventType = "DISPLAY_REMOVED"
	displayUnknownEvent eventType = "DISLAY_UNKNOWN_EVENT"
	lidSwitchEvent      eventType = "LID_SWITCH"
	powerChangedEvent   eventType = "POWER_CHANGED"
	suspendCmdEvent     eventType = "SUSPEND_CMD"
	wakeCmdEvent        eventType = "WAKE_CMD"
	cmdSockName                   = "hyprlaptop.sock"
)

func newListener(p listenerParams) (*listener, error) {
	return &listener{
		hctlSocketConn: p.hyprSockConn,
		lidHandler:     p.lidHandler,
		powerHandler:   p.powerHandler,
		cfgPath:        p.cfgPath,
	}, nil
}

// listenAndHandle starts hyprlaptop's listener, which handles hyprctl display add/remove events
// and events from the hyprlaptop CLI.
func (a *app) listenAndHandle(ctx context.Context) error {
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
				m, err := a.hctl.listMonitors()
				if err != nil {
					slog.Error("listing current monitors", "error", err)
					continue
				}
				if !reflect.DeepEqual(a.currentState.monitors, m) {
					a.currentState.monitors = m
					slog.Debug("monitors state updated", "state", a.currentState.monitors)
				}

			case lidSwitchEvent:
				a.currentState.lidState = parseLidState(ev.Details)
				slog.Debug("lid state updated", "state", a.currentState.lidState)

			case powerChangedEvent:
				a.currentState.powerState = parsePowerState(ev.Details)
				slog.Debug("power state updated", "state", a.currentState.powerState)

			case configUpdatedEvent:
				// Update config values
				err := a.cfg.reload(5)
				if err != nil {
					slog.Error("reloading config", "error", err)
					continue
				}
			case suspendCmdEvent:
				slog.Info("suspended command received")
				a.currentState.mode = modeSuspending

			case wakeCmdEvent:
				slog.Info("wake command received")
				a.currentState.mode = modeWaking
			}

			if !a.currentState.ready() {
				slog.Debug("not ready; awaiting initial values")
				continue
			}

			if a.updating || time.Since(a.lastUpdateEnd) < 500*time.Millisecond {
				slog.Debug("skipping: in cooldown")
				continue
			}

			if err := a.runUpdater(); err != nil {
				slog.Error("running updater", "error", err)
			}

			if a.currentState.mode == modeWaking {
				a.currentState.mode = modeNormal
				slog.Debug("wake done; resetting suspended monitors", "total_to_remove", len(a.currentState.suspendedMonitors))
				a.currentState.suspendedMonitors = []monitor{}
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
		slog.Debug("listening for config changes")
		if err := l.listenConfigChanges(ctx, events); err != nil {
			errc <- fmt.Errorf("config listener: %w", err)
		}
	}()

	go func() {
		slog.Debug("listening for lid events")
		if err := l.listenLidEvents(ctx, events); err != nil {
			errc <- fmt.Errorf("lid listener: %w", err)
		}
	}()

	go func() {
		slog.Debug("listening for power events")
		if err := l.listenPowerEvents(ctx, events); err != nil {
			errc <- fmt.Errorf("power listener: %w", err)
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

// listenHyprctl listens for hyprctl events and sends an event if it is a monitor add or removal.
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

// listenConfigChanges changes listens for changes in the config file; if a change is detected,
// hyprlaptop performs a live reload.
func (l *listener) listenConfigChanges(ctx context.Context, events chan<- listenerEvent) error {
	var lastHash [32]byte

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating config file watcher: %w", err)
	}
	slog.Debug("config watcher: fsnotify watcher created")

	defer func() {
		if err := w.Close(); err != nil {
			slog.Error("closing config file watcher", "error", err)
		}
	}()

	dir := filepath.Dir(l.cfgPath)
	err = w.Add(dir)
	if err != nil {
		return fmt.Errorf("adding config directory to watcher: %w", err)
	}
	slog.Debug("config watcher: fsnotify watch list", "list", w.WatchList())

	for {
		select {
		case <-ctx.Done():
			return nil

		case event, ok := <-w.Events:
			if !ok {
				return nil
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Rename) {
				h, err := fileHash(l.cfgPath)
				if err != nil {
					continue
				}

				if h == lastHash {
					slog.Debug("config watcher: received identical hash for file update, no changes needed")
					continue
				}

				lastHash = h

				slog.Debug("fsnotify: file modified", "file", event.Name)
				events <- listenerEvent{
					Type:    configUpdatedEvent,
					Details: l.cfgPath,
				}
			}

		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			return fmt.Errorf("config watcher fsnotify error: %w", err)
		}
	}
}

func (l *listener) listenLidEvents(ctx context.Context, events chan<- listenerEvent) error {
	go func() {
		if err := l.lidHandler.listenForChanges(ctx); err != nil && err != context.Canceled {
			slog.Error("lid listener stopped", "error", err)
		}
	}()

	for lidEvent := range l.lidHandler.events {
		select {
		case events <- listenerEvent{Type: lidSwitchEvent, Details: string(lidEvent.State)}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (l *listener) listenPowerEvents(ctx context.Context, events chan<- listenerEvent) error {
	go func() {
		if err := l.powerHandler.listenForChanges(ctx); err != nil && err != context.Canceled {
			slog.Error("power listener stopped", "error", err)
		}
	}()

	for powerEvent := range l.powerHandler.events {
		select {
		case events <- listenerEvent{Type: powerChangedEvent, Details: string(powerEvent.State)}:
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
			slog.Error("command listener: closing hyprlaptop socket", "error", err)
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
				case string(wakeCmdEvent):
					events <- listenerEvent{Type: wakeCmdEvent}
				case string(suspendCmdEvent):
					events <- listenerEvent{Type: suspendCmdEvent}
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

	if et, ok := monitorEvents[parts[0]]; ok {
		ev.Type = et
		ev.Details = parts[1]
	}

	return *ev, nil
}

func fileHash(path string) ([32]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(data), nil
}
