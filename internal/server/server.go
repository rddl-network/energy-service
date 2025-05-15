package server

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/rddl-network/logger-service/internal/database"
	"github.com/rddl-network/logger-service/internal/utils"
)

// Global configuration variables for InfluxDB
var (
	InfluxDBURL    string
	InfluxDBToken  string
	InfluxDBOrg    string
	InfluxDBBucket string
)

// Response represents API response format
type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Server represents the web server
type Server struct {
	db    *database.Database
	utils *utils.Utils
}

// NewServer creates a new server instance
func NewServer() (*Server, error) {
	db, err := database.NewDatabase()
	if err != nil {
		return nil, err
	}

	return &Server{
		db:    db,
		utils: &utils.Utils{},
	}, nil
}

// Close shuts down the server and closes the database
func (s *Server) Close() {
	s.db.Close()
}

// Routes sets up the HTTP routes for the server
func (s *Server) Routes() {
	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Main page
	http.HandleFunc("/", s.handleIndex)

	// API endpoints
	http.HandleFunc("/register", s.handleRegister)
	http.HandleFunc("/api/devices", s.handleGetDevices)
	http.HandleFunc("/api/devices/", s.handleGetDevice)
	http.HandleFunc("/api/energy", s.handleEnergyData)
}

// handleIndex renders the main page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Error during template execution %v", err.Error())
	}
}

// handleEnergyData handles POST requests, decodes JSON data, logs it, writes to InfluxDB, and responds with a success message
func (s *Server) handleEnergyData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var energyData struct {
		Version  int         `json:"version"`
		ZigbeeID string      `json:"zigbee_id"`
		Date     string      `json:"date"`
		Data     [96]float64 `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&energyData); err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received energy data: %+v", energyData)

	// Write data to InfluxDB
	client := influxdb2.NewClient(InfluxDBURL, InfluxDBToken)
	defer client.Close()

	writeAPI := client.WriteAPIBlocking(InfluxDBOrg, InfluxDBBucket)

	// Prepare data point

	for i := 0; i < 96; i++ {
		hour, minutes := utils.Index2Time(i)
		ts := utils.CreateTimestamp(energyData.Date, hour, minutes)
		log.Printf("Timestamp: %s", ts)
	}

	// p := influxdb2.NewPoint(
	// "energy_data",
	// map[string]string{"zigbee_id": energyData.ZigbeeID},
	// map[string]interface{}{
	// "version": energyData.Version,
	// "date":    energyData.Date,
	// "data":    energyData.Data,
	// },
	// time.Now(), // Use the current timestamp
	// )

	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		log.Printf("Failed to write to InfluxDB: %v", err)
		http.Error(w, "Failed to write to InfluxDB", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Energy data received and written to InfluxDB successfully"))
}
