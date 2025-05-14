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

func TestMapEnergyDataToTime(t *testing.T) {
	utils := &Utils{}
	data := [96]float64{}
	for i := 0; i < 96; i++ {
		data[i] = float64(i)
	}

	result := utils.MapEnergyDataToTime(data)

	if len(result) != 96 {
		t.Fatalf("Expected 96 entries, got %d", len(result))
	}

	expectedTimes := []string{
		"00:15", "00:30", "00:45", "01:00", "01:15", "01:30", "01:45",
		"02:00", "02:15", "02:30", "02:45", "03:00", "03:15", "03:30", "03:45",
		"04:00", "04:15", "04:30", "04:45", "05:00", "05:15", "05:30", "05:45",
		"06:00", "06:15", "06:30", "06:45", "07:00", "07:15", "07:30", "07:45",
		"08:00", "08:15", "08:30", "08:45", "09:00", "09:15", "09:30", "09:45",
		"10:00", "10:15", "10:30", "10:45", "11:00", "11:15", "11:30", "11:45",
		"12:00", "12:15", "12:30", "12:45", "13:00", "13:15", "13:30", "13:45",
		"14:00", "14:15", "14:30", "14:45", "15:00", "15:15", "15:30", "15:45",
		"16:00", "16:15", "16:30", "16:45", "17:00", "17:15", "17:30", "17:45",
		"18:00", "18:15", "18:30", "18:45", "19:00", "19:15", "19:30", "19:45",
		"20:00", "20:15", "20:30", "20:45", "21:00", "21:15", "21:30", "21:45",
		"22:00", "22:15", "22:30", "22:45", "23:00", "23:15", "23:30", "23:45",
		"00:00",
	}

	for i, entry := range result {
		time, ok := entry[0].(string)
		if !ok {
			t.Fatalf("Expected time to be a string, got %T", entry[0])
		}

		if time != expectedTimes[i] {
			t.Errorf("Expected time %s, got %s", expectedTimes[i], time)
		}

		value, ok := entry[1].(float64)
		if !ok {
			t.Fatalf("Expected value to be a float64, got %T", entry[1])
		}

		if value != float64(i) {
			t.Errorf("Expected value %f, got %f", float64(i), value)
		}
	}
}
