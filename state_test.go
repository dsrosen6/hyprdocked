package main

import "testing"

func TestStateReady(t *testing.T) {
	type testVars struct {
		name      string
		state     *state
		wantReady bool
	}

	tests := []testVars{
		{
			name: "ready",
			state: &state{
				lidState:   lidStateOpened,
				powerState: powerStateOnBattery,
				monitors:   []monitor{testMonitorExternal},
			},
			wantReady: true,
		},
		// test multiple "not ready" reasons
		{
			name: "notReady1",
			state: &state{
				lidState:   lidStateUnknown,
				powerState: powerStateOnBattery,
				monitors:   []monitor{testMonitorExternal},
			},
			wantReady: false,
		},
		{
			name: "notReady2",
			state: &state{
				lidState:   lidStateOpened,
				powerState: powerStateUnknown,
				monitors:   []monitor{testMonitorExternal},
			},
			wantReady: false,
		},
		{
			name: "notReady3",
			state: &state{
				lidState:   lidStateOpened,
				powerState: powerStateOnAC,
				monitors:   []monitor{},
			},
			wantReady: false,
		},
	}

	for _, v := range tests {
		ready := v.state.ready()
		if ready != v.wantReady {
			t.Errorf("%s: want: %v, got: %v", v.name, v.wantReady, ready)
		}
	}
}
