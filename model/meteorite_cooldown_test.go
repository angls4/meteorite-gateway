package model

import (
	"testing"
	"time"
)

func TestMeteoriteCooldownDropsAChannelForSixtySeconds(t *testing.T) {
	MeteoriteCoolDownChannel(901)
	kept := meteoriteDropCooledDownIDs([]int{901, 902})
	if len(kept) != 1 || kept[0] != 902 {
		t.Fatalf("a cooled-down channel must be dropped, got %v", kept)
	}
	meteoriteCooledUntilMu.Lock()
	meteoriteCooledUntil[901] = time.Now().Add(-time.Second)
	meteoriteCooledUntilMu.Unlock()
	if kept := meteoriteDropCooledDownIDs([]int{901, 902}); len(kept) != 2 {
		t.Fatalf("after the cooldown the channel must be served again, got %v", kept)
	}
}

func TestMeteoriteCooldownIsSixtySeconds(t *testing.T) {
	if MeteoriteCooldown != 60*time.Second {
		t.Fatalf("cooldown is %v, want 60s", MeteoriteCooldown)
	}
}
