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
		ZigbeeID          string `json:"zigbee_id"`
		LiquidAddress     string `json:"liquid_address"`
		DeviceName        string `json:"device_name"`
		PlanetmintAddress string `json:"planetmint_address"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&formData)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Invalid JSON data"}, http.StatusBadRequest)
		return
	}

	zigbeeID := formData.ZigbeeID
	liquidAddress := formData.LiquidAddress
	plmntAddress := formData.PlanetmintAddress
	deviceName := formData.DeviceName

	metadataJson := "{ \"Device\": \"}" + deviceName + "\"}"

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
	_, existsDB, err := s.db.GetDevice(zigbeeID)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Database error"}, http.StatusInternalServerError)
		return
	}
	existsPlmnt, err := s.plmntClient.IsZigbeeRegistered(zigbeeID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {

		} else {
			sendJSONResponse(w, Response{Error: "Database error"}, http.StatusInternalServerError)
			return
		}
	}

	if existsDB || existsPlmnt {
		sendJSONResponse(w, Response{Error: fmt.Sprintf("Zigbee ID %s already exists", zigbeeID)}, http.StatusBadRequest)
		return
	}

	// Add device to database
	err = s.db.AddDevice(zigbeeID, liquidAddress, deviceName)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Failed to add device"}, http.StatusInternalServerError)
		return
	}
	err = s.plmntClient.RegisterDER(zigbeeID, plmntAddress, liquidAddress, metadataJson)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Failed to attest device to Planetmint"}, http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, Response{Message: fmt.Sprintf("Device %s registered successfully", deviceName)}, http.StatusCreated)
}
