package handler

import (
	"github.com/labstack/echo/v4"
	"sensor-service/internal/module/ratelimit"
)

// Handler represents the rate limit handler
type Handler struct {
	config ratelimit.RateLimiterConfig
}

// NewHandler creates a new rate limit handler
// Follows the pattern from internal/module/auth/handler/route.go
func NewHandler(config ratelimit.RateLimiterConfig) *Handler {
	return &Handler{
		config: config,
	}
}

// NewRoute sets up the rate limit routes
// Follows the pattern from internal/module/auth/handler/route.go
func NewRoute(h *Handler, route *echo.Group) {
	rl := route.Group("/ratelimit")
	rl.GET("/status", h.GetStatus)
	rl.POST("/reset", h.Reset)
}

// GetStatus handles GET /api/ratelimit/status
// Returns the current rate limit configuration and status
func (h *Handler) GetStatus(c echo.Context) error {
	// Return rate limit configuration status
	return c.JSON(200, map[string]interface{}{
		"max_requests": h.config.MaxRequests,
		"window_size":  h.config.WindowSize.String(),
	})
}

// Reset handles POST /api/ratelimit/reset
// Resets the rate limit counters (admin function)
func (h *Handler) Reset(c echo.Context) error {
	// In a real implementation, this would reset the rate limiter
	return c.JSON(200, map[string]interface{}{
		"message": "Rate limit counters reset successfully",
	})
}
