package setting

import "testing"

func TestMeteoritePriceFactor(t *testing.T) {
	if err := UpdateMeteoritePricePercentByJSONString(`{"claude-x": 60, "bad": 0}`); err != nil {
		t.Fatal(err)
	}
	if got := MeteoritePriceFactor("claude-x"); got != 0.6 {
		t.Errorf("configured model: %v", got)
	}
	if got := MeteoritePriceFactor("other"); got != 0.5 {
		t.Errorf("default must be 50%%, got %v", got)
	}
	if got := MeteoritePriceFactor("bad"); got != 0.5 {
		t.Errorf("a non-positive percentage must fall back to the default, got %v", got)
	}
}
