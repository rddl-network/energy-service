package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rddl-network/energy-service/internal/config"
	"github.com/rddl-network/energy-service/internal/database"
)

// handleGetDevices returns all devices in the database
func (s *Server) handleGetDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Password protection using config
	cfgPwd := ""
	if config.GetConfig() != nil {
		cfgPwd = config.GetConfig().Server.Password
	}
	pwd := r.URL.Query().Get("pwd")
	if cfgPwd == "" || pwd != cfgPwd {
		http.Error(w, "Unauthorized: missing or incorrect password", http.StatusUnauthorized)
		return
	}

	devices, err := s.db.GetAllDevices()
	if err != nil {
		sendJSONResponse(w, Response{Error: "Failed to retrieve devices"}, http.StatusInternalServerError)
		return
	}

	// Optionally, you could transform the map to a slice if you want to control the order or add extra fields
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(devices)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Failed to encode devices"}, http.StatusInternalServerError)
		return
	}
}

// handleGetDevice handles requests for specific devices
func (s *Server) handleGetDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract Zigbee ID or liquid address from URL
	path := r.URL.Path[len("/api/devices/"):]

	if path == "" {
		sendJSONResponse(w, Response{Error: "Invalid device identifier"}, http.StatusBadRequest)
		return
	}

	// Handle liquid address lookup
	if len(path) > 7 && path[:7] == "liquid/" {
		liquidAddress := path[7:]
		devices, err := s.db.GetByLiquidAddress(liquidAddress)
		if err != nil {
			sendJSONResponse(w, Response{Error: "Failed to retrieve devices"}, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(devices)
		if err != nil {
			sendJSONResponse(w, Response{Error: "Failed to encode devices"}, http.StatusInternalServerError)
			return
		}
		return
	}

	// Handle Zigbee ID lookup
	device, exists, err := s.db.GetDevice(path)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Database error"}, http.StatusInternalServerError)
		return
	}

	if !exists {
		sendJSONResponse(w, Response{Error: "Device not found"}, http.StatusNotFound)
		return
	}

	result := make(map[string]database.Device)
	result[path] = device

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Printf("Failed to encode devices %v", err.Error())
	}
}
