package proxy

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewLegacyToyBoxProxy(t *testing.T) {
	var receivedPath string
	var receivedQuery string
	var receivedHost string
	var receivedCookie string
	var receivedAuthorization string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedQuery = r.URL.RawQuery
		receivedHost = r.Host
		receivedCookie = r.Header.Get("Cookie")
		receivedAuthorization = r.Header.Get("Authorization")

		w.Header().Set("Access-Control-Allow-Origin", "https://frontend.example")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Set-Cookie", "refresh_token=legacy; Path=/")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))
	defer upstream.Close()

	reverseProxy, err := NewLegacyToyBoxProxy(upstream.URL, "", nil)
	if err != nil {
		t.Fatalf("NewLegacyToyBoxProxy() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://current.example/api/v1/works?page=2", nil)
	req.Header.Set("Cookie", "refresh_token=current")
	req.Header.Set("Authorization", "Bearer current-token")
	rec := httptest.NewRecorder()
	reverseProxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if receivedPath != "/api/v1/works" || receivedQuery != "page=2" {
		t.Fatalf("upstream target = %q?%s, want /api/v1/works?page=2", receivedPath, receivedQuery)
	}
	wantHost := strings.TrimPrefix(upstream.URL, "http://")
	if receivedHost != wantHost {
		t.Errorf("upstream Host = %q, want %q", receivedHost, wantHost)
	}
	if receivedCookie != "" || receivedAuthorization != "" {
		t.Errorf("credentials crossed proxy boundary: Cookie=%q Authorization=%q", receivedCookie, receivedAuthorization)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want it removed", got)
	}
	if cookies := rec.Header().Values("Set-Cookie"); len(cookies) != 0 {
		t.Errorf("Set-Cookie = %#v, want none", cookies)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
}

func TestNewLegacyToyBoxProxyOverridesHost(t *testing.T) {
	const proxyHost = "legacy.example:8443"
	var receivedHost string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHost = r.Host
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	reverseProxy, err := NewLegacyToyBoxProxy(upstream.URL, proxyHost, nil)
	if err != nil {
		t.Fatalf("NewLegacyToyBoxProxy() error = %v", err)
	}
	reverseProxy.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://current.example/api/v1/works", nil))

	if receivedHost != proxyHost {
		t.Errorf("upstream Host = %q, want %q", receivedHost, proxyHost)
	}
}

func TestNewLegacyToyBoxProxyValidatesConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		proxyHost string
		wantErr   error
	}{
		{name: "missing base URL", wantErr: ErrBaseURLNotConfigured},
		{name: "scheme-less base URL", baseURL: "legacy.example", wantErr: ErrInvalidBaseURL},
		{name: "unsupported scheme", baseURL: "ftp://legacy.example", wantErr: ErrInvalidBaseURL},
		{name: "missing host", baseURL: "https:///api", wantErr: ErrInvalidBaseURL},
		{name: "credentials in URL", baseURL: "https://user:password@legacy.example", wantErr: ErrInvalidBaseURL},
		{name: "fragment in URL", baseURL: "https://legacy.example/#fragment", wantErr: ErrInvalidBaseURL},
		{name: "invalid proxy host", baseURL: "https://legacy.example", proxyHost: "other.example/path", wantErr: ErrInvalidProxyHost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLegacyToyBoxProxy(tt.baseURL, tt.proxyHost, nil)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want errors.Is(_, %v)", err, tt.wantErr)
			}
		})
	}
}

func TestNewLegacyToyBoxProxyConfiguresResponseHeaderTimeout(t *testing.T) {
	reverseProxy, err := NewLegacyToyBoxProxy("https://legacy.example", "", nil)
	if err != nil {
		t.Fatalf("NewLegacyToyBoxProxy() error = %v", err)
	}

	transport, ok := reverseProxy.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport type = %T, want *http.Transport", reverseProxy.Transport)
	}
	if transport.ResponseHeaderTimeout != responseHeaderTimeout {
		t.Errorf("ResponseHeaderTimeout = %v, want %v", transport.ResponseHeaderTimeout, responseHeaderTimeout)
	}
}
