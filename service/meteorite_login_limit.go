package service

import (
	"strings"
	"sync"
	"time"
)

// meteorite: five failed sign-ins for an account from one source address within 15 minutes
// block further sign-ins for that account and address for 15 minutes (R-023). The state is in
// memory and shared by all requests of this process.
const (
	MeteoriteLoginMaxFailures = 5
	MeteoriteLoginWindow      = 15 * time.Minute
)

type meteoriteLoginLimiter struct {
	mu       sync.Mutex
	now      func() time.Time
	failures map[string][]time.Time
	blocked  map[string]time.Time
}

func newMeteoriteLoginLimiter(now func() time.Time) *meteoriteLoginLimiter {
	return &meteoriteLoginLimiter{
		now:      now,
		failures: map[string][]time.Time{},
		blocked:  map[string]time.Time{},
	}
}

var meteoriteLogins = newMeteoriteLoginLimiter(time.Now)

func meteoriteLoginKey(username, ip string) string {
	return strings.ToLower(strings.TrimSpace(username)) + "|" + ip
}

// Blocked reports whether this account and address may not try to sign in now.
func (l *meteoriteLoginLimiter) Blocked(username, ip string) bool {
	key := meteoriteLoginKey(username, ip)
	l.mu.Lock()
	defer l.mu.Unlock()
	until, ok := l.blocked[key]
	if !ok {
		return false
	}
	if l.now().After(until) {
		delete(l.blocked, key)
		delete(l.failures, key)
		return false
	}
	return true
}

// Failed records a failed sign-in and starts the block on the fifth one inside the window.
func (l *meteoriteLoginLimiter) Failed(username, ip string) {
	key := meteoriteLoginKey(username, ip)
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	recent := l.failures[key][:0:0]
	for _, at := range l.failures[key] {
		if now.Sub(at) < MeteoriteLoginWindow {
			recent = append(recent, at)
		}
	}
	recent = append(recent, now)
	l.failures[key] = recent
	if len(recent) >= MeteoriteLoginMaxFailures {
		l.blocked[key] = now.Add(MeteoriteLoginWindow)
	}
}

// Succeeded forgets the failures of this account and address.
func (l *meteoriteLoginLimiter) Succeeded(username, ip string) {
	key := meteoriteLoginKey(username, ip)
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
	delete(l.blocked, key)
}

func MeteoriteLoginBlocked(username, ip string) bool { return meteoriteLogins.Blocked(username, ip) }
func MeteoriteLoginFailed(username, ip string)       { meteoriteLogins.Failed(username, ip) }
func MeteoriteLoginSucceeded(username, ip string)    { meteoriteLogins.Succeeded(username, ip) }
