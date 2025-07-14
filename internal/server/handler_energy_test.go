package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
	dbMock.On("SetReportStatus", "unregistered123", "2025-06-04", "valid").Return(nil)
	influxMock.On("WritePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	_, mux := setupEnergyTestServer(t, plmntMock, influxMock, dbMock)

	var data [96]model.EnergyTuple
	for i := 0; i < 96; i++ {
		data[i] = model.EnergyTuple{
			Value:     0,
			Timestamp: model.TimeStamp(time.Now().UTC()),
		}
	}
	energy := model.EnergyData{
		Version:      1,
		ZigbeeID:     "unregistered123",
		Date:         "2025-06-04",
		TimezoneName: "Vienna/Europe",
		Data:         data,
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
	dbMock.On("SetReportStatus", "registered123", "2025-06-04", "valid").Return(nil)
	dbMock.On("GetReportStatus", "registered123", "2025-06-04").Return("", nil)
	influxMock.On("GetLastPoint", mock.Anything, mock.Anything, mock.Anything).Return(&influxdb.LastPointResult{
		Fields:    map[string]interface{}{"kW/h": 0.0},
		Tags:      map[string]string{"zigbee_id": "registered123"},
		Timestamp: time.Now().UTC(),
	}, nil)

	_, mux := setupEnergyTestServer(t, plmntMock, influxMock, dbMock)

	// Use a fully increasing array for valid test
	var increasingData [96]model.EnergyTuple
	for i := 0; i < 96; i++ {
		increasingData[i] = model.EnergyTuple{
			Value:     float64(i),
			Timestamp: model.TimeStamp(time.Now().UTC()),
		}
	}

	energy := model.EnergyData{
		Version:      1,
		ZigbeeID:     "registered123",
		Date:         "2025-06-04",
		TimezoneName: "Vienna/Europe",
		Data:         increasingData,
	}
	body, _ := json.Marshal(energy)
	req := httptest.NewRequest("POST", "/api/energy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Energy data received and written to database successfully")
}

func TestDownloadEnergyData_EmptyFile(t *testing.T) {
	// Setup temp file
	tempFile, err := os.CreateTemp("", "energydata_empty_*.json")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tempFile.Name()) }()
	// Write nothing (empty file)
	if err := tempFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	cfg := config.GetConfig()
	cfg.Server.DataFile = tempFile.Name()
	cfg.Server.Password = "testpwd"

	srv, mux := setupEnergyTestServer(t, &planetmint.MockPlanetmintClient{}, &influxdb.MockClient{}, &database.MockDatabase{})
	defer srv.Close()

	req := httptest.NewRequest("GET", "/api/energy/download?pwd=testpwd", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, "[]\n", rr.Body.String())
}

func TestDownloadEnergyData_InvalidPassword(t *testing.T) {
	tempFile, err := os.CreateTemp("", "energydata_invalidpwd_*.json")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tempFile.Name()) }()
	if err := tempFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	cfg := config.GetConfig()
	cfg.Server.DataFile = tempFile.Name()
	cfg.Server.Password = "testpwd"

	srv, mux := setupEnergyTestServer(t, &planetmint.MockPlanetmintClient{}, &influxdb.MockClient{}, &database.MockDatabase{})
	defer srv.Close()

	req := httptest.NewRequest("GET", "/api/energy/download?pwd=wrongpwd", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")
}

