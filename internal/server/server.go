package server

import (
	"html/template"
	"log"
	"net/http"

	"github.com/rddl-network/logger-service/internal/database"
	"github.com/rddl-network/logger-service/internal/utils"
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
