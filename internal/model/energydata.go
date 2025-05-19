package model

type EnergyData struct {
	Version  int         `json:"version"`
	ZigbeeID string      `json:"zigbee_id"`
	Date     string      `json:"date"`
	Data     [96]float64 `json:"data"`
}
