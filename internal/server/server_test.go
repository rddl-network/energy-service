package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleEnergyData(t *testing.T) {
	// Create a new server instance using NewServer
	srv, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	// Create a sample energy data payload
	payload := struct {
		Version  int         `json:"version"`
		ZigbeeID string      `json:"zigbee_id"`
		Date     string      `json:"date"`
		Data     [96]float64 `json:"data"`
	}{
		Version:  1,
		ZigbeeID: "12345",
		Date:     "2025-05-14T12:00:00Z",
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

	// Call the handleEnergyData method
	handler := http.HandlerFunc(srv.handleEnergyData)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handleEnergyData returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Optionally, verify the server's internal state or response if needed
}
