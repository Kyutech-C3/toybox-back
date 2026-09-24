package proxy

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const responseHeaderTimeout = 10 * time.Second

var (
	ErrBaseURLNotConfigured = errors.New("legacy toybox base URL is not configured")
	ErrInvalidBaseURL       = errors.New("legacy toybox base URL is invalid")
	ErrInvalidProxyHost     = errors.New("legacy toybox proxy host is invalid")
)

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

// NewLegacyToyBoxProxy builds the HTTP transport used by the temporary legacy
// API bridge. HTTP response formatting remains the responsibility of the
// interface layer through errorHandler.
func NewLegacyToyBoxProxy(upstreamBaseURL, proxyHost string, errorHandler ErrorHandler) (*httputil.ReverseProxy, error) {
	if strings.TrimSpace(upstreamBaseURL) == "" {
		return nil, ErrBaseURLNotConfigured
	}

	targetURL, err := url.Parse(upstreamBaseURL)
	if err != nil || (targetURL.Scheme != "http" && targetURL.Scheme != "https") || targetURL.Host == "" || targetURL.User != nil || targetURL.Fragment != "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidBaseURL, upstreamBaseURL)
	}

	upstreamHost := targetURL.Host
	if proxyHost != "" {
		if !isValidProxyHost(proxyHost) {
			return nil, fmt.Errorf("%w: %q", ErrInvalidProxyHost, proxyHost)
		}
		upstreamHost = proxyHost
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := reverseProxy.Director
	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)

		// The bridged endpoints are public. Never cross the legacy-service
		// boundary with credentials belonging to the current application.
		req.Header.Del("Cookie")
		req.Header.Del("Authorization")
		req.Host = upstreamHost
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = responseHeaderTimeout
	reverseProxy.Transport = transport

	reverseProxy.ModifyResponse = func(response *http.Response) error {
		// CORS is owned by Echo. Forwarding the legacy CORS headers produces
		// duplicate Access-Control-Allow-* values that browsers reject.
		for header := range response.Header {
			if strings.HasPrefix(http.CanonicalHeaderKey(header), "Access-Control-") {
				response.Header.Del(header)
			}
		}

		// The legacy service must not be able to create or overwrite cookies on
		// the current application's origin.
		response.Header.Del("Set-Cookie")
		return nil
	}

	if errorHandler != nil {
		reverseProxy.ErrorHandler = errorHandler
	}

	return reverseProxy, nil
}

func isValidProxyHost(proxyHost string) bool {
	parsedHost, err := url.Parse("http://" + proxyHost)
	return err == nil && parsedHost.User == nil && parsedHost.Host == proxyHost && parsedHost.Hostname() != "" && parsedHost.Path == "" && parsedHost.RawQuery == "" && parsedHost.Fragment == ""
}
