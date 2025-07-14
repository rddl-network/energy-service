package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rddl-network/energy-service/internal/config"
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

	reportStatus, err := s.db.GetReportStatus(energyData.ZigbeeID, energyData.Date)
	if err != nil {
		log.Printf("Failed to check report status: %v", err)
		sendJSONResponse(w, Response{Error: "Database error"}, http.StatusInternalServerError)
		return
	}
	if reportStatus != "" {
		sendJSONResponse(w, Response{Error: "report for this ZigbeeID and date already exists"}, http.StatusConflict)
		return
	}

	// get first data element from payload
	// compare it against the last registered datapoint of this reporting device
	lastPoints, err := s.influxDBClient.GetLastPoint(context.Background(),
		"energy_data",
		map[string]string{
			"Inspelning": energyData.ZigbeeID,
			"timezone":   energyData.TimezoneName,
		})
	if err != nil {
		log.Printf("Failed to get last point from InfluxDB: %v", err)
		sendJSONResponse(w, Response{Error: "Failed to retrieve last point from database"}, http.StatusInternalServerError)
		return
	}
	// check if data is equal or increased
	if energyData.Data[0].Value < lastPoints.Fields["kW/h"].(float64) {
		sendJSONResponse(w, Response{Error: "Incompatible data: data does not increase."}, http.StatusConflict)
		return
	}

	status := "valid"
	if !model.IsEnergyDataIncreasing(energyData.Data) {
		status = "invalid"
		log.Printf("Energy data for Zigbee ID %s is not increasing", energyData.ZigbeeID)
	}

	err = s.db.SetReportStatus(energyData.ZigbeeID, energyData.Date, status)
	if err != nil {
		log.Printf("Failed to store report status: %v", err)
	}
	if status == "invalid" {
		log.Printf("Energy data for Zigbee ID %s is not compliant", energyData.ZigbeeID)
		sendJSONResponse(w, Response{Error: "data set is not compliant"}, http.StatusBadRequest)
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

// handleDownloadEnergyData serves the energy data JSON file, password protected
func (s *Server) handleDownloadEnergyData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := config.GetConfig()
	cfgPwd := ""
	if cfg != nil {
		cfgPwd = cfg.Server.Password
	}
	pwd := r.URL.Query().Get("pwd")
	if cfgPwd == "" || pwd != cfgPwd {
		http.Error(w, "Unauthorized: missing or incorrect password", http.StatusUnauthorized)
		return
	}
	file, err := os.Open(cfg.Server.DataFile)
	if err != nil {
		http.Error(w, "Failed to open data file", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close data file: %v", err)
		}
	}()
	dec := json.NewDecoder(file)
	var results []interface{}
	for {
		var entry interface{}
		if err := dec.Decode(&entry); err != nil {
			if err.Error() == "EOF" {
				break
			}
			http.Error(w, "Failed to decode data file", http.StatusInternalServerError)
			return
		}
		results = append(results, entry)
	}
	w.Header().Set("Content-Type", "application/json")
	if len(results) == 0 {
		if err := json.NewEncoder(w).Encode([]interface{}{}); err != nil {
			log.Printf("failed to encode empty array: %v", err)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("failed to encode results: %v", err)
	}
}
