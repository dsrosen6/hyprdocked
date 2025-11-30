package main

import "testing"

func TestMonitorToConfigString(t *testing.T) {
	m := Monitor{
		Name:        "eDP-1",
		Width:       1920,
		Height:      1200,
		RefreshRate: 60.001,
		X:           0,
		Y:           0,
		Scale:       1.25,
	}

	t.Log(monitorToConfigString(m))
}
