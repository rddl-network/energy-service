package planetmint

import (
	"github.com/stretchr/testify/mock"
)

type MockPlanetmintClient struct {
	mock.Mock
}

func (m *MockPlanetmintClient) IsZigbeeRegistered(zigbeeID string) (bool, error) {
	args := m.Called(zigbeeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPlanetmintClient) RegisterDER(zigbeeID, planetmintAddress, liquidAddress, metadataJson string) error {
	args := m.Called(zigbeeID, planetmintAddress, liquidAddress, metadataJson)
	return args.Error(0)
}
