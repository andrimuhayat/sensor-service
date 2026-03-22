package handler

import (
	"github.com/labstack/echo/v4"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

// Handler represents the rate limit handler
type Handler struct {
}

// NewHandler creates a new rate limit handler
func NewHandler() Handler {
	return Handler{}
}

// HealthCheck is a simple health check endpoint
func (h Handler) HealthCheck(c echo.Context) error {
	return httpresponse.ResponseWithSuccess(c, map[string]string{
		"status": "ok",
	})
}

// NewRateLimitRoute initializes rate limit related routes
func NewRateLimitRoute(h Handler, route *echo.Group) {
	rateLimit := route.Group("/ratelimit")
	rateLimit.GET("/health", h.HealthCheck)
}
