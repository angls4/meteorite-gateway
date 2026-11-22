package service

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
)

func TestMeteoriteChargedInputTokens(t *testing.T) {
	cases := []struct {
		name                string
		reported, injection int
		want                int
	}{
		{"deduction applies", 1000, 120, 880},
		{"never below the floor of 10", 50, 120, 10},
		{"never above the reported count", 5, 120, 5},
		{"no injection", 1000, 0, 1000},
		{"negative injection is ignored", 1000, -12, 1000},
		{"exactly at the floor", 130, 120, 10},
		{"nothing reported", 0, 120, 0},
	}
	for _, c := range cases {
		if got := MeteoriteChargedInputTokens(c.reported, c.injection); got != c.want {
			t.Errorf("%s: charged %d, want %d", c.name, got, c.want)
		}
	}
}

func TestMeteoriteRelayErrorStatusIs402ForNoBalance(t *testing.T) {
	noBalance := types.NewErrorWithStatusCode(errors.New("test"), types.ErrorCodeInsufficientUserQuota, 403)
	if got := MeteoriteRelayErrorStatus(noBalance); got != 402 {
		t.Fatalf("no balance answered %d, want 402", got)
	}
	other := types.NewErrorWithStatusCode(errors.New("test"), types.ErrorCodeBadResponseStatusCode, 502)
	if got := MeteoriteRelayErrorStatus(other); got != 502 {
		t.Fatalf("other error changed to %d", got)
	}
}

func TestMeteoriteAdjustUsageKeepsReportedUsageUntouched(t *testing.T) {
	usage := &dto.Usage{PromptTokens: 1000, CompletionTokens: 300, TotalTokens: 1300}
	if got, note := meteoriteAdjustUsage(nil, usage); got != usage || note != nil {
		t.Fatal("without a channel the usage must pass through")
	}
}
