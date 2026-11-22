package model

import (
	"strings"
	"testing"
)

const testSealKey = "0123456789abcdef0123456789abcdef0123456789abcdef"

func TestMeteoriteSealRoundTripAndDeterminism(t *testing.T) {
	t.Setenv("METEORITE_SEAL_KEY", testSealKey)
	key := "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKL"
	first, err := MeteoriteSeal(key)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(first, key) || !strings.HasPrefix(first, "mk1:") {
		t.Fatalf("the sealed value must not contain the key: %q", first)
	}
	second, _ := MeteoriteSeal(key)
	if first != second {
		t.Fatal("the same key must seal to the same text so lookups by key work")
	}
	if len(first) > 128 {
		t.Fatalf("a sealed 48-character key must fit the 128-character column, got %d", len(first))
	}
	if got := MeteoriteOpen(first); got != key {
		t.Fatalf("opened %q", got)
	}
	again, _ := MeteoriteSeal(first)
	if again != first {
		t.Fatal("sealing a sealed value must not seal it twice")
	}
}

func TestMeteoriteSealRefusesWithoutAKeyAndReadsOldPlainValues(t *testing.T) {
	t.Setenv("METEORITE_SEAL_KEY", "")
	if _, err := MeteoriteSeal("secret"); err == nil {
		t.Fatal("sealing without METEORITE_SEAL_KEY must fail, never store plain text")
	}
	if got := MeteoriteOpen("old-plain-key"); got != "old-plain-key" {
		t.Fatalf("an older plain value is read as it is, got %q", got)
	}
}

func TestMeteoriteOpenRejectsTamperedValues(t *testing.T) {
	t.Setenv("METEORITE_SEAL_KEY", testSealKey)
	sealed, _ := MeteoriteSeal("secret-key-value")
	tampered := sealed[:len(sealed)-2] + "AA"
	if got := MeteoriteOpen(tampered); got != "" {
		t.Fatalf("a tampered value must open to nothing, got %q", got)
	}
}
