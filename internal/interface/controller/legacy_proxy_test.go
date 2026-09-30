package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestLegacyProxyControllerReturnsServiceUnavailableForSetupError(t *testing.T) {
	e := echo.New()
	controller := NewLegacyProxyController(nil, errors.New("invalid legacy proxy configuration"))
	e.GET("/api/v1/works", controller.GetWorksV1)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/works", nil))

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(t, `{"message":"legacy toybox proxy is unavailable"}`, rec.Body.String())
}

func TestWriteLegacyProxyError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteLegacyProxyError(rec, httptest.NewRequest(http.MethodGet, "/api/v1/works", nil), errors.New("upstream unavailable"))

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.True(t, strings.Contains(rec.Header().Get(echo.HeaderContentType), echo.MIMEApplicationJSON))
	assert.JSONEq(t, `{"message":"legacy toybox proxy error"}`, rec.Body.String())
}
