package database

import (
	"github.com/stretchr/testify/mock"
)

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) GetDevice(zigbeeID string) (Device, bool, error) {
	args := m.Called(zigbeeID)
	return args.Get(0).(Device), args.Bool(1), args.Error(2)
}

func (m *MockDatabase) AddDevice(zigbeeID, liquidAddress, deviceName, deviceType, planetmintAddress string) error {
	args := m.Called(zigbeeID, liquidAddress, deviceName, deviceType, planetmintAddress)
	return args.Error(0)
}

func (m *MockDatabase) ExistsID(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockDatabase) GetAllDevices() (map[string]Device, error) {
	args := m.Called()
	return args.Get(0).(map[string]Device), args.Error(1)
}

func (m *MockDatabase) GetByLiquidAddress(liquidAddress string) (map[string]Device, error) {
	args := m.Called(liquidAddress)
	return args.Get(0).(map[string]Device), args.Error(1)
}

func (m *MockDatabase) SetReportStatus(zigbeeID, date, status string) error {
	args := m.Called(zigbeeID, date, status)
	return args.Error(0)
}

func (m *MockDatabase) GetReportStatus(zigbeeID, date string) (string, error) {
	args := m.Called(zigbeeID, date)
	return args.String(0), args.Error(1)
}
