package utils

import (
	"fmt"
	"regexp"
	"time"
)

// Utils provides utility functions
type Utils struct{}

// IsValidZigbeeID validates Zigbee ID format
func (u *Utils) IsValidZigbeeID(id string) bool {
	// Zigbee IDs: 16 hex or UUID with optional _N suffix
	zigbee16 := regexp.MustCompile(`^[0-9A-Fa-f]{16}$`)
	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}(?:_\d+)?$`)
	return zigbee16.MatchString(id) || uuidPattern.MatchString(id)
}

// MapEnergyDataToTime maps a 96-float array to a two-dimensional array with corresponding daytime
func (u *Utils) MapEnergyDataToTime(data [96]float64) [][2]interface{} {
	result := make([][2]interface{}, 96)

	for i, value := range data {
		hour, minute := u.Index2Time(i)
		time := fmt.Sprintf("%02d:%02d", hour, minute)
		result[i] = [2]interface{}{time, value}
	}

	return result
}

// Index2Time converts an index to hour and minute values
func (u *Utils) Index2Time(index int) (int, int) {
	hour := ((index + 1) * 15) / 60
	minute := ((index + 1) * 15) % 60
	if hour == 24 {
		hour = 0
	}
	return hour, minute
}

// CreateTimestamp generates a timestamp from a given date, hour, and minutes
func (u *Utils) CreateTimestamp(date string, hour, minute int) (time.Time, error) {
	// Parse the date string into a time.Time object
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Now(), fmt.Errorf("invalid date format: %v", err)
	}

	// Combine the date with the provided hour and minute
	timestamp := time.Date(
		parsedDate.Year(),
		parsedDate.Month(),
		parsedDate.Day(),
		hour,
		minute,
		0, // seconds
		0, // nanoseconds
		time.UTC,
	)

	return timestamp, nil
}
