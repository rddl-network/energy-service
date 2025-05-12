package utils

import "testing"

func TestIsValidZigbeeID(t *testing.T) {
	utils := &Utils{}

	tests := []struct {
		id       string
		expected bool
	}{
		{"00124B0001A2B3C4", true},   // Valid Zigbee ID
		{"00124B0001A2B3C", false},   // Too short
		{"00124B0001A2B3C4D", false}, // Too long
		{"00124B0001A2B3CZ", false},  // Contains invalid character
		{"", false},                  // Empty string
	}

	for _, test := range tests {
		result := utils.IsValidZigbeeID(test.id)
		if result != test.expected {
			t.Errorf("IsValidZigbeeID(%q) = %v; want %v", test.id, result, test.expected)
		}
	}
}
