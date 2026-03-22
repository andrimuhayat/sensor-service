package ratelimit

import (
	"github.com/labstack/echo/v4"
	ratelimitHandler "sensor-service/internal/module/ratelimit/handler"
	"sensor-service/internal/platform/app"
	module "sensor-service/internal/platform/common"
	"time"
)

// RateLimiterConfig holds the rate limiter configuration
type RateLimiterConfig struct {
	MaxRequests int           // Maximum requests per window
	WindowSize  time.Duration // Time window duration
}

// defaultRateLimiterConfig returns the default configuration
func defaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		MaxRequests: 100,        // 100 requests
		WindowSize:  time.Minute, // per minute
	}
}

// StartService initializes and starts the rate limiter service
// Follows the pattern from internal/module/auth/service.go
func StartService(dependency module.Dependency, router *echo.Echo, app app.App) {
	// Use default configuration
	config := defaultRateLimiterConfig()

	// Initialize handler with configuration
	handler := ratelimitHandler.NewHandler(config)

	// Define routes
	versionRoute := router.Group("/api")
	ratelimitHandler.NewRoute(handler, versionRoute)
}
