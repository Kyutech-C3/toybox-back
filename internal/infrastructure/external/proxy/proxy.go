package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo/v4"
)

func NewLegacyToyBoxProxy(upstreamBaseURL string, proxyHost string) echo.HandlerFunc {
	if upstreamBaseURL == "" {
		return func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "legacy toybox base url is not configured")
		}
	}

	targetURL, err := url.Parse(upstreamBaseURL)
	if err != nil {
		return func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "legacy toybox base url is invalid")
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// 認証不要なためCookieとAuthorizationヘッダーを削除
		req.Header.Del("Cookie")
		req.Header.Del("Authorization")

		if proxyHost != "" {
			req.Host = proxyHost
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, proxyErr error) {
		w.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"message":"legacy toybox proxy error"}`))
	}

	return func(c echo.Context) error {
		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
