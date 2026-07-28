package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewLegacyToyBoxProxy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Mock upstream server
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message": "pong"}`))
		}))
		defer upstream.Close()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/works", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewLegacyToyBoxProxy(upstream.URL, "")
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"message": "pong"}`, rec.Body.String())
	})

	t.Run("empty base url", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/works", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewLegacyToyBoxProxy("", "")
		err := handler(c)

		assert.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusServiceUnavailable, he.Code)
		assert.Equal(t, "legacy toybox base url is not configured", he.Message)
	})

	t.Run("invalid base url", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/works", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// url.Parse is quite lenient, but ":" is invalid as a URL
		handler := NewLegacyToyBoxProxy(":", "")
		err := handler(c)

		assert.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusServiceUnavailable, he.Code)
		assert.Equal(t, "legacy toybox base url is invalid", he.Message)
	})

	t.Run("upstream server error", func(t *testing.T) {
		// We want to test the proxy error handler.
		// One way to trigger it is to shut down the server before calling.
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		url := upstream.URL
		upstream.Close() // Close immediately to cause a connection error

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/works", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewLegacyToyBoxProxy(url, "")
		err := handler(c)

		assert.NoError(t, err) // echo.HandlerFunc returns nil, error is handled by proxy.ErrorHandler
		assert.Equal(t, http.StatusBadGateway, rec.Code)
		assert.JSONEq(t, `{"message":"legacy toybox proxy error"}`, rec.Body.String())
	})

	t.Run("with proxy host overwrite", func(t *testing.T) {
		proxyHost := "overwritten-host.com"
		// Mock upstream server to verify Host header
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, proxyHost, r.Host)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message": "pong"}`))
		}))
		defer upstream.Close()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/works", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewLegacyToyBoxProxy(upstream.URL, proxyHost)
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"message": "pong"}`, rec.Body.String())
	})

	t.Run("remove cookies and authorization headers", func(t *testing.T) {
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, r.Header.Get("Cookie"))
			assert.Empty(t, r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
		}))
		defer upstream.Close()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/works", nil)
		req.Header.Set("Cookie", "session=123")
		req.Header.Set("Authorization", "Bearer token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewLegacyToyBoxProxy(upstream.URL, "")
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
