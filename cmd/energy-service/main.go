package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/planetmint/planetmint-go/app"
	"github.com/planetmint/planetmint-go/lib"
	"github.com/rddl-network/energy-service/internal/config"
	"github.com/rddl-network/energy-service/internal/database"
	"github.com/rddl-network/energy-service/internal/planetmint"
	"github.com/rddl-network/energy-service/internal/server"

	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api"
	"github.com/influxdata/influxdb-client-go/api/write"

	_ "embed"
	"os"
)

// Adapter to match the server's expected WritePoint interface
type InfluxWriteAPIAdapter struct {
	api api.WriteAPIBlocking
}

func (a *InfluxWriteAPIAdapter) WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts interface{}) error {
	p := write.NewPoint(measurement, tags, fields, ts.(time.Time))
	return a.api.WritePoint(ctx, p)
}

//go:embed static/rddl-sidepane.png
var rddlSidepanePNG []byte

//go:embed templates/index.html
var indexHTML []byte

func writeContentToFiles() {
	// Create templates directory
	if err := os.MkdirAll("templates", 0755); err != nil {
		log.Fatalf("Failed to create folder templates: %v", err)
	}

	// Create static directory
	if err := os.MkdirAll("static", 0755); err != nil {
		log.Fatalf("Failed to create folder static: %v", err)
	}

	// Write embedded PNG to disk
	err := os.WriteFile("static/rddl-sidepane.png", rddlSidepanePNG, 0644)
	if err != nil {
		log.Fatalf("Failed to write rddl-sidepane.png: %v", err)
	}

	err = os.WriteFile("templates/index.html", indexHTML, 0644)
	if err != nil {
		log.Fatalf("Failed to write index.html: %v", err)
	}
}

var libConfig *lib.Config

func init() {
	encodingConfig := app.MakeEncodingConfig()
	libConfig = lib.GetConfig()
	libConfig.SetEncodingConfig(encodingConfig)
}

func main() {
	// Create templates
	writeContentToFiles()

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

	libConfig.SetChainID(cfg.Planetmint.ChainID)
	grpcConn, err := planetmint.SetupGRPCConnection(cfg)
	if err != nil {
		log.Fatalf("Connection to Planetmint failed: %v", err)
	}
	plmntClient := planetmint.NewPlanetmintClient(cfg.Planetmint.Actor, grpcConn)

	// Create and configure server
	db, err := database.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	srv, err := server.NewServer(plmntClient, adapter, db)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close() // Ensure database is closed properly

	mux := http.NewServeMux()
	srv.Routes(mux)

	// Start the server
	log.Println("Server starting on http://localhost:" + strconv.Itoa(cfg.Server.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Server.Port), mux))
}
