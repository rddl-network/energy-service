package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/planetmint/planetmint-go/app"
	"github.com/planetmint/planetmint-go/lib"
	"github.com/rddl-network/energy-service/internal/config"
	"github.com/rddl-network/energy-service/internal/database"
	"github.com/rddl-network/energy-service/internal/influxdb"
	"github.com/rddl-network/energy-service/internal/planetmint"
	"github.com/rddl-network/energy-service/internal/server"

	influxdb2 "github.com/influxdata/influxdb-client-go"

	"context"
	_ "embed"
	"os"
)

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
	influxClient := influxdb.NewLocalInfluxClient(client, cfg.InfluxDB.Org, cfg.InfluxDB.Bucket)

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
	srv, err := server.NewServer(plmntClient, influxClient, db)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close() // Ensure database is closed properly

	mux := http.NewServeMux()
	srv.Routes(mux)

	// Connectivity check: simple test query to verify bucket/org
	// Simple test query: list measurements in the bucket
	testQuery := `import "influxdata/influxdb/schema"
schema.measurements(bucket: "` + cfg.InfluxDB.Bucket + `")`
	queryAPI := client.QueryAPI(cfg.InfluxDB.Org)
	result, err := queryAPI.Query(context.Background(), testQuery)
	if err != nil {
		log.Fatalf("InfluxDB test query failed: %v", err)
	}
	log.Println("InfluxDB connectivity and test query succeeded. Measurements:")
	for result.Next() {
		log.Println(result.Record().Value())
	}
	if result.Err() != nil {
		log.Fatalf("InfluxDB test query result error: %v", result.Err())
	}

	// Start the server
	log.Println("Server starting on http://localhost:" + strconv.Itoa(cfg.Server.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Server.Port), mux))
}
