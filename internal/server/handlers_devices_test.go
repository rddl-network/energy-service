package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rddl-network/energy-service/internal/config"
	"github.com/rddl-network/energy-service/internal/database"
	"github.com/rddl-network/energy-service/internal/influxdb"
	"github.com/rddl-network/energy-service/internal/planetmint"
	"github.com/rddl-network/energy-service/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setTestConfig(cfg *config.Config) {
	config.ConfigTestOnly = &cfg
}

func setupServerWithPwd(t *testing.T, pwd string) *http.ServeMux {
	cfg, err := config.LoadConfig("")
	assert.NoError(t, err, "Failed to load configuration")
	cfg.Server.Password = pwd
	setTestConfig(cfg)
	mockInflux := &influxdb.MockClient{}
	mockInflux.On("WritePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockInflux.On("Close").Return()
	mockPlmntclient := &planetmint.MockPlanetmintClient{}
	mockDB := &database.MockDatabase{}
	mockDB.On("GetAllDevices").Return(map[string]database.Device{}, nil)
	srv, err := server.NewServer(mockPlmntclient, mockInflux, mockDB)
	assert.NoError(t, err)
	mux := http.NewServeMux()
	srv.Routes(mux)
	t.Cleanup(func() { srv.Close() })
	return mux
}

func TestGetDevices_ValidPassword(t *testing.T) {
	mux := setupServerWithPwd(t, "testpass")
	req := httptest.NewRequest("GET", "/api/devices?pwd=testpass", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetDevices_EmptyPassword(t *testing.T) {
	mux := setupServerWithPwd(t, "testpass")
	req := httptest.NewRequest("GET", "/api/devices?pwd=", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestGetDevices_InvalidPassword(t *testing.T) {
	mux := setupServerWithPwd(t, "testpass")
	req := httptest.NewRequest("GET", "/api/devices?pwd=wrong", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestGetDevices_NoPasswordSetInService(t *testing.T) {
	mux := setupServerWithPwd(t, "") // no password set
	req := httptest.NewRequest("GET", "/api/devices", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandleIsDeviceRegistered_PathParsing(t *testing.T) {
	mockInflux := &influxdb.MockClient{}
	mockPlmntclient := &planetmint.MockPlanetmintClient{}
	mockDB := &database.MockDatabase{}
	// Device exists
	mockDB.On("GetDevice", "dev123").Return(database.Device{LiquidAddress: "Liquid_address", DeviceName: "dev123", DeviceType: "washing machine", PlanetmintAddress: "plmnt...", Timestamp: time.Now()}, true, nil)
	srv, err := server.NewServer(mockPlmntclient, mockInflux, mockDB)
	assert.NoError(t, err)
	mux := http.NewServeMux()
	srv.Routes(mux)
	mux.HandleFunc("/api/device/", srv.HandleIsDeviceRegistered)
	//t.Cleanup(func() { srv.Close() })

	// Valid path: /api/device/dev123
	req := httptest.NewRequest("GET", "/api/device/dev123", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "dev123")

	// Invalid path: /api/device (missing ID)
	req2 := httptest.NewRequest("GET", "/api/device", nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusBadRequest, rr2.Code)
	assert.Contains(t, rr2.Body.String(), "Missing device ID")

	// Too many segments: /api/device/dev123/extra
	req3 := httptest.NewRequest("GET", "/api/device/dev123/extra", nil)
	rr3 := httptest.NewRecorder()
	mux.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusBadRequest, rr3.Code)
	assert.Contains(t, rr3.Body.String(), "Invalid device ID")

	// Device not found
	mockDB.On("GetDevice", "notfound").Return(database.Device{}, false, nil)
	req4 := httptest.NewRequest("GET", "/api/device/notfound", nil)
	rr4 := httptest.NewRecorder()
	mux.ServeHTTP(rr4, req4)
	assert.Equal(t, http.StatusNotFound, rr4.Code)
	assert.Contains(t, rr4.Body.String(), "not found")
}
