package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rddl-network/energy-service/internal/config"
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

func (s *Server) HandleIsDeviceRegistered(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract device ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 || parts[3] == "" {
		sendJSONResponse(w, Response{Error: "Missing device ID"}, http.StatusBadRequest)
		return
	}
	if len(parts) > 4 {
		sendJSONResponse(w, Response{Error: "Invalid device ID"}, http.StatusBadRequest)
		return
	}

	deviceID := parts[3]

	devices, found, err := s.db.GetDevice(deviceID)
	if err != nil || !found {
		sendJSONResponse(w, Response{Error: "Device not found"}, http.StatusNotFound)
		return
	}

	// Optionally, you could transform the map to a slice if you want to control the order or add extra fields
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(devices)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Failed to encode device"}, http.StatusNotFound)
		return
	}
}
