package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	MaxRequests int           // Maximum requests allowed within the time window
	WindowSize  time.Duration // Time window for rate limiting
}

// rateLimitEntry stores rate limit data for each client
type rateLimitEntry struct {
	Count     int
	StartTime time.Time
}

// RateLimiter implements rate limiting middleware
type RateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*rateLimitEntry
	config  RateLimiterConfig
}

// NewRateLimiter creates a new RateLimiter instance
// O(n) space complexity where n is number of unique clients
func NewRateLimiter(maxRequests int, windowSize time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*rateLimitEntry),
		config: RateLimiterConfig{
			MaxRequests: maxRequests,
			WindowSize:  windowSize,
		},
	}
}

// RateLimit returns Echo middleware for rate limiting
// O(1) time complexity for checking and updating rate limit
func (rl *RateLimiter) RateLimit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get client identifier (IP address)
			clientID := getClientID(c)
			if clientID == "" {
				return httpresponse.ResponseWithErrors(c, http.StatusBadRequest, &httpresponse.HTTPError{
					Code:    http.StatusBadRequest,
					Message: "Unable to identify client",
				})
			}

			// Check if request is allowed
			allowed := rl.checkRateLimit(clientID)
			if !allowed {
				return httpresponse.ResponseWithErrors(c, http.StatusTooManyRequests, &httpresponse.HTTPError{
					Code:    http.StatusTooManyRequests,
					Message: "RATE_LIMIT_EXCEEDED",
				})
			}

			// Add rate limit headers
			remaining := rl.config.MaxRequests - rl.GetClientCount(clientID)
			if remaining < 0 {
				remaining = 0
			}
			c.Response().Header().Set("X-RateLimit-Limit", string(rune(rl.config.MaxRequests)))
			c.Response().Header().Set("X-RateLimit-Remaining", string(rune(remaining)))
			c.Response().Header().Set("X-RateLimit-Reset", string(rune(rl.getResetTime(clientID))))

			return next(c)
		}
	}
}

// checkRateLimit verifies if the request is allowed for the given client
// O(1) time complexity
func (rl *RateLimiter) checkRateLimit(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.clients[clientID]

	if !exists {
		// First request from this client
		rl.clients[clientID] = &rateLimitEntry{
			Count:     1,
			StartTime: now,
		}
		return true
	}

	// Check if time window has expired
	if now.Sub(entry.StartTime) > rl.config.WindowSize {
		// Reset the counter
		entry.Count = 1
		entry.StartTime = now
		return true
	}

	// Check if under limit
	if entry.Count < rl.config.MaxRequests {
		entry.Count++
		return true
	}

	// Rate limit exceeded
	return false
}

// GetClientCount returns the number of requests made by the client
// O(1) time complexity
func (rl *RateLimiter) GetClientCount(clientID string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	entry, exists := rl.clients[clientID]
	if !exists {
		return 0
	}

	// Check if time window has expired
	if time.Now().Sub(entry.StartTime) > rl.config.WindowSize {
		return 0
	}

	return entry.Count
}

// IsClientLimited checks if the client has exceeded their rate limit
// O(1) time complexity
func (rl *RateLimiter) IsClientLimited(clientID string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	entry, exists := rl.clients[clientID]
	if !exists {
		return false
	}

	// Check if time window has expired
	if time.Now().Sub(entry.StartTime) > rl.config.WindowSize {
		return false
	}

	return entry.Count >= rl.config.MaxRequests
}

// getResetTime returns the Unix timestamp when the rate limit resets
// O(1) time complexity
func (rl *RateLimiter) getResetTime(clientID string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	entry, exists := rl.clients[clientID]
	if !exists {
		return int(time.Now().Add(rl.config.WindowSize).Unix())
	}

	resetTime := entry.StartTime.Add(rl.config.WindowSize)
	return int(resetTime.Unix())
}

// getClientID extracts the client IP address from the request
// O(1) time complexity
func getClientID(c echo.Context) string {
	// Check X-Forwarded-For header first (for proxied requests)
	ip := c.Request().Header.Get("X-Forwarded-For")
	if ip != "" {
		// Take the first IP if multiple are present
		for _, part := range split(ip, ",") {
			trimmed := trim(part)
			if trimmed != "" {
				return trimmed
			}
		}
	}

	// Check X-Real-IP header
	ip = c.Request().Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Fall back to remote address
	remoteAddr := c.Request().RemoteAddr
	if remoteAddr == "" {
		return ""
	}

	// Extract IP from remote address (format: "IP:port")
	for i := len(remoteAddr) - 1; i >= 0; i-- {
		if remoteAddr[i] == ':' {
			return remoteAddr[:i]
		}
	}
	return remoteAddr
}

// split is a simple string split function
func split(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start - 1
		}
	}
	result = append(result, s[start:])
	return result
}

// trim removes leading and trailing whitespace
func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// Reset clears all client data
// O(n) time complexity where n is number of tracked clients
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.clients = make(map[string]*rateLimitEntry)
}
