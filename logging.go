package main

import (
	"fmt"
	"log/slog"
)

// monitorLogGroup creates a log group with the provided monitor's info.
func monitorLogGroup(name string, m monitor) slog.Attr {
	ident := slog.Group("identifiers",
		slog.String("name", m.Name),
		slog.String("description", m.Description),
	)

	settings := slog.Group("settings",
		slog.Int64("width", m.Width),
		slog.Int64("height", m.Height),
		slog.Float64("refresh_rate", m.RefreshRate),
		slog.String("position", fmt.Sprintf("%dx%d", m.X, m.Y)),
		slog.Float64("scale", m.Scale),
	)

	return slog.Group(name, ident, settings)
}
