package utils

import (
	"fmt"
	"regexp"
)

// Utils provides utility functions
type Utils struct{}

// IsValidZigbeeID validates Zigbee ID format
func (u *Utils) IsValidZigbeeID(id string) bool {
	// Zigbee IDs are typically 16 hexadecimal characters
	regex := regexp.MustCompile(`^[0-9A-Fa-f]{16}$`)
	return regex.MatchString(id)
}

// MapEnergyDataToTime maps a 96-float array to a two-dimensional array with corresponding daytime
func (u *Utils) MapEnergyDataToTime(data [96]float64) [][2]interface{} {
	result := make([][2]interface{}, 96)

	for i, value := range data {
		hour := ((i + 1) * 15) / 60
		minute := ((i + 1) * 15) % 60
		if hour == 24 {
			hour = 0
		}
		time := fmt.Sprintf("%02d:%02d", hour, minute)
		result[i] = [2]interface{}{time, value}
	}

	return result
}
