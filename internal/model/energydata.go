package model

type EnergyData struct {
	Version  int         `json:"version"`
	ZigbeeID string      `json:"zigbee_id"`
	Date     string      `json:"date"`
	Data     [96]float64 `json:"data"`
}

// IsEnergyDataIncreasing checks if the Data array is monotonically non-decreasing
func IsEnergyDataIncreasing(data [96]float64) bool {
	for i := 1; i < len(data); i++ {
		if data[i] < data[i-1] {
			return false
		}
	}
	return true
}
