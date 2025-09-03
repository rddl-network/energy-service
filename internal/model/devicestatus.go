package model

type DeviceStatus struct {
	IsOn                bool     `json:"isOn"`
	CurrentVoltage      float64  `json:"currentVoltage"`
	CurrentAmps         float64  `json:"currentAmps"`
	CurrentActivePower  float64  `json:"currentActivePower"`
	TotalEnergyConsumed *float64 `json:"totalEnergyConsumed"`
}

type DeviceStatusExt struct {
	ID           string       `json:"id"`
	DeviceStatus DeviceStatus `json:"deviceStatus"`
}
