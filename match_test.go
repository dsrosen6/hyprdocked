package main

import "testing"

func TestMatchesIdentifiers(t *testing.T) {
	type testVars struct {
		name        string
		hyprMonitor monitor
		cfgMonitor  monitorIdentifiers
		wantMatch   bool
	}

	tests := []testVars{
		{
			name:        "match1",
			hyprMonitor: testMonitorExternal,
			cfgMonitor:  testCfgIdentExternal,
			wantMatch:   true,
		},
		{
			name:        "match2",
			hyprMonitor: testMonitorLaptop,
			cfgMonitor:  testCfgIdentLaptop,
			wantMatch:   true,
		},
		{
			name:        "noMatch1",
			hyprMonitor: testMonitorExternal,
			cfgMonitor:  testCfgIdentLaptop,
			wantMatch:   false,
		},
		{
			name:        "noMatch2",
			hyprMonitor: testMonitorLaptop,
			cfgMonitor:  testCfgIdentExternal,
			wantMatch:   false,
		},
	}

	for _, v := range tests {
		matches := matchesIdentifiers(v.hyprMonitor, v.cfgMonitor)
		if matches != v.wantMatch {
			t.Errorf("%s: want: %v, got: %v", v.name, v.wantMatch, matches)
		}
	}
}
