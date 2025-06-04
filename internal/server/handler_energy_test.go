package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rddl-network/energy-service/internal/config"
	"github.com/rddl-network/energy-service/internal/database"
	"github.com/rddl-network/energy-service/internal/influxdb"
	"github.com/rddl-network/energy-service/internal/model"
	"github.com/rddl-network/energy-service/internal/planetmint"
	"github.com/rddl-network/energy-service/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupEnergyTestServer(t *testing.T, plmntMock *planetmint.MockPlanetmintClient, influxMock *influxdb.MockClient, dbMock *database.MockDatabase) (*server.Server, *http.ServeMux) {
	_, err := config.LoadConfig("")
	assert.NoError(t, err, "Failed to load configuration")
	srv, err := server.NewServer(plmntMock, influxMock, dbMock)
	assert.NoError(t, err)
	mux := http.NewServeMux()
	srv.Routes(mux)
	t.Cleanup(func() { srv.Close() })
	return srv, mux
}

func TestHandleEnergyData_InvalidJSON(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	influxMock := &influxdb.MockClient{}
	dbMock := &database.MockDatabase{}
	_, mux := setupEnergyTestServer(t, plmntMock, influxMock, dbMock)

	// Send invalid JSON
	req := httptest.NewRequest("POST", "/api/energy", bytes.NewBuffer([]byte("not a json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to decode JSON")
}

func TestHandleEnergyData_UnregisteredZigbeeID(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	influxMock := &influxdb.MockClient{}
	dbMock := &database.MockDatabase{}
	// Mock IsZigbeeRegistered to return false for any zigbeeID except "registered123"
	plmntMock.On("IsZigbeeRegistered", mock.Anything).Return(false, nil)
	_, mux := setupEnergyTestServer(t, plmntMock, influxMock, dbMock)

	energy := model.EnergyData{
		Version:  1,
		ZigbeeID: "unregistered123",
		Date:     "2025-06-04",
		Data:     [96]float64{},
	}
	body, _ := json.Marshal(energy)
	req := httptest.NewRequest("POST", "/api/energy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "not registered in Planetmint")
}

func TestHandleEnergyData_Valid(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	influxMock := &influxdb.MockClient{}
	dbMock := &database.MockDatabase{}
	// Mock IsZigbeeRegistered to return true for "registered123"
	plmntMock.On("IsZigbeeRegistered", "registered123").Return(true, nil)
	influxMock.On("WritePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	_, mux := setupEnergyTestServer(t, plmntMock, influxMock, dbMock)

	energy := model.EnergyData{
		Version:  1,
		ZigbeeID: "registered123",
		Date:     "2025-06-04",
		Data:     [96]float64{1, 2, 3},
	}
	body, _ := json.Marshal(energy)
	req := httptest.NewRequest("POST", "/api/energy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Energy data received and written to database successfully")
}
