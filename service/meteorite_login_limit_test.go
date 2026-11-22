package service

import (
	"testing"
	"time"
)

func TestMeteoriteLoginLimiterBlocksTheSixthAttemptForFifteenMinutes(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter := newMeteoriteLoginLimiter(func() time.Time { return now })
	for i := 0; i < 5; i++ {
		if limiter.Blocked("Ann", "1.2.3.4") {
			t.Fatalf("attempt %d must not be blocked yet", i+1)
		}
		limiter.Failed("Ann", "1.2.3.4")
		now = now.Add(time.Minute)
	}
	if !limiter.Blocked("ann", "1.2.3.4") {
		t.Fatal("the sixth attempt within 15 minutes must be refused")
	}
	if limiter.Blocked("ann", "5.6.7.8") || limiter.Blocked("bob", "1.2.3.4") {
		t.Fatal("the block is for one account and one address only")
	}
	now = now.Add(16 * time.Minute)
	if limiter.Blocked("ann", "1.2.3.4") {
		t.Fatal("the block ends after 15 minutes")
	}
}

func TestMeteoriteLoginLimiterForgetsFailuresAfterASuccess(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter := newMeteoriteLoginLimiter(func() time.Time { return now })
	for i := 0; i < 4; i++ {
		limiter.Failed("ann", "1.2.3.4")
	}
	limiter.Succeeded("ann", "1.2.3.4")
	limiter.Failed("ann", "1.2.3.4")
	if limiter.Blocked("ann", "1.2.3.4") {
		t.Fatal("failures before a success must not count")
	}
}

func TestMeteoriteLoginLimiterIgnoresOldFailures(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter := newMeteoriteLoginLimiter(func() time.Time { return now })
	for i := 0; i < 4; i++ {
		limiter.Failed("ann", "1.2.3.4")
	}
	now = now.Add(20 * time.Minute)
	limiter.Failed("ann", "1.2.3.4")
	if limiter.Blocked("ann", "1.2.3.4") {
		t.Fatal("failures older than the window must not count")
	}
}
