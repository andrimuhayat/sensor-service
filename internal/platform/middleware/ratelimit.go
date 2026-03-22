package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

// RateLimiterConfig holds configuration for rate limiter
type RateLimiterConfig struct {
	MaxRequests int           // Maximum requests allowed per window
	WindowSize time.Duration // Time window for rate limiting
}

// RateLimiterData holds rate limiting data per client
type RateLimiterData struct {
	Count     int       // Current request count in window
	ResetAt   time.Time // When the rate limit resets
	mu       sync.Mutex
}

// RateLimiter is a token bucket rate limiter implementation
// O(1) complexity for check and update operations
type RateLimiter struct {
	clients map[string]*RateLimiterData
	config  RateLimiterConfig
	mu      sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with the given configuration
// Complexity: O(1)
func NewRateLimiter(maxRequests int, windowSize time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*RateLimiterData),
		config: RateLimiterConfig{
			MaxRequests: maxRequests,
			WindowSize:  windowSize,
		},
	}
}

// getClientKey extracts client identifier from request
// Uses IP address as default identifier
func getClientKey(c echo.Context) string {
	// Try to get IP from X-Forwarded-For header first (for proxied requests)
	if ip := c.Request().Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	// Fall back to real IP
	ip := c.RealIP()
	return ip
}

// getOrCreateClient gets or creates rate limiter data for a client
// O(1) complexity with map lookup
func (rl *RateLimiter) getOrCreateClient(key string) *RateLimiterData {
	rl.mu.RLock()
	data, exists := rl.clients[key]
	rl.mu.RUnlock()

	if exists {
		return data
	}

	// Create new client entry
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if data, exists := rl.clients[key]; exists {
		return data
	}

	now := time.Now()
	data = &RateLimiterData{
		Count:   0,
		ResetAt: now.Add(rl.config.WindowSize),
	}
	rl.clients[key] = data
	return data
}

// CheckAndIncrement checks if request is allowed and increments counter
// Returns: allowed (bool), remaining requests (int), reset time (time.Duration)
// Complexity: O(1)
func (rl *RateLimiter) CheckAndIncrement(key string) (bool, int, time.Duration) {
	data := rl.getOrCreateClient(key)
	data.mu.Lock()
	defer data.mu.Unlock()

	now := time.Now()

	// Reset counter if window has expired
	if now.After(data.ResetAt) {
		data.Count = 0
		data.ResetAt = now.Add(rl.config.WindowSize)
	}

	// Check if limit exceeded
	if data.Count >= rl.config.MaxRequests {
		remaining := 0
		resetIn := data.ResetAt.Sub(now)
		return false, remaining, resetIn
	}

	// Increment counter
	data.Count++
	remaining := rl.config.MaxRequests - data.Count
	resetIn := data.ResetAt.Sub(now)

	return true, remaining, resetIn
}

// RateLimitMiddleware creates Echo middleware for rate limiting
// Adds X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset headers
func RateLimitMiddleware(rl *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := getClientKey(c)
			allowed, remaining, resetIn := rl.CheckAndIncrement(key)

			// Set rate limit headers
			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.MaxRequests))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(resetIn).Unix()))

			if !allowed {
				return httpresponse.ResponseWithErrors(c, http.StatusTooManyRequests, &httpresponse.HTTPError{
					Code:    http.StatusTooManyRequests,
					Message: "TOO_MANY_REQUESTS",
				})
			}

			return next(c)
		}
	}
}

// Reset clears all rate limiter data
// Complexity: O(n) where n is number of clients
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.clients = make(map[string]*RateLimiterData)
}

// GetConfig returns the current rate limiter configuration
func (rl *RateLimiter) GetConfig() RateLimiterConfig {
	return rl.config
}
