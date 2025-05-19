package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rddl-network/logger-service/internal/config"
	"github.com/rddl-network/logger-service/internal/influxdb"
	"github.com/rddl-network/logger-service/internal/model"
	"github.com/rddl-network/logger-service/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleEnergyData(t *testing.T) {
	// Ensure we load app.toml from the module root
	_, err := config.LoadConfig("app.toml")
	assert.NoError(t, err, "Failed to load configuration")
	_ = config.GetConfig()

	// Set up mock InfluxDB client
	mockInflux := &influxdb.MockClient{}
	mockInflux.On("WritePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockInflux.On("Close").Return()

	srv, err := server.NewServer(mockInflux)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	// Register routes
	srv.Routes()

	// Create a sample energy data payload
	payload := model.EnergyData{
		Version:  1,
		ZigbeeID: "12345",
		Date:     "2025-05-14",
		Data:     [96]float64{1, 2, 3, 4, 5},
	}

	// Marshal the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create a POST request to the /api/energy endpoint
	req, err := http.NewRequest("POST", "/api/energy", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Serve the request using the default mux
	http.DefaultServeMux.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handleEnergyData returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Optionally, verify the server's internal state or response if needed
}
