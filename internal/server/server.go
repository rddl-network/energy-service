package server

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/rddl-network/logger-service/internal/database"
	"github.com/rddl-network/logger-service/internal/model"
	"github.com/rddl-network/logger-service/internal/utils"
)

// Response represents API response format
type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Server represents the web server
type Server struct {
	db                  *database.Database
	utils               *utils.Utils
	energyDataFileMutex sync.Mutex
	influxWriteAPI      interface {
		WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts interface{}) error
	}
}

// NewServer creates a new server instance, now accepts influxWriteAPI
func NewServer(writeAPI interface {
	WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts interface{}) error
}) (*Server, error) {
	db, err := database.NewDatabase()
	if err != nil {
		return nil, err
	}

	return &Server{
		db:             db,
		utils:          &utils.Utils{},
		influxWriteAPI: writeAPI,
	}, nil
}

// SetInfluxWriteAPI allows injecting a custom InfluxDB write API (for testing/mocking)
func (s *Server) SetInfluxWriteAPI(writeAPI interface {
	WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts interface{}) error
}) {
	s.influxWriteAPI = writeAPI
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

	var energyData model.EnergyData

	if err := json.NewDecoder(r.Body).Decode(&energyData); err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received energy data: %+v", energyData)

	go s.writeJSON2File(energyData)
	err := s.write2InfluxDB(energyData)
	if err != nil {
		http.Error(w, "Failed to write to InfluxDB", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Energy data received and written to InfluxDB successfully"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
