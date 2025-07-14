package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rddl-network/energy-service/internal/config"
	"github.com/rddl-network/energy-service/internal/model"
)

func (s *Server) writeJSON2File(data model.EnergyData) {
	cfg := config.GetConfig()
	// Store data in a JSON file (append as JSON Lines)
	s.energyDataFileMutex.Lock()
	f, err := os.OpenFile(cfg.Server.DataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open JSON file: %v", err)
	} else {
		enc := json.NewEncoder(f)
		if err := enc.Encode(data); err != nil {
			log.Printf("Failed to write energy data to JSON file: %v", err)
		}
		if err := f.Close(); err != nil {
			log.Printf("Failed to close JSON file: %v", err)
		}
	}
	s.energyDataFileMutex.Unlock()
}

func (s *Server) write2InfluxDB(data model.EnergyData) error {
	writeAPI := s.influxDBClient
	if writeAPI == nil {
		log.Printf("No InfluxDB write API set")
		return nil
	}

	for i := 0; i < 96; i++ {
		err := writeAPI.WritePoint(
			context.Background(),
			"energy_data",
			map[string]string{
				"Inspelning": data.ZigbeeID,
				"timezone":   data.TimezoneName,
			},
			map[string]interface{}{"kW/h": data.Data[i].Value},
			time.Time(data.Data[i].Timestamp),
		)
		if err != nil {
			log.Printf("Failed to write to InfluxDB: %v", err)
			return err
		}
	}
	return nil
}

// sendJSONResponse sends a JSON response with the given status code
func sendJSONResponse(w http.ResponseWriter, resp Response, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Failed to encode devices %v", err.Error())
	}
}
