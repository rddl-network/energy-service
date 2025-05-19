package main

import (
	"log"
	"net/http"

	"github.com/rddl-network/energy-service/internal/config"
	"github.com/rddl-network/energy-service/internal/server"

	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api"
	"github.com/influxdata/influxdb-client-go/api/write"
)

// Adapter to match the server's expected WritePoint interface
type InfluxWriteAPIAdapter struct {
	api api.WriteAPIBlocking
}

func (a *InfluxWriteAPIAdapter) WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts interface{}) error {
	p := write.NewPoint(measurement, tags, fields, ts.(time.Time))
	return a.api.WritePoint(ctx, p)
}

func main() {
	// Create templates
	//err := server.CreateTemplates()
	// if err != nil {
	// log.Fatalf("Failed to create templates: %v", err)
	// }

	// Load configuration
	cfg, err := config.LoadConfig("app.toml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Access configuration
	log.Printf("Server running on port: %d", cfg.Server.Port)
	log.Printf("InfluxDB URL: %s", cfg.InfluxDB.URL)
	client := influxdb2.NewClient(cfg.InfluxDB.URL, cfg.InfluxDB.Token)
	defer client.Close() // Ensure client is closed properly
	writeAPI := client.WriteAPIBlocking(cfg.InfluxDB.Org, cfg.InfluxDB.Bucket)
	adapter := &InfluxWriteAPIAdapter{api: writeAPI}

	// Create and configure server
	srv, err := server.NewServer(adapter)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close() // Ensure database is closed properly

	srv.Routes()

	// Start the server
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
