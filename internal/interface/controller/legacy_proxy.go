package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
)

const legacyProxyUnavailableMessage = "legacy toybox proxy is unavailable"

type LegacyProxyController struct {
	proxyHandler http.Handler
	setupErr     error
}

func NewLegacyProxyController(proxyHandler http.Handler, setupErr error) *LegacyProxyController {
	return &LegacyProxyController{
		proxyHandler: proxyHandler,
		setupErr:     setupErr,
	}
}

// GetWorksV1 godoc
// @Summary Get works from the legacy API (v1)
// @Description Temporary read-only pass-through to the legacy ToyBox API. The response schema is owned by the legacy service.
// @Tags legacy-proxy
// @Produce json
// @Param limit query int false "Maximum number of works"
// @Param visibility query string false "Visibility filter"
// @Param oldest_work_id query string false "Oldest work cursor"
// @Param newest_work_id query string false "Newest work cursor"
// @Param tag_names query string false "Comma-separated tag names"
// @Param tag_ids query string false "Comma-separated tag IDs"
// @Param search_word query string false "Search text"
// @Success 200 {object} map[string]interface{}
// @Failure 502 {object} schema.LegacyProxyErrorResponse
// @Failure 503 {object} schema.LegacyProxyErrorResponse
// @Router /api/v1/works [get]
func (lc *LegacyProxyController) GetWorksV1(c echo.Context) error {
	return lc.serve(c)
}

// GetWorksV2 godoc
// @Summary Get works from the legacy API (v2)
// @Description Temporary read-only pass-through to the legacy ToyBox API. The response schema is owned by the legacy service.
// @Tags legacy-proxy
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Maximum number of works"
// @Param visibility query string false "Visibility filter"
// @Param tag_names query string false "Comma-separated tag names"
// @Param tag_ids query string false "Comma-separated tag IDs"
// @Param search_word query string false "Search text"
// @Success 200 {object} map[string]interface{}
// @Failure 502 {object} schema.LegacyProxyErrorResponse
// @Failure 503 {object} schema.LegacyProxyErrorResponse
// @Router /api/v2/works [get]
func (lc *LegacyProxyController) GetWorksV2(c echo.Context) error {
	return lc.serve(c)
}

// GetBlogs godoc
// @Summary Get blogs from the legacy API
// @Description Temporary read-only pass-through to the legacy ToyBox API. The response schema is owned by the legacy service.
// @Tags legacy-proxy
// @Produce json
// @Param visibility query string false "Visibility filter"
// @Param limit query int false "Maximum number of blogs"
// @Param page query int false "Page number"
// @Param disable_pagination query bool false "Disable pagination"
// @Success 200 {object} map[string]interface{}
// @Failure 502 {object} schema.LegacyProxyErrorResponse
// @Failure 503 {object} schema.LegacyProxyErrorResponse
// @Router /api/v1/blogs [get]
func (lc *LegacyProxyController) GetBlogs(c echo.Context) error {
	return lc.serve(c)
}

// GetBlog godoc
// @Summary Get a blog from the legacy API
// @Description Temporary read-only pass-through to the legacy ToyBox API. The response schema is owned by the legacy service.
// @Tags legacy-proxy
// @Produce json
// @Param blog_id path string true "Blog ID"
// @Success 200 {object} map[string]interface{}
// @Failure 502 {object} schema.LegacyProxyErrorResponse
// @Failure 503 {object} schema.LegacyProxyErrorResponse
// @Router /api/v1/blogs/{blog_id} [get]
func (lc *LegacyProxyController) GetBlog(c echo.Context) error {
	return lc.serve(c)
}

func (lc *LegacyProxyController) serve(c echo.Context) error {
	if lc.setupErr != nil || lc.proxyHandler == nil {
		if lc.setupErr != nil {
			c.Logger().Error(lc.setupErr)
		}
		return c.JSON(http.StatusServiceUnavailable, schema.LegacyProxyErrorResponse{Message: legacyProxyUnavailableMessage})
	}

	lc.proxyHandler.ServeHTTP(c.Response(), c.Request())
	return nil
}

// WriteLegacyProxyError is passed to the infrastructure proxy so transport
// errors are converted to the public HTTP error contract in the interface layer.
func WriteLegacyProxyError(w http.ResponseWriter, _ *http.Request, proxyErr error) {
	slog.Error("legacy ToyBox proxy request failed", "error", proxyErr)
	w.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	w.WriteHeader(http.StatusBadGateway)
	_ = json.NewEncoder(w).Encode(schema.LegacyProxyErrorResponse{Message: "legacy toybox proxy error"})
}
