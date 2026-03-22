package middleware

import (
	"errors"
	"net/http"
	"sensor-service/internal/platform/httpengine/httpresponse"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	MaxRequests int           // Maximum requests allowed within the time window
	TimeWindow  time.Duration // Time window for rate limiting
}

// RateLimiterData stores rate limit data per client
type RateLimiterData struct {
	Count       int
	ResetAt     time.Time
	FirstAt     time.Time
}

// RateLimiter is a thread-safe rate limiter implementation
// Uses in-memory storage with O(1) lookup complexity
type RateLimiter struct {
	mu      sync.RWMutex
	data    map[string]*RateLimiterData
	config  RateLimiterConfig
}

// NewRateLimiter creates a new rate limiter with the given configuration
// O(1) space complexity
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if config.MaxRequests <= 0 {
		config.MaxRequests = 100 // default
	}
	if config.TimeWindow <= 0 {
		config.TimeWindow = time.Minute // default 1 minute
	}
	return &RateLimiter{
		data:   make(map[string]*RateLimiterData),
		config: config,
	}
}

// RateLimitMiddleware returns an Echo middleware for rate limiting
// O(1) time complexity for checking and updating rate limit
func RateLimitMiddleware(rl *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get client identifier (IP address)
			clientID := getClientIP(c)
			if clientID == "" {
				clientID = c.RealIP()
			}

			// Check rate limit - O(1) operation
			allowed, err := rl.Check(clientID)
			if err != nil {
				return httpresponse.ResponseWithErrors(c, http.StatusTooManyRequests, &httpresponse.HTTPError{
					Code:    http.StatusTooManyRequests,
					Message: "RATE_LIMIT_EXCEEDED",
				})
			}

			if !allowed {
				return httpresponse.ResponseWithErrors(c, http.StatusTooManyRequests, &httpresponse.HTTPError{
					Code:    http.StatusTooManyRequests,
					Message: "RATE_LIMIT_EXCEEDED",
				})
			}

			// Add rate limit headers to response
			c.Response().Header().Set("X-RateLimit-Limit", string(rune(rl.config.MaxRequests)))
			c.Response().Header().Set("X-RateLimit-Remaining", string(rune(rl.Remaining(clientID))))
			c.Response().Header().Set("X-RateLimit-Reset", string(rune(rl.ResetTime(clientID).Unix())))

			return next(c)
		}
	}
}

// Check verifies if the request is allowed under rate limit
// Returns true if allowed, false if rate limit exceeded
// O(1) time complexity
func (rl *RateLimiter) Check(clientID string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	data, exists := rl.data[clientID]

	if !exists {
		// First request from this client - O(1) insertion
		 rl.data[clientID] = &RateLimiterData{
			Count:   1,
			ResetAt: now.Add(rl.config.TimeWindow),
			FirstAt: now,
		}
		return true, nil
	}

	// Check if time window has expired - reset counter
	if now.After(data.ResetAt) {
		data.Count = 1
		data.ResetAt = now.Add(rl.config.TimeWindow)
		data.FirstAt = now
		return true, nil
	}

	// Check if under limit
	if data.Count < rl.config.MaxRequests {
		data.Count++
		return true, nil
	}

	// Rate limit exceeded
	return false, errors.New("rate limit exceeded")
}

// Remaining returns the number of remaining requests for a client
// O(1) time complexity
func (rl *RateLimiter) Remaining(clientID string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	data, exists := rl.data[clientID]
	if !exists {
		return rl.config.MaxRequests
	}

	if time.Now().After(data.ResetAt) {
		return rl.config.MaxRequests
	}

	remaining := rl.config.MaxRequests - data.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ResetTime returns the time when the rate limit will reset
// O(1) time complexity
func (rl *RateLimiter) ResetTime(clientID string) time.Time {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	data, exists := rl.data[clientID]
	if !exists {
		return time.Now().Add(rl.config.TimeWindow)
	}

	return data.ResetAt
}

// Reset clears the rate limit data for a client
// O(1) time complexity
func (rl *RateLimiter) Reset(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.data, clientID)
}

// ResetAll clears all rate limit data
// O(n) time complexity where n is number of tracked clients
func (rl *RateLimiter) ResetAll() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.data = make(map[string]*RateLimiterData)
}

// getClientIP extracts the client IP from the request
// Checks X-Forwarded-For header first, then falls back to RealIP
func getClientIP(c echo.Context) string {
	// Check X-Forwarded-For header (for reverse proxy setups)
	forwarded := c.Request().Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := c.Request().Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	return ""
}