func TestDownloadEnergyData_CrowdedFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "energydata_crowded_*.json")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tempFile.Name()) }()

	var data1, data2 [96]model.EnergyTuple
	for i := 0; i < 96; i++ {
		data1[i] = model.EnergyTuple{
			Value:     1,
			Timestamp: model.TimeStamp(time.Now().UTC()),
		}
		data2[i] = model.EnergyTuple{
			Value:     4,
			Timestamp: model.TimeStamp(time.Now().UTC()),
		}
	}
	entries := []model.EnergyData{
		{Version: 1, ZigbeeID: "id1", Date: "2025-06-04", TimezoneName: "Vienna/Europe", Data: data1},
		{Version: 1, ZigbeeID: "id2", Date: "2025-06-05", TimezoneName: "Vienna/Europe", Data: data2},
	}
	enc := json.NewEncoder(tempFile)
	for _, e := range entries {
		assert.NoError(t, enc.Encode(e))
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	cfg := config.GetConfig()
	cfg.Server.DataFile = tempFile.Name()
	cfg.Server.Password = "testpwd"

	srv, mux := setupEnergyTestServer(t, &planetmint.MockPlanetmintClient{}, &influxdb.MockClient{}, &database.MockDatabase{})
	defer srv.Close()

	req := httptest.NewRequest("GET", "/api/energy/download?pwd=testpwd", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	var arr []map[string]interface{}
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &arr))
	assert.Len(t, arr, 2)
	assert.Equal(t, "id1", arr[0]["zigbee_id"])
	assert.Equal(t, "id2", arr[1]["zigbee_id"])
}

func TestDownloadEnergyData_CorruptedFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "energydata_corrupt_*.json")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tempFile.Name()) }()
	_, err = tempFile.WriteString("not a json line\n")
	assert.NoError(t, err)
	if err := tempFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	cfg := config.GetConfig()
	cfg.Server.DataFile = tempFile.Name()
	cfg.Server.Password = "testpwd"

	srv, mux := setupEnergyTestServer(t, &planetmint.MockPlanetmintClient{}, &influxdb.MockClient{}, &database.MockDatabase{})
	defer srv.Close()

	req := httptest.NewRequest("GET", "/api/energy/download?pwd=testpwd", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to decode data file")
}

func TestHandleEnergyData_IncrementalVsLastPoint(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	influxMock := &influxdb.MockClient{}
	dbMock := &database.MockDatabase{}
	plmntMock.On("IsZigbeeRegistered", "zigbeeInc").Return(true, nil)
	dbMock.On("SetReportStatus", "zigbeeInc", "2025-06-04", "valid").Return(nil)
	dbMock.On("GetReportStatus", "zigbeeInc", "2025-06-04").Return("", nil)
	influxMock.On("WritePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	influxMock.On("GetLastPoint", mock.Anything, mock.Anything, mock.Anything).Return(&influxdb.LastPointResult{
		Fields:    map[string]interface{}{"kW/h": 10.0},
		Tags:      map[string]string{"zigbee_id": "zigbeeInc"},
		Timestamp: time.Now().UTC(),
	}, nil)

	_, mux := setupEnergyTestServer(t, plmntMock, influxMock, dbMock)
	var data [96]model.EnergyTuple
	for i := 0; i < 96; i++ {
		data[i] = model.EnergyTuple{
			Value:     10.0 + float64(i+1), // strictly increasing
			Timestamp: model.TimeStamp(time.Now().UTC()),
		}
	}
	energy := model.EnergyData{
		Version:      1,
		ZigbeeID:     "zigbeeInc",
		Date:         "2025-06-04",
		TimezoneName: "Vienna/Europe",
		Data:         data,
	}
	body, _ := json.Marshal(energy)
	req := httptest.NewRequest("POST", "/api/energy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Energy data received and written to database successfully")
}

