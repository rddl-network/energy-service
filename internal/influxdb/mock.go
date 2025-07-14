package influxdb

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	args := m.Called(ctx, measurement, tags, fields, ts)
	return args.Error(0)
}

func (m *MockClient) GetLastPoint(ctx context.Context, measurement string, tags map[string]string) (*LastPointResult, error) {
	args := m.Called(ctx, measurement, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LastPointResult), args.Error(1)
}
