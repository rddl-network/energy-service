package utils

import (
	"fmt"
	"regexp"
	"time"
)

// Utils provides utility functions
type Utils struct{}

// IsValidID validates for valid SHA256 hash
func (u *Utils) IsValidID(id string) bool {
	// Valid IDs: valid SHA256 hash
	sha256Pattern := regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
	return sha256Pattern.MatchString(id)
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
func (u *Utils) CreateTimestamp(date string, hour int, minute int) (time.Time, error) {
	// Parse the date string into a time.Time object
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Now(), fmt.Errorf("invalid date format: %v", err)
	}
	day := parsedDate.Day()
	if hour == 0 && minute == 0 {
		day = day + 1
	}
	// Combine the date with the provided hour and minute
	timestamp := time.Date(
		parsedDate.Year(),
		parsedDate.Month(),
		day,
		hour,
		minute,
		0, // seconds
		0, // nanoseconds
		time.UTC,
	)

	return timestamp, nil
}
