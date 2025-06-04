package server_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rddl-network/energy-service/internal/config"
	"github.com/rddl-network/energy-service/internal/database"
	"github.com/rddl-network/energy-service/internal/influxdb"
	"github.com/rddl-network/energy-service/internal/planetmint"
	"github.com/rddl-network/energy-service/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRegisterTestServer(t *testing.T, plmntMock *planetmint.MockPlanetmintClient, dbMock *database.MockDatabase) (*server.Server, *http.ServeMux) {
	cfg := config.DefaultConfig()
	config.ConfigTestOnly = &cfg
	mockInflux := &influxdb.MockClient{}
	srv, err := server.NewServer(plmntMock, mockInflux, dbMock)
	assert.NoError(t, err)
	mux := http.NewServeMux()
	srv.Routes(mux)
	t.Cleanup(func() { srv.Close() })
	return srv, mux
}

func TestRegister_InvalidJSON(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	dbMock := &database.MockDatabase{}
	_, mux := setupRegisterTestServer(t, plmntMock, dbMock)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer([]byte("not a json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid JSON data")
}

func TestRegister_MissingFields(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	dbMock := &database.MockDatabase{}
	_, mux := setupRegisterTestServer(t, plmntMock, dbMock)
	form := map[string]interface{}{
		"zigbee_id": "",
		"liquid_address": "",
		"device_name": "",
		"planetmint_address": "",
		"device_type": "",
	}
	body, _ := json.Marshal(form)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "All fields are required")
}

func TestRegister_InvalidZigbeeIDFormat(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	dbMock := &database.MockDatabase{}
	_, mux := setupRegisterTestServer(t, plmntMock, dbMock)
	form := map[string]interface{}{
		"zigbee_id": "badid",
		"liquid_address": "liq1",
		"device_name": "dev1",
		"planetmint_address": "plmnt1",
		"device_type": "type1",
	}
	body, _ := json.Marshal(form)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid Zigbee ID format")
}

func TestRegister_DBError(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	dbMock := &database.MockDatabase{}
	validZigbeeID := "1234567890123456"
	form := map[string]interface{}{
		"zigbee_id": validZigbeeID,
		"liquid_address": "liq1",
		"device_name": "dev1",
		"planetmint_address": "plmnt1",
		"device_type": "type1",
	}
	// Simulate DB error on GetDevice
	dbMock.On("GetDevice", validZigbeeID).Return(database.Device{}, false, errors.New("db error"))
	_, mux := setupRegisterTestServer(t, plmntMock, dbMock)
	body, _ := json.Marshal(form)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Database error")
}

func TestRegister_AlreadyExists(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	dbMock := &database.MockDatabase{}
	validZigbeeID := "1234567890123456"
	form := map[string]interface{}{
		"zigbee_id": validZigbeeID,
		"liquid_address": "liq1",
		"device_name": "dev1",
		"planetmint_address": "plmnt1",
		"device_type": "type1",
	}
	dbMock.On("GetDevice", validZigbeeID).Return(database.Device{}, true, nil)
	plmntMock.On("IsZigbeeRegistered", validZigbeeID).Return(false, nil)
	_, mux := setupRegisterTestServer(t, plmntMock, dbMock)
	body, _ := json.Marshal(form)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "already exists")
}

func TestRegister_PlanetmintError(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	dbMock := &database.MockDatabase{}
	validZigbeeID := "1234567890123456"
	dbMock.On("GetDevice", validZigbeeID).Return(database.Device{}, false, nil)
	plmntMock.On("IsZigbeeRegistered", validZigbeeID).Return(false, assert.AnError)
	_, mux := setupRegisterTestServer(t, plmntMock, dbMock)
	form := map[string]interface{}{
		"zigbee_id": validZigbeeID,
		"liquid_address": "liq1",
		"device_name": "dev1",
		"planetmint_address": "plmnt1",
		"device_type": "type1",
	}
	body, _ := json.Marshal(form)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Database error")
}

func TestRegister_Success(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	dbMock := &database.MockDatabase{}
	validZigbeeID := "1234567890123456"
	dbMock.On("GetDevice", validZigbeeID).Return(database.Device{}, false, nil)
	dbMock.On("AddDevice", validZigbeeID, "liq1", "dev1", "type1", "plmnt1").Return(nil)
	plmntMock.On("IsZigbeeRegistered", validZigbeeID).Return(false, nil)
	plmntMock.On("RegisterDER", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	_, mux := setupRegisterTestServer(t, plmntMock, dbMock)
	form := map[string]interface{}{
		"zigbee_id": validZigbeeID,
		"liquid_address": "liq1",
		"device_name": "dev1",
		"planetmint_address": "plmnt1",
		"device_type": "type1",
	}
	body, _ := json.Marshal(form)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "registered successfully")
}
