package influxdb

import (
	"context"
	"time"
)

type Point interface{}

type LastPointResult struct {
	Timestamp time.Time
	Fields    map[string]interface{}
	Tags      map[string]string
}

type Client interface {
	WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error
	GetLastPoint(ctx context.Context, measurement string, tags map[string]string) (*LastPointResult, error)
}
