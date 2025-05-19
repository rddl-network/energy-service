package influxdb

import (
	"context"
	"time"
)

type Point interface{}

type Client interface {
	WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error
}
