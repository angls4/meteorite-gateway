package setting

import (
	"sync"

	"github.com/QuantumNous/new-api/common"
)

// meteorite: the price a user pays is a percentage of the model's official price (R-013). The
// official price stays in new-api's own ratio tables; this map holds the admin's percentage per
// model, and a model without an entry pays DefaultMeteoritePricePercent.
const DefaultMeteoritePricePercent = 50.0

var (
	meteoritePricePercent   = map[string]float64{}
	meteoritePricePercentMu sync.RWMutex
)

// MeteoritePriceFactor is the multiplier for a model's price: 0.5 for the default 50%.
func MeteoritePriceFactor(modelName string) float64 {
	meteoritePricePercentMu.RLock()
	defer meteoritePricePercentMu.RUnlock()
	if percent, ok := meteoritePricePercent[modelName]; ok {
		return percent / 100
	}
	return DefaultMeteoritePricePercent / 100
}

// UpdateMeteoritePricePercentByJSONString replaces the map from the option value, for example
// {"claude-sonnet-5-5": 60}. A percentage must be above 0 and at most 1000.
func UpdateMeteoritePricePercentByJSONString(jsonStr string) error {
	parsed := map[string]float64{}
	if jsonStr != "" {
		if err := common.UnmarshalJsonStr(jsonStr, &parsed); err != nil {
			return err
		}
	}
	for name, percent := range parsed {
		if percent <= 0 || percent > 1000 {
			delete(parsed, name)
		}
	}
	meteoritePricePercentMu.Lock()
	meteoritePricePercent = parsed
	meteoritePricePercentMu.Unlock()
	return nil
}

func MeteoritePricePercent2JSONString() string {
	meteoritePricePercentMu.RLock()
	defer meteoritePricePercentMu.RUnlock()
	data, err := common.Marshal(meteoritePricePercent)
	if err != nil {
		return "{}"
	}
	return string(data)
}
