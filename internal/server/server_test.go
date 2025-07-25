package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rddl-network/energy-service/internal/config"
	"github.com/rddl-network/energy-service/internal/database"
	"github.com/rddl-network/energy-service/internal/influxdb"
	"github.com/rddl-network/energy-service/internal/model"
	"github.com/rddl-network/energy-service/internal/planetmint"
	"github.com/rddl-network/energy-service/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleEnergyData(t *testing.T) {
	// Ensure we load app.toml from the module root
	_, err := config.LoadConfig("app.toml")
	assert.NoError(t, err, "Failed to load configuration")
	_ = config.GetConfig()

	// Set up mocks
	plmntMock := &planetmint.MockPlanetmintClient{}
	influxMock := &influxdb.MockClient{}
	dbMock := &database.MockDatabase{}
	influxMock.On("WritePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	influxMock.On("Close").Return()
	plmntMock.On("IsZigbeeRegistered", "12345").Return(true, nil)
	// Add mock expectation for SetReportStatus with 'invalid' since the test data is not fully increasing
	dbMock.On("SetReportStatus", "12345", "2025-05-14", "invalid").Return(nil)
	// Add mock expectation for GetReportStatus (no report exists yet)
	dbMock.On("GetReportStatus", "12345", "2025-05-14").Return("", nil)
	influxMock.On("GetLastPoint", mock.Anything, mock.Anything, mock.Anything).Return(&influxdb.LastPointResult{
		Fields:    map[string]interface{}{"kW/h": 0.0},
		Tags:      map[string]string{"zigbee_id": "12345"},
		Timestamp: time.Now().UTC(),
	}, nil)

	srv, err := server.NewServer(plmntMock, influxMock, dbMock)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	mux := http.NewServeMux()
	srv.Routes(mux)

	// Create a sample energy data payload
	var data [96]model.EnergyTuple
	for i := 0; i < 10; i++ {
		data[i] = model.EnergyTuple{
			Value:     float64(i + 1),
			Timestamp: model.TimeStamp(time.Now().UTC()), // Use a fixed or test value if needed
		}
	}
	payload := model.EnergyData{
		Version:      1,
		ZigbeeID:     "12345",
		Date:         "2025-05-14",
		TimezoneName: "Vienna/Europe",
		Data:         data,
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

	// Serve the request using the test mux
	mux.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handleEnergyData returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Optionally, verify the server's internal state or response if needed
}

func TestHandleEnergyData_ValidIncreasing(t *testing.T) {
	_, err := config.LoadConfig("app.toml")
	assert.NoError(t, err, "Failed to load configuration")
	_ = config.GetConfig()

	plmntMock := &planetmint.MockPlanetmintClient{}
	influxMock := &influxdb.MockClient{}
	dbMock := &database.MockDatabase{}
	influxMock.On("WritePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	influxMock.On("Close").Return()
	plmntMock.On("IsZigbeeRegistered", "incrid").Return(true, nil)
	dbMock.On("SetReportStatus", "incrid", "2025-06-04", "valid").Return(nil)
	dbMock.On("GetReportStatus", "incrid", "2025-06-04").Return("", nil)
	influxMock.On("GetLastPoint", mock.Anything, mock.Anything, mock.Anything).Return(&influxdb.LastPointResult{
		Fields:    map[string]interface{}{"kW/h": 0.0},
		Tags:      map[string]string{"zigbee_id": "incrid"},
		Timestamp: time.Now().UTC(),
	}, nil)
	srv, err := server.NewServer(plmntMock, influxMock, dbMock)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	mux := http.NewServeMux()
	srv.Routes(mux)

	var increasing [96]model.EnergyTuple
	for i := 0; i < 96; i++ {
		increasing[i] = model.EnergyTuple{
			Value:     float64(i),
			Timestamp: model.TimeStamp(time.Now().UTC()),
		}
	}
	payload := model.EnergyData{
		Version:      1,
		ZigbeeID:     "incrid",
		Date:         "2025-06-04",
		TimezoneName: "Vienna/Europe",
		Data:         increasing,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", "/api/energy", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handleEnergyData (increasing) returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	assert.Contains(t, rr.Body.String(), "Energy data received and written to database successfully")
}

func TestHandleEnergyData_AlreadyExists(t *testing.T) {
	_, err := config.LoadConfig("app.toml")
	assert.NoError(t, err, "Failed to load configuration")
	_ = config.GetConfig()

	plmntMock := &planetmint.MockPlanetmintClient{}
	influxMock := &influxdb.MockClient{}
	dbMock := &database.MockDatabase{}
	influxMock.On("WritePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	influxMock.On("Close").Return()
	plmntMock.On("IsZigbeeRegistered", "dupeid").Return(true, nil)
	dbMock.On("GetReportStatus", "dupeid", "2025-06-05").Return("valid", nil)
	influxMock.On("GetLastPoint", mock.Anything, mock.Anything, mock.Anything).Return(&influxdb.LastPointResult{
		Fields:    map[string]interface{}{"kW/h": 0.0},
		Tags:      map[string]string{"zigbee_id": "dupeid"},
		Timestamp: time.Now().UTC(),
	}, nil)

	srv, err := server.NewServer(plmntMock, influxMock, dbMock)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	mux := http.NewServeMux()
	srv.Routes(mux)

	var increasing [96]model.EnergyTuple
	for i := 0; i < 96; i++ {
		increasing[i] = model.EnergyTuple{
			Value:     float64(i),
			Timestamp: model.TimeStamp(time.Now().UTC()),
		}
	}
	payload := model.EnergyData{
		Version:      1,
		ZigbeeID:     "dupeid",
		Date:         "2025-06-05",
		TimezoneName: "Vienna/Europe",
		Data:         increasing,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", "/api/energy", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("handleEnergyData (already exists) returned wrong status code: got %v want %v", status, http.StatusConflict)
	}
	assert.Contains(t, rr.Body.String(), "already exists")
}
