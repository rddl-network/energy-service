package server

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/rddl-network/energy-service/internal/database"
	service "github.com/rddl-network/energy-service/internal/planetmint"
	"github.com/rddl-network/energy-service/internal/utils"
)

// Response represents API response format
type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Server represents the web server
// db is now DeviceStore (interface)
type Server struct {
	db                  database.DeviceStore
	utils               *utils.Utils
	energyDataFileMutex sync.Mutex
	influxWriteAPI      interface {
		WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts interface{}) error
	}
	plmntClient service.IPlanetmintClient
}

// NewServer creates a new server instance, now accepts influxWriteAPI and DeviceStore
func NewServer(
	plmntClient service.IPlanetmintClient,
	writeAPI interface {
		WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts interface{}) error
	},
	db database.DeviceStore, // <-- new param
) (*Server, error) {
	return &Server{
		db:             db,
		utils:          &utils.Utils{},
		influxWriteAPI: writeAPI,
		plmntClient:    plmntClient,
	}, nil
}

// NewDefaultServer creates a new server with a real database (for production)
func NewDefaultServer(
	plmntClient service.IPlanetmintClient,
	writeAPI interface {
		WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts interface{}) error
	},
) (*Server, error) {
	db, err := database.NewDatabase()
	if err != nil {
		return nil, err
	}
	return &Server{
		db:             db,
		utils:          &utils.Utils{},
		influxWriteAPI: writeAPI,
		plmntClient:    plmntClient,
	}, nil
}

// Close shuts down the server and closes the database if possible
func (s *Server) Close() {
	if closer, ok := s.db.(interface{ Close() }); ok {
		closer.Close()
	}
}

// Routes sets up the HTTP routes for the server
func (s *Server) Routes(mux *http.ServeMux) {
	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Main page
	mux.HandleFunc("/", s.handleIndex)

	// API endpoints
	mux.HandleFunc("/register", s.handleRegister)
	mux.HandleFunc("/api/devices", s.handleGetDevices)
	mux.HandleFunc("/api/devices/", s.handleGetDevice)
	mux.HandleFunc("/api/energy", s.handleEnergyData)
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
