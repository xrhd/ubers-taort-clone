package mlgateway

import (
	"encoding/json"
	"hash/fnv"
	"math"
)

// syntheticPredict generates deterministic predictions for a (userID, leverID) pair.
// Uses FNV hashing for reproducibility across runs.
func syntheticPredict(userID, leverID int64, weights map[string]float64) (cost float64, roi float64, upliftVector json.RawMessage) {
	h := fnv.New64a()
	b := make([]byte, 16)
	b[0] = byte(userID)
	b[1] = byte(userID >> 8)
	b[2] = byte(userID >> 16)
	b[3] = byte(userID >> 24)
	b[4] = byte(userID >> 32)
	b[5] = byte(userID >> 40)
	b[6] = byte(userID >> 48)
	b[7] = byte(userID >> 56)
	b[8] = byte(leverID)
	b[9] = byte(leverID >> 8)
	b[10] = byte(leverID >> 16)
	b[11] = byte(leverID >> 24)
	b[12] = byte(leverID >> 32)
	b[13] = byte(leverID >> 40)
	b[14] = byte(leverID >> 48)
	b[15] = byte(leverID >> 56)
	h.Write(b)
	hash := h.Sum64()

	// Cost: 1.0 to 10.0
	cost = 1.0 + float64(hash%900)/100.0

	// Uplift vector components
	ridesTrips := float64((hash>>16)%100) / 100.0   // 0.0 to 0.99
	eatsOrders := float64((hash>>32)%100) / 100.0   // 0.0 to 0.99

	uplift := map[string]float64{
		"rides_trips": ridesTrips,
		"eats_orders": eatsOrders,
	}

	// ROI = w . x (dot product of strategic weights and uplift vector)
	roi = 0
	for k, w := range weights {
		if v, ok := uplift[k]; ok {
			roi += w * v
		}
	}
	roi = math.Round(roi*1000) / 1000

	upliftBytes, _ := json.Marshal(uplift)
	return cost, roi, json.RawMessage(upliftBytes)
}
