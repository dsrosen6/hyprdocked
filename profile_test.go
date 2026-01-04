package main

import "testing"

func TestValidateProfile(t *testing.T) {
	type testVars struct {
		name      string
		profile   *profile
		wantValid bool
	}

	cfgMtrs := testMcmDefault
	tests := []testVars{
		{
			name:      "dockedClosedValid",
			profile:   testProfileDockedClosed,
			wantValid: true,
		},
		{
			name:      "dockedClosedInvalid",
			profile:   testProfileDockedClosedInvalid,
			wantValid: false,
		},
		{
			name:      "dockedOpenedValid",
			profile:   testProfileDockedOpened,
			wantValid: true,
		},
		{
			name:      "dockedOpenedInvalid",
			profile:   testProfileDockedOpenedInvalid,
			wantValid: false,
		},
		{
			name:      "laptopOnlyOpenValid",
			profile:   testProfileLaptopOnlyOpen,
			wantValid: true,
		},
		{
			name:      "noMonitorStates",
			profile:   testProfileNoMonitorStates,
			wantValid: false,
		},
	}

	for _, v := range tests {
		v.profile.validate(cfgMtrs)
		if v.profile.valid != v.wantValid {
			t.Errorf("%s: want: %v, got: %v", v.name, v.wantValid, v.profile.valid)
		}
	}
}
