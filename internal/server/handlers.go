package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rddl-network/energy-service/internal/database"
)

// handleRegister handles device registration requests
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// For JSON data
	var formData struct {
		ZigbeeID      string `json:"zigbee_id"`
		LiquidAddress string `json:"liquid_address"`
		DeviceName    string `json:"device_name"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&formData)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Invalid JSON data"}, http.StatusBadRequest)
		return
	}

	zigbeeID := formData.ZigbeeID
	liquidAddress := formData.LiquidAddress
	deviceName := formData.DeviceName

	// Validate form data
	if zigbeeID == "" || liquidAddress == "" || deviceName == "" {
		sendJSONResponse(w, Response{Error: "All fields are required"}, http.StatusBadRequest)
		return
	}

	// Validate Zigbee ID format
	if !s.utils.IsValidZigbeeID(zigbeeID) {
		sendJSONResponse(w, Response{Error: "Invalid Zigbee ID format"}, http.StatusBadRequest)
		return
	}

	// Check if Zigbee ID already exists
	_, exists, err := s.db.GetDevice(zigbeeID)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Database error"}, http.StatusInternalServerError)
		return
	}

	if exists {
		sendJSONResponse(w, Response{Error: fmt.Sprintf("Zigbee ID %s already exists", zigbeeID)}, http.StatusBadRequest)
		return
	}

	// Add device to database
	err = s.db.AddDevice(zigbeeID, liquidAddress, deviceName)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Failed to add device"}, http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, Response{Message: fmt.Sprintf("Device %s registered successfully", deviceName)}, http.StatusCreated)
}

// handleGetDevices returns all devices in the database
func (s *Server) handleGetDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devices, err := s.db.GetAllDevices()
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
