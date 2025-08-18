package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// handleRegister handles device registration requests
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// For JSON data
	var formData struct {
		ID                string `json:"id"`
		LiquidAddress     string `json:"liquid_address"`
		DeviceName        string `json:"device_name"`
		PlanetmintAddress string `json:"planetmint_address"`
		DeviceType        string `json:"device_type"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&formData)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Invalid JSON data"}, http.StatusBadRequest)
		return
	}

	id := formData.ID
	liquidAddress := formData.LiquidAddress
	plmntAddress := formData.PlanetmintAddress
	deviceName := formData.DeviceName
	deviceType := formData.DeviceType

	metadataJson := "{ \"Device\": \"}" + deviceName + "\"}"

	// Validate form data
	if id == "" || liquidAddress == "" || deviceName == "" || deviceType == "" || plmntAddress == "" {
		sendJSONResponse(w, Response{Error: "All fields are required"}, http.StatusBadRequest)
		return
	}

	// Validate Zigbee ID format
	if !s.utils.IsValidID(id) {
		sendJSONResponse(w, Response{Error: "Invalid ID format"}, http.StatusBadRequest)
		return
	}

	// Check if Zigbee ID already exists
	_, existsDB, err := s.db.GetDevice(id)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Database error " + err.Error()}, http.StatusInternalServerError)
		return
	}
	existsPlmnt, err := s.plmntClient.IsZigbeeRegistered(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {

		} else {
			sendJSONResponse(w, Response{Error: "Planetmint error " + err.Error()}, http.StatusInternalServerError)
			return
		}
	}

	if existsDB || existsPlmnt {
		sendJSONResponse(w, Response{Error: fmt.Sprintf("ID %s already exists", id)}, http.StatusBadRequest)
		return
	}

	// Add device to database
	err = s.db.AddDevice(id, liquidAddress, deviceName, deviceType, plmntAddress)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Failed to add device"}, http.StatusInternalServerError)
		return
	}
	err = s.plmntClient.RegisterDER(id, plmntAddress, liquidAddress, metadataJson)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Failed to attest device to Planetmint"}, http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, Response{Message: fmt.Sprintf("Device %s registered successfully", deviceName)}, http.StatusCreated)
}
