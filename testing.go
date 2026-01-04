package main

type (
	testMtrConfigMap struct {
		name   string
		cfgMap monitorConfigMap
	}
)

var (
	// test state for detecting that laptop should become enabled
	testStateExternalLidOpen = &state{
		lidState:   lidStateOpened,
		powerState: powerStateOnAC,
		monitors:   []monitor{testMonitorExternal},
	}

	// test state for detecting that laptop should become disabled
	testStateExternalLidClosed = &state{
		lidState:   lidStateClosed,
		powerState: powerStateOnAC,
		monitors:   []monitor{testMonitorExternal, testMonitorLaptop},
	}

	testHyprMonitors = []monitor{testMonitorExternal, testMonitorLaptop}

	testMonitorExternal = monitor{
		monitorIdentifiers: monitorIdentifiers{
			Name:        "DP-1",
			Description: "Samsung Electric Company Odyssey G85SD H1AK500000",
		},
		monitorSettings: monitorSettings{
			Width:       3440,
			Height:      1440,
			RefreshRate: 174.96201,
			X:           0,
			Y:           0,
			Scale:       1,
		},
	}

	testMonitorLaptop = monitor{
		monitorIdentifiers: monitorIdentifiers{
			Name:        "eDP-1",
			Description: "China Star Optoelectronics Technology Co. Ltd MNE007JA1-3",
		},
		monitorSettings: monitorSettings{
			Width:       1920,
			Height:      1200,
			RefreshRate: 60.001,
			X:           3440,
			Y:           0,
			Scale:       1.25,
		},
	}

	testCfgIdentLaptop = monitorIdentifiers{
		Name: "eDP-1",
	}

	testCfgIdentExternal = monitorIdentifiers{
		Name:        "DP-1",
		Description: "Samsung Electric Company Odyssey G85SD H1AK500000",
	}

	testMcmDefault = monitorConfigMap{
		"laptop": monitorConfig{
			Identifiers: testCfgIdentLaptop,
			Presets: monitorPresetMap{
				"default": monitorSettings{
					Width:       1920,
					Height:      1200,
					RefreshRate: 60.001,
					X:           3440,
					Y:           0,
					Scale:       1.25,
				},
			},
		},
		"external": monitorConfig{
			Identifiers: testCfgIdentExternal,
			Presets: monitorPresetMap{
				"default": monitorSettings{
					Width:       3440,
					Height:      1440,
					RefreshRate: 174.96201,
					X:           0,
					Y:           0,
					Scale:       1,
				},
			},
		},
	}

	testProfileDockedClosed = &profile{
		Name: "docked-laptop-closed",
		Conditions: conditions{
			LidState:        lidStateToPtr(lidStateClosed),
			EnabledMonitors: []string{"external"},
		},
		MonitorStates: []monitorState{
			{
				Label:  "external",
				Preset: strToPtr("default"),
			},
			{
				Label:   "laptop",
				Disable: true,
			},
		},
	}

	testProfileDockedClosedInvalid = &profile{
		Name: "docked-laptop-closed",
		Conditions: conditions{
			LidState:        lidStateToPtr(lidStateUnknown),
			EnabledMonitors: []string{"externall"},
		},
		MonitorStates: []monitorState{
			{
				Label:  "external",
				Preset: strToPtr("default"),
			},
			{
				Label:   "laptop",
				Disable: true,
			},
		},
	}

	testProfileDockedOpened = &profile{
		Name: "docked-laptop-open",
		Conditions: conditions{
			LidState:        lidStateToPtr(lidStateOpened),
			EnabledMonitors: []string{"external"},
		},
		MonitorStates: []monitorState{
			{
				Label:  "external",
				Preset: strToPtr("default"),
			},
			{
				Label:  "laptop",
				Preset: strToPtr("default"),
			},
		},
	}

	testProfileDockedOpenedInvalid = &profile{
		Name: "docked-laptop-open",
		Conditions: conditions{
			LidState:        lidStateToPtr(lidStateOpened),
			EnabledMonitors: []string{"something"},
		},
		MonitorStates: []monitorState{
			{
				Label:  "external",
				Preset: strToPtr("default"),
			},
			{
				Label:  "idk",
				Preset: strToPtr("default"),
			},
		},
	}

	testProfileLaptopOnlyOpen = &profile{
		Name:       "laptop-only-open",
		ExactMatch: true,
		Conditions: conditions{
			LidState:        lidStateToPtr(lidStateOpened),
			EnabledMonitors: []string{"laptop"},
		},
		MonitorStates: []monitorState{
			{
				Label:  "laptop",
				Preset: strToPtr("default"),
			},
		},
	}

	testProfileNoMonitorStates = &profile{
		Name: "no-monitor-states",
		Conditions: conditions{
			LidState:        lidStateToPtr(lidStateOpened),
			EnabledMonitors: []string{"laptop"},
		},
	}

	testProfileSetDefault = []*profile{
		testProfileDockedClosed,
		testProfileDockedOpened,
		testProfileLaptopOnlyOpen,
	}
)
