package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/rddl-network/energy-service/internal/model"
)

// handleEnergyData handles POST requests, decodes JSON data, logs it, writes to InfluxDB, and responds with a success message
func (s *Server) handleEnergyData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var energyData model.EnergyData

	if err := json.NewDecoder(r.Body).Decode(&energyData); err != nil {
		sendJSONResponse(w, Response{Error: "Failed to decode JSON"}, http.StatusBadRequest)
		return
	}

	log.Printf("Received energy data: %+v", energyData)

	existsPlmnt, err := s.plmntClient.IsZigbeeRegistered(energyData.ZigbeeID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			sendJSONResponse(w, Response{Error: "Inspelning not found"}, http.StatusBadRequest)
		} else {
			sendJSONResponse(w, Response{Error: "Database error"}, http.StatusInternalServerError)
			return
		}
	}
	if !existsPlmnt {
		log.Printf("Zigbee ID %s not registered in Planetmint", energyData.ZigbeeID)
		sendJSONResponse(w, Response{Error: "Inspelning not registered in Planetmint"}, http.StatusBadRequest)
		return
	}

	go s.writeJSON2File(energyData)
	err = s.write2InfluxDB(energyData)
	if err != nil {
		sendJSONResponse(w, Response{Error: "Failed to write to database"}, http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, Response{Message: "Energy data received and written to database successfully"}, http.StatusOK)
}
