package model

import (
	"sync"
	"time"
)

// meteorite: a channel that failed a user request rests for 60 seconds instead of being disabled
// (R-015). Only the verification flow switches a channel off; the cooldown is the gateway's own
// short-term answer to a failing upstream. State is kept in memory: it is lost on restart, which
// at worst sends one more request to a failing channel.
const MeteoriteCooldown = 60 * time.Second

var (
	meteoriteCooledUntil   = map[int]time.Time{}
	meteoriteCooledUntilMu sync.Mutex
)

// MeteoriteCoolDownChannel starts (or restarts) the channel's cooldown.
func MeteoriteCoolDownChannel(channelId int) {
	meteoriteCooledUntilMu.Lock()
	meteoriteCooledUntil[channelId] = time.Now().Add(MeteoriteCooldown)
	meteoriteCooledUntilMu.Unlock()
}

func meteoriteIsCooledDown(channelId int) bool {
	meteoriteCooledUntilMu.Lock()
	defer meteoriteCooledUntilMu.Unlock()
	until, ok := meteoriteCooledUntil[channelId]
	if !ok {
		return false
	}
	if time.Now().After(until) {
		delete(meteoriteCooledUntil, channelId)
		return false
	}
	return true
}

// meteoriteDropCooledDownIDs removes cooled-down channels from a candidate id list.
func meteoriteDropCooledDownIDs(ids []int) []int {
	kept := make([]int, 0, len(ids))
	for _, id := range ids {
		if !meteoriteIsCooledDown(id) {
			kept = append(kept, id)
		}
	}
	return kept
}

func meteoriteDropCooledDownAbilities(abilities []Ability) []Ability {
	kept := make([]Ability, 0, len(abilities))
	for _, ability := range abilities {
		if !meteoriteIsCooledDown(ability.ChannelId) {
			kept = append(kept, ability)
		}
	}
	return kept
}
