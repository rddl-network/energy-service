package main

import (
	"log"
	"net/http"

	"rddl/logger-service/internal/server"
)

func main() {
	// Create templates
	err := server.CreateTemplates()
	if err != nil {
		log.Fatalf("Failed to create templates: %v", err)
	}

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
