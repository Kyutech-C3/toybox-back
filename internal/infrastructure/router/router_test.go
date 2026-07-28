package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

/*
ルーターに登録した旧ToyBox向けAPIのパスとクエリが、プロキシ先に渡されることを確認。
認証情報のヘッダーがプロキシ先へ送信されないことを確認。
*/
func TestNewLegacyToyBoxProxy_ForwardsPathQueryAndHeaders(t *testing.T) {
	t.Helper()

	var receivedPath string
	var receivedQuery string
	var receivedAuthorization string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedQuery = r.URL.RawQuery
		receivedAuthorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	e := echo.New()
	e.GET("/api/v1/blogs/:blog_id", newLegacyToyBoxProxy(upstream.URL))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/blogs/abc123?page=2", nil)
	req.Header.Set("Authorization", "Bearer sample-token")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusOK)
	}
	if receivedPath != "/api/v1/blogs/abc123" {
		t.Fatalf("unexpected upstream path: got %q want %q", receivedPath, "/api/v1/blogs/abc123")
	}
	if receivedQuery != "page=2" {
		t.Fatalf("unexpected upstream query: got %q want %q", receivedQuery, "page=2")
	}
	if receivedAuthorization != "" {
		t.Fatalf("unexpected upstream authorization header: got %q want %q", receivedAuthorization, "")
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("unexpected response content type: got %q", contentType)
	}
}

/*
旧ToyBoxの接続先のURLが設定されていないときに、routerを経由していても503を返すことを確認
*/
func TestNewLegacyToyBoxProxy_ReturnsServiceUnavailableWhenBaseURLMissing(t *testing.T) {
	t.Helper()

	e := echo.New()
	e.GET("/api/v1/works", newLegacyToyBoxProxy(""))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if body := rec.Body.String(); !strings.Contains(body, "legacy toybox base url is not configured") {
		t.Fatalf("unexpected response body: %q", body)
	}
}
