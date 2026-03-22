package ratelimit

import (
	"sensor-service/internal/platform/app"
	module "sensor-service/internal/platform/common"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"sensor-service/internal/platform/middleware"
)

// Service holds the rate limiter instance
type Service struct {
	ratelimiter *middleware.RateLimiter
}

// NewService creates a new rate limit service
func NewService(maxRequests int, windowSize time.Duration) *Service {
	return &Service{
		ratelimiter: middleware.NewRateLimiter(maxRequests, windowSize),
	}
}

// GetRateLimiter returns the rate limiter instance
func (s *Service) GetRateLimiter() *middleware.RateLimiter {
	return s.ratelimiter
}

// GetMiddleware returns the Echo middleware for rate limiting
func (s *Service) GetMiddleware() echo.MiddlewareFunc {
	return middleware.RateLimitMiddleware(s.ratelimiter)
}

// Reset resets the rate limiter
func (s *Service) Reset() {
	s.ratelimiter.Reset()
}

// GetConfig returns the rate limiter configuration
func (s *Service) GetConfig() middleware.RateLimiterConfig {
	return s.ratelimiter.GetConfig()
}

func RunConsumer(wg *sync.WaitGroup, f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
}

// StartService initializes and wires the rate limiter service
// This is the wiring function mentioned in the QA findings
func StartService(dependency module.Dependency, router *echo.Echo, app app.App) {
	// Initialize rate limiter service with default configuration
	// 100 requests per minute
	rateLimitService := NewService(100, time.Minute)

	// Store in dependency for other modules to use
	dependency.RateLimitService = rateLimitService

	// Apply rate limiter middleware globally to all routes
	router.Use(rateLimitService.GetMiddleware())
}
