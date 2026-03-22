package middleware

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	MaxRequests int           // Maximum requests allowed within the time window
	TimeWindow  time.Duration // Time window for rate limiting
}

// RateLimiterData stores rate limit data for each client
type RateLimiterData struct {
	Count       int
	ResetAt    time.Time
}

// RateLimiterStore is thread-safe storage for rate limit data
type RateLimiterStore struct {
	mu   sync.RWMutex
	data map[string]*RateLimiterData
}

// RateLimiter implements rate limiting middleware
type RateLimiter struct {
	store  *RateLimiterStore
	config RateLimiterConfig
}

// NewRateLimiter creates a new RateLimiter instance
// O(n) space complexity where n is number of unique clients
func NewRateLimiter(maxRequests int, timeWindow time.Duration) *RateLimiter {
	return &RateLimiter{
		store: &RateLimiterStore{
			data: make(map[string]*RateLimiterData),
		},
		config: RateLimiterConfig{
			MaxRequests: maxRequests,
			TimeWindow:  timeWindow,
		},
	}
}

// RateLimitMiddleware returns Echo middleware for rate limiting
// O(1) time complexity for checking and updating rate limit
func (rl *RateLimiter) RateLimitMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get client identifier (IP address)
			clientID := getClientIP(c)
			if clientID == "" {
				return httpresponse.ResponseWithErrors(c, http.StatusBadRequest, &httpresponse.HTTPError{
					Code:    http.StatusBadRequest,
					Message: "Unable to identify client",
				})
			}

			// Check if request is allowed
			allowed, err := rl.checkRateLimit(clientID)
			if err != nil {
				return err
			}

			if !allowed {
				return httpresponse.ResponseWithErrors(c, http.StatusTooManyRequests, &httpresponse.HTTPError{
					Code:    http.StatusTooManyRequests,
					Message: "RATE_LIMIT_EXCEEDED",
				})
			}

			// Add rate limit headers
			c.Response().Header().Set("X-RateLimit-Limit", string(rune(rl.config.MaxRequests))
			c.Response().Header().Set("X-RateLimit-Remaining", string(rune(rl.getRemainingRequests(clientID)))
			c.Response().Header().Set("X-RateLimit-Reset", string(rune(rl.getResetTime(clientID)))

			return next(c)
		}
	}
}

// checkRateLimit verifies if the request is allowed for the given client
// O(1) time complexity
func (rl *RateLimiter) checkRateLimit(clientID string) (bool, error) {
	rl.store.mu.Lock()
	defer rl.store.mu.Unlock()

	now := time.Now()
	data, exists := rl.store.data[clientID]

	if !exists {
		// First request from this client
		rl.store.data[clientID] = &RateLimiterData{
			Count:    1,
			ResetAt:  now.Add(rl.config.TimeWindow),
		}
		return true, nil
	}

	// Check if time window has expired
	if now.After(data.ResetAt) {
		// Reset the counter
		data.Count = 1
		data.ResetAt = now.Add(rl.config.TimeWindow)
		return true, nil
	}

	// Check if under limit
	if data.Count < rl.config.MaxRequests {
		data.Count++
		return true, nil
	}

	// Rate limit exceeded
	return false, nil
}

// getRemainingRequests returns the number of remaining requests allowed
// O(1) time complexity
func (rl *RateLimiter) getRemainingRequests(clientID string) int {
	rl.store.mu.RLock()
	defer rl.store.mu.RUnlock()

	data, exists := rl.store.data[clientID]
	if !exists {
		return rl.config.MaxRequests
	}

	remaining := rl.config.MaxRequests - data.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// getResetTime returns the Unix timestamp when the rate limit resets
// O(1) time complexity
func (rl *RateLimiter) getResetTime(clientID string) int {
	rl.store.mu.RLock()
	defer rl.store.mu.RUnlock()

	data, exists := rl.store.data[clientID]
	if !exists {
		return int(time.Now().Add(rl.config.TimeWindow).Unix())
	}

	return int(data.ResetAt.Unix())
}

// getClientIP extracts the client IP address from the request
// O(1) time complexity
func getClientIP(c echo.Context) string {
	// Check X-Forwarded-For header first (for proxied requests)
	ip := c.Request().Header.Get("X-Forwarded-For")
	if ip != "" {
		// Take the first IP if multiple are present
		return splitAndTrim(ip)[0]
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

	return splitAndTrim(remoteAddr)[0]
}

// splitAndTrim splits a string by comma and trims whitespace
// O(n) time complexity where n is the length of the input string
func splitAndTrim(s string) []string {
	var result []string
	for _, part := range split(s, ",") {
		result = append(result, trim(part))
	}
	return result
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

// Reset clears all rate limit data
// O(n) time complexity where n is number of tracked clients
func (rl *RateLimiter) Reset() {
	rl.store.mu.Lock()
	defer rl.store.mu.Unlock()
	rl.store.data = make(map[string]*RateLimiterData)
}

// GetClientData returns the rate limit data for a specific client
// O(1) time complexity
func (rl *RateLimiter) GetClientData(clientID string) (*RateLimiterData, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}

	rl.store.mu.RLock()
	defer rl.store.mu.RUnlock()

	data, exists := rl.store.data[clientID]
	if !exists {
		return nil, errors.New("client not found")
	}

	return data, nil
}
