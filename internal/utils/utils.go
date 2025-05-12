package utils

import "regexp"

// Utils provides utility functions
type Utils struct{}

// IsValidZigbeeID validates Zigbee ID format
func (u *Utils) IsValidZigbeeID(id string) bool {
	// Zigbee IDs are typically 16 hexadecimal characters
	regex := regexp.MustCompile(`^[0-9A-Fa-f]{16}$`)
	return regex.MatchString(id)
}