func TestHandleEnergyData_EqualVsLastPoint(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	influxMock := &influxdb.MockClient{}
	dbMock := &database.MockDatabase{}
	plmntMock.On("IsZigbeeRegistered", "zigbeeEq").Return(true, nil)
	dbMock.On("SetReportStatus", "zigbeeEq", "2025-06-04", "valid").Return(nil)
	dbMock.On("GetReportStatus", "zigbeeEq", "2025-06-04").Return("", nil)
	influxMock.On("WritePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	influxMock.On("GetLastPoint", mock.Anything, mock.Anything, mock.Anything).Return(&influxdb.LastPointResult{
		Fields:    map[string]interface{}{"kW/h": 10.0},
		Tags:      map[string]string{"zigbee_id": "zigbeeEq"},
		Timestamp: time.Now().UTC(),
	}, nil)

	_, mux := setupEnergyTestServer(t, plmntMock, influxMock, dbMock)
	var data [96]model.EnergyTuple
	for i := 0; i < 96; i++ {
		data[i] = model.EnergyTuple{
			Value:     10.0, // equal to last point
			Timestamp: model.TimeStamp(time.Now().UTC()),
		}
	}
	energy := model.EnergyData{
		Version:      1,
		ZigbeeID:     "zigbeeEq",
		Date:         "2025-06-04",
		TimezoneName: "Vienna/Europe",
		Data:         data,
	}
	body, _ := json.Marshal(energy)
	req := httptest.NewRequest("POST", "/api/energy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	// Expect error or rejection (status code depends on your handler logic)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Energy data received and written to database successfull")
}

func TestHandleEnergyData_EqualVsLastPoint_no_previous_data(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	influxMock := &influxdb.MockClient{}
	dbMock := &database.MockDatabase{}
	plmntMock.On("IsZigbeeRegistered", "zigbeeEq").Return(true, nil)
	dbMock.On("SetReportStatus", "zigbeeEq", "2025-06-04", "valid").Return(nil)
	dbMock.On("GetReportStatus", "zigbeeEq", "2025-06-04").Return("", nil)
	influxMock.On("WritePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	influxMock.On("GetLastPoint", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	_, mux := setupEnergyTestServer(t, plmntMock, influxMock, dbMock)
	var data [96]model.EnergyTuple
	for i := 0; i < 96; i++ {
		data[i] = model.EnergyTuple{
			Value:     10.0, // equal to last point
			Timestamp: model.TimeStamp(time.Now().UTC()),
		}
	}
	energy := model.EnergyData{
		Version:      1,
		ZigbeeID:     "zigbeeEq",
		Date:         "2025-06-04",
		TimezoneName: "Vienna/Europe",
		Data:         data,
	}
	body, _ := json.Marshal(energy)
	req := httptest.NewRequest("POST", "/api/energy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	// Expect error or rejection (status code depends on your handler logic)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Energy data received and written to database successfull")
}

func TestHandleEnergyData_LowerVsLastPoint(t *testing.T) {
	plmntMock := &planetmint.MockPlanetmintClient{}
	influxMock := &influxdb.MockClient{}
	dbMock := &database.MockDatabase{}
	plmntMock.On("IsZigbeeRegistered", "zigbeeLow").Return(true, nil)
	dbMock.On("SetReportStatus", "zigbeeLow", "2025-06-04", "valid").Return(nil)
	dbMock.On("GetReportStatus", "zigbeeLow", "2025-06-04").Return("", nil)
	influxMock.On("WritePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	influxMock.On("GetLastPoint", mock.Anything, mock.Anything, mock.Anything).Return(&influxdb.LastPointResult{
		Fields:    map[string]interface{}{"kW/h": 10.0},
		Tags:      map[string]string{"zigbee_id": "zigbeeLow"},
		Timestamp: time.Now().UTC(),
	}, nil)

	_, mux := setupEnergyTestServer(t, plmntMock, influxMock, dbMock)
	var data [96]model.EnergyTuple
	for i := 0; i < 96; i++ {
		data[i] = model.EnergyTuple{
			Value:     9.0, // lower than last point
			Timestamp: model.TimeStamp(time.Now().UTC()),
		}
	}
	energy := model.EnergyData{
		Version:      1,
		ZigbeeID:     "zigbeeLow",
		Date:         "2025-06-04",
		TimezoneName: "Vienna/Europe",
		Data:         data,
	}
	body, _ := json.Marshal(energy)
	req := httptest.NewRequest("POST", "/api/energy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.NotEqual(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Incompatible data: data does not increase")
}
