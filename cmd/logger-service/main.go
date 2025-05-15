package main

import (
	"log"
	"net/http"

	"github.com/rddl-network/logger-service/internal/config"
	"github.com/rddl-network/logger-service/internal/server"
)

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

	// Create and configure server
	srv, err := server.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close() // Ensure database is closed properly

	srv.Routes()

	// Start the server
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
