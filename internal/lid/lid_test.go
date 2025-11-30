package lid

import "testing"

func TestGetLidState(t *testing.T) {
	state, err := GetState()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("lid state: %s\n", state.string())
}
