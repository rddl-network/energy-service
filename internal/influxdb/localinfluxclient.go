package influxdb

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api"
	"github.com/influxdata/influxdb-client-go/api/write"

	_ "embed"
)

// LocalInfluxClient implements the influxdb.Client interface for both writing and querying
// InfluxDB using the Go client.
type LocalInfluxClient struct {
	writeAPI api.WriteAPIBlocking
	queryAPI api.QueryAPI
	org      string
	bucket   string
}

func NewLocalInfluxClient(client influxdb2.Client, org, bucket string) *LocalInfluxClient {
	return &LocalInfluxClient{
		writeAPI: client.WriteAPIBlocking(org, bucket),
		queryAPI: client.QueryAPI(org),
		org:      org,
		bucket:   bucket,
	}
}

func (c *LocalInfluxClient) WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	return c.WritePointTyped(ctx, measurement, tags, fields, ts)
}

// WritePointTyped is the strongly-typed version used internally
func (c *LocalInfluxClient) WritePointTyped(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	p := write.NewPoint(measurement, tags, fields, ts)
	return c.writeAPI.WritePoint(ctx, p)
}

func (c *LocalInfluxClient) GetLastPoint(ctx context.Context, measurement string, tags map[string]string) (*LastPointResult, error) {
	// Compose Flux query to get the last point for the given tags
	flux := `from(bucket: "` + c.bucket + `")` +
		`|> range(start: -30d)` +
		`|> filter(fn: (r) => r["_measurement"] == "` + measurement + `"`
	for k, v := range tags {
		flux += ` and r["` + k + `"] == "` + v + `"`
	}
	flux += `)|> last()`
	result, err := c.queryAPI.Query(ctx, flux)
	if err != nil {
		return nil, err
	}
	var last *LastPointResult
	for result.Next() {
		if last == nil {
			last = &LastPointResult{
				Fields: make(map[string]interface{}),
				Tags:   make(map[string]string),
			}
		}
		last.Timestamp = result.Record().Time()
		last.Fields[result.Record().Field()] = result.Record().Value()
		for k, v := range result.Record().Values() {
			if k != "_value" && k != "_field" && k != "_time" {
				if str, ok := v.(string); ok {
					last.Tags[k] = str
				}
			}
		}
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	if last == nil {
		return nil, nil // No data found
	}
	return last, nil
}
