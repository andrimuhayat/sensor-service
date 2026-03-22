package middleware

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	MaxRequests int           // Maximum requests allowed per window
	WindowSize  time.Duration // Time window for rate limiting
}

// rateLimitEntry tracks rate limit data for a client
type rateLimitEntry struct {
	Count     int
	StartTime time.Time
}

// RateLimiter is a middleware for rate limiting requests
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*rateLimitEntry
	config   RateLimiterConfig
}

// NewRateLimiter creates a new RateLimiter instance
// O(1) complexity for creating a new rate limiter
func NewRateLimiter(maxRequests int, windowSize time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*rateLimitEntry),
		config: RateLimiterConfig{
			MaxRequests: maxRequests,
			WindowSize:  windowSize,
		},
	}
}

// RateLimit returns an Echo middleware function for rate limiting
// O(1) complexity for checking rate limit per request
func (rl *RateLimiter) RateLimit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientID := getClientID(c)

			// Check if rate limit is exceeded
			allowed, err := rl.checkLimit(clientID)
			if err != nil {
				return httpresponse.ResponseWithErrors(c, http.StatusTooManyRequests, &httpresponse.HTTPError{
					Code:    http.StatusTooManyRequests,
					Message: "RATE_LIMIT_EXCEEDED",
				})
			}

			if !allowed {
				retryAfter := rl.getRetryAfter(clientID)
				c.Response().Header().Set("Retry-After", retryAfter)
				return httpresponse.ResponseWithErrors(c, http.StatusTooManyRequests, &httpresponse.HTTPError{
					Code:    http.StatusTooManyRequests,
					Message: "RATE_LIMIT_EXCEEDED",
				})
			}

			return next(c)
		}
	}
}

// checkLimit checks if the client is within rate limits
// O(1) complexity for lookup and update
func (rl *RateLimiter) checkLimit(clientID string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.clients[clientID]
	now := time.Now()

	if !exists {
		// First request from this client
		rl.clients[clientID] = &rateLimitEntry{
			Count:     1,
			StartTime: now,
		}
		return true, nil
	}

	// Check if time window has expired
	elapsed := now.Sub(entry.StartTime)
	if elapsed >= rl.config.WindowSize {
		// Reset counter for new window
		entry.Count = 1
		entry.StartTime = now
		return true, nil
	}

	// Check if request count is within limit
	if entry.Count >= rl.config.MaxRequests {
		return false, errors.New("rate limit exceeded")
	}

	// Increment request count
	entry.Count++
	return true, nil
}

// getRetryAfter returns seconds until the rate limit resets
// O(1) complexity
func (rl *RateLimiter) getRetryAfter(clientID string) string {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	entry, exists := rl.clients[clientID]
	if !exists {
		return "0"
	}

	remainingTime := entry.StartTime.Add(rl.config.WindowSize).Sub(time.Now())
	if remainingTime <= 0 {
		return "0"
	}

	return formatSeconds(remainingTime)
}

// getClientID extracts client identifier from the request
// Supports: X-Forwarded-For, X-Real-IP, or remote addr
// O(1) complexity
func getClientID(c echo.Context) string {
	// Check X-Forwarded-For header first (for reverse proxy)
	xff := c.Request().Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the chain
		parts := slices.DeleteFunc(slices.Clone(slices.ValuesFunc(xff, func(r rune) bool { return r == ',' || r == ' ' })), func(s string) bool { return s == "" })
		if len(parts) > 0 {
			return parts[0]
		}
	}

	// Check X-Real-IP header
	xri := c.Request().Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to remote address
	return c.RealIP()
}

// formatSeconds formats duration as seconds string
// O(1) complexity
func formatSeconds(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 1 {
		return "1"
	}
	return string(rune('0' + seconds/10%10))
}

// Reset clears all rate limit data
// O(n) complexity where n is number of tracked clients
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.clients = make(map[string]*rateLimitEntry)
}

// GetClientCount returns the current request count for a client
// O(1) complexity
func (rl *RateLimiter) GetClientCount(clientID string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	entry, exists := rl.clients[clientID]
	if !exists {
		return 0
	}

	// Check if window has expired
	if time.Since(entry.StartTime) >= rl.config.WindowSize {
		return 0
	}

	return entry.Count
}

// IsClientLimited checks if a specific client is currently rate limited
// O(1) complexity
func (rl *RateLimiter) IsClientLimited(clientID string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	entry, exists := rl.clients[clientID]
	if !exists {
		return false
	}

	// Check if window has expired
	if time.Since(entry.StartTime) >= rl.config.WindowSize {
		return false
	}

	return entry.Count >= rl.config.MaxRequests
}
