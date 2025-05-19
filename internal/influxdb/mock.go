package influxdb

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts interface{}) error {
	args := m.Called(ctx, measurement, tags, fields, ts)
	return args.Error(0)
}
