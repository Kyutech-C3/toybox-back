package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
)

func TestLegacyProxyUsesCurrentApplicationCORSOnly(t *testing.T) {
	const frontendOrigin = "https://frontend.example"
	var upstreamHost string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamHost = r.Host
		w.Header().Set(echo.HeaderAccessControlAllowOrigin, frontendOrigin)
		w.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
		w.Header().Add(echo.HeaderSetCookie, "refresh_token=legacy; Path=/")
		w.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	restore := setLegacyProxyConfigForTest(upstream.URL, "", []string{frontendOrigin})
	defer restore()

	e := NewRouter(echo.New(), nil, nil, nil, nil, nil, nil, nil).Setup()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/blogs/blog-id?page=2", nil)
	req.Host = "current.example"
	req.Header.Set(echo.HeaderOrigin, frontendOrigin)
	req.Header.Set(echo.HeaderAuthorization, "Bearer current-token")
	req.Header.Set(echo.HeaderCookie, "refresh_token=current")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Result().Header.Values(echo.HeaderAccessControlAllowOrigin); len(got) != 1 || got[0] != frontendOrigin {
		t.Errorf("Access-Control-Allow-Origin = %#v, want [%q]", got, frontendOrigin)
	}
	if got := rec.Result().Header.Values(echo.HeaderSetCookie); len(got) != 0 {
		t.Errorf("Set-Cookie = %#v, want none", got)
	}
	if upstreamHost != strings.TrimPrefix(upstream.URL, "http://") {
		t.Errorf("upstream Host = %q, want target host", upstreamHost)
	}
}

func TestLegacyProxyReturnsServiceUnavailableWhenNotConfigured(t *testing.T) {
	restore := setLegacyProxyConfigForTest("", "", []string{"http://localhost:3000"})
	defer restore()

	e := NewRouter(echo.New(), nil, nil, nil, nil, nil, nil, nil).Setup()
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/works", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(rec.Body.String(), "legacy toybox proxy is unavailable") {
		t.Errorf("body = %q, want unavailable message", rec.Body.String())
	}
}

func setLegacyProxyConfigForTest(baseURL, proxyHost string, frontendURL []string) func() {
	oldBaseURL := config.LEGACY_TOYBOX_BASE_URL
	oldProxyHost := config.LEGACY_TOYBOX_PROXY_HOST
	oldFrontendURL := config.FRONTEND_URL
	config.LEGACY_TOYBOX_BASE_URL = baseURL
	config.LEGACY_TOYBOX_PROXY_HOST = proxyHost
	config.FRONTEND_URL = frontendURL

	return func() {
		config.LEGACY_TOYBOX_BASE_URL = oldBaseURL
		config.LEGACY_TOYBOX_PROXY_HOST = oldProxyHost
		config.FRONTEND_URL = oldFrontendURL
	}
}
