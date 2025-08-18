package model

import (
	"fmt"
	"time"
)

type EnergyTuple struct {
	Value     float64   `json:"value"`
	Timestamp TimeStamp `json:"timestamp"`
}

type TimeStamp time.Time

const timeLayout = "2006-01-02 15:04:05"

func (t TimeStamp) MarshalJSON() ([]byte, error) {
	utc := time.Time(t).UTC()
	return []byte(fmt.Sprintf("\"%s\"", utc.Format(timeLayout))), nil
}

func (t *TimeStamp) UnmarshalJSON(b []byte) error {
	s := string(b)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	tp, err := time.ParseInLocation(timeLayout, s, time.UTC)
	if err != nil {
		return err
	}
	*t = TimeStamp(tp)
	return nil
}

type EnergyData struct {
	Version      int             `json:"version"`
	ID           string          `json:"id"`
	Date         string          `json:"date"`
	TimezoneName string          `json:"timezone_name"`
	Data         [96]EnergyTuple `json:"data"`
}

// IsEnergyDataIncreasing checks if the Data array is monotonically non-decreasing by Value
func IsEnergyDataIncreasing(data [96]EnergyTuple) bool {
	for i := 1; i < len(data); i++ {
		if data[i].Value < data[i-1].Value {
			return false
		}
	}
	return true
}
