package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rddl-network/energy-service/internal/config"
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
	srv, err := server.NewServer(mockPlmntclient, mockInflux)
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
	req := httptest.NewRequest("GET", "/api/devices?pwd=anything", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
