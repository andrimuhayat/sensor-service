package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ============================================================================
// Helper Functions
// ============================================================================

// createTestRateLimiter creates a rate limiter for testing
// Follows pattern from createTestAppConfig in auth tests
func createTestRateLimiter(maxRequests int, windowSize time.Duration) *RateLimiter {
	return NewRateLimiter(maxRequests, windowSize)
}

// createEchoContext creates a test Echo context
// Helper for testing middleware
func createEchoContext(clientID string) echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if clientID != "" {
		req.Header.Set("X-Real-IP", clientID)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c
}

// ============================================================================
// NewRateLimiter Tests
// ============================================================================

// TestNewRateLimiter_ShouldCreateInstanceWithCorrectConfig tests successful creation
func TestNewRateLimiter_ShouldCreateInstanceWithCorrectConfig(t *testing.T) {
	// Arrange
	maxRequests := 50
	windowSize := 2 * time.Minute

	// Act
	rl := NewRateLimiter(maxRequests, windowSize)

	// Assert
	if rl == nil {
		t.Error("Expected rate limiter to be created")
	}
	if rl.config.MaxRequests != maxRequests {
		t.Errorf("Expected MaxRequests %d, got %d", maxRequests, rl.config.MaxRequests)
	}
	if rl.config.WindowSize != windowSize {
		t.Errorf("Expected WindowSize %v, got %v", windowSize, rl.config.WindowSize)
	}
}

// TestNewRateLimiter_ShouldInitializeEmptyClientMap tests empty client map initialization
func TestNewRateLimiter_ShouldInitializeEmptyClientMap(t *testing.T) {
	// Arrange & Act
	rl := NewRateLimiter(10, time.Minute)

	// Assert
	if rl.clients == nil {
		t.Error("Expected clients map to be initialized")
	}
	if len(rl.clients) != 0 {
		t.Errorf("Expected empty clients map, got %d entries", len(rl.clients))
	}
}

// TestNewRateLimiter_ShouldUseDefaultValues tests default configuration
func TestNewRateLimiter_ShouldUseDefaultValues(t *testing.T) {
	// Arrange & Act
	rl := NewRateLimiter(0, 0)

	// Assert - should handle zero values gracefully
	if rl.config.MaxRequests != 0 {
		t.Errorf("Expected MaxRequests 0, got %d", rl.config.MaxRequests)
	}
}

// ============================================================================
// RateLimit Middleware Tests
// ============================================================================

// TestRateLimit_ShouldAllowRequestWithinLimit tests happy path - request allowed
func TestRateLimit_ShouldAllowRequestWithinLimit(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Minute)
	middleware := rl.RateLimit()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("X-Real-IP", "test-client-1")

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	// Act
	err := middleware(next)(c)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !nextCalled {
		t.Error("Expected next handler to be called")
	}
}

// TestRateLimit_ShouldBlockRequestWhenLimitExceeded tests error case - rate limit exceeded
func TestRateLimit_ShouldBlockRequestWhenLimitExceeded(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(2, time.Minute)
	middleware := rl.RateLimit()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("X-Real-IP", "test-client-2")

	next := func(c echo.Context) error {
		return nil
	}

	// Act - Make 3 requests (2 allowed, 3rd should be blocked)
	var lastErr error
	for i := 0; i < 3; i++ {
		lastErr = middleware(next)(c)
	}

	// Assert
	if lastErr == nil {
		t.Error("Expected error when rate limit exceeded")
	}
}

// TestRateLimit_ShouldReturn429StatusCode tests correct HTTP status code
func TestRateLimit_ShouldReturn429StatusCode(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(1, time.Minute)
	middleware := rl.RateLimit()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("X-Real-IP", "test-client-3")

	next := func(c echo.Context) error {
		return nil
	}

	// Act - Make 2 requests
	middleware(next)(c)
	middleware(next)(c)

	// Assert
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
}

// ============================================================================
// Reset Tests
// ============================================================================

// TestReset_ShouldClearAllClientData tests happy path - reset clears data
func TestReset_ShouldClearAllClientData(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Minute)

	// Add some client data
	rl.mu.Lock()
	rl.clients["client1"] = &rateLimitEntry{Count: 5, StartTime: time.Now()}
	rl.clients["client2"] = &rateLimitEntry{Count: 3, StartTime: time.Now()}
	rl.mu.Unlock()

	// Act
	rl.Reset()

	// Assert
	if len(rl.clients) != 0 {
		t.Errorf("Expected 0 clients after reset, got %d", len(rl.clients))
	}
}

// TestReset_ShouldAllowNewRequestsAfterReset tests edge case - new requests allowed after reset
func TestReset_ShouldAllowNewRequestsAfterReset(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(1, time.Minute)
	middleware := rl.RateLimit()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("X-Real-IP", "test-client-reset")

	next := func(c echo.Context) error {
		return nil
	}

	// Use up the limit
	middleware(next)(c)
	middleware(next)(c) // Should be blocked

	// Act - Reset and try again
	rl.Reset()
	err := middleware(next)(c)

	// Assert
	if err != nil {
		t.Errorf("Expected no error after reset, got %v", err)
	}
}

// ============================================================================
// GetClientCount Tests
// ============================================================================

// TestGetClientCount_ShouldReturnZeroForUnknownClient tests edge case - unknown client
func TestGetClientCount_ShouldReturnZeroForUnknownClient(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Minute)

	// Act
	count := rl.GetClientCount("unknown-client")

	// Assert
	if count != 0 {
		t.Errorf("Expected 0 for unknown client, got %d", count)
	}
}

// TestGetClientCount_ShouldReturnCorrectCount tests happy path - correct count
func TestGetClientCount_ShouldReturnCorrectCount(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Minute)
	middleware := rl.RateLimit()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("X-Real-IP", "test-client-count")

	next := func(c echo.Context) error {
		return nil
	}

	// Make 3 requests
	for i := 0; i < 3; i++ {
		middleware(next)(c)
	}

	// Act
	count := rl.GetClientCount("test-client-count")

	// Assert
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

// TestGetClientCount_ShouldReturnZeroAfterWindowExpiry tests edge case - window expired
func TestGetClientCount_ShouldReturnZeroAfterWindowExpiry(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Millisecond)
	middleware := rl.RateLimit()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("X-Real-IP", "test-client-expiry")

	next := func(c echo.Context) error {
		return nil
	}

	// Make a request
	middleware(next)(c)

	// Wait for window to expire
	time.Sleep(10 * time.Millisecond)

	// Act
	count := rl.GetClientCount("test-client-expiry")

	// Assert
	if count != 0 {
		t.Errorf("Expected 0 after window expiry, got %d", count)
	}
}

// ============================================================================
// IsClientLimited Tests
// ============================================================================

// TestIsClientLimited_ShouldReturnFalseForUnknownClient tests edge case - unknown client
func TestIsClientLimited_ShouldReturnFalseForUnknownClient(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Minute)

	// Act
	limited := rl.IsClientLimited("unknown-client")

	// Assert
	if limited {
		t.Error("Expected unknown client to not be limited")
	}
}

// TestIsClientLimited_ShouldReturnFalseWhenUnderLimit tests happy path - under limit
func TestIsClientLimited_ShouldReturnFalseWhenUnderLimit(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Minute)
	middleware := rl.RateLimit()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("X-Real-IP", "test-client-limited")

	next := func(c echo.Context) error {
		return nil
	}

	// Make 3 requests (under limit of 5)
	for i := 0; i < 3; i++ {
		middleware(next)(c)
	}

	// Act
	limited := rl.IsClientLimited("test-client-limited")

	// Assert
	if limited {
		t.Error("Expected client to not be limited when under limit")
	}
}

// TestIsClientLimited_ShouldReturnTrueWhenAtLimit tests edge case - at limit
func TestIsClientLimited_ShouldReturnTrueWhenAtLimit(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(2, time.Minute)
	middleware := rl.RateLimit()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("X-Real-IP", "test-client-at-limit")

	next := func(c echo.Context) error {
		return nil
	}

	// Use exactly 2 requests (the limit)
	for i := 0; i < 2; i++ {
		middleware(next)(c)
	}

	// Act
	limited := rl.IsClientLimited("test-client-at-limit")

	// Assert
	if !limited {
		t.Error("Expected client to be limited when at limit")
	}
}

// TestIsClientLimited_ShouldReturnFalseAfterWindowExpiry tests edge case - window expired
func TestIsClientLimited_ShouldReturnFalseAfterWindowExpiry(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(1, time.Millisecond)
	middleware := rl.RateLimit()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("X-Real-IP", "test-client-expiry-limited")

	next := func(c echo.Context) error {
		return nil
	}

	// Make request to hit limit
	middleware(next)(c)

	// Wait for window to expire
	time.Sleep(10 * time.Millisecond)

	// Act
	limited := rl.IsClientLimited("test-client-expiry-limited")

	// Assert
	if limited {
		t.Error("Expected client to not be limited after window expiry")
	}
}

// ============================================================================
// getClientID Tests (Internal Function)
// ============================================================================

// TestGetClientID_ShouldUseXRealIPHeader tests X-Real-IP header priority
func TestGetClientID_ShouldUseXRealIPHeader(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	clientID := getClientID(c)

	// Assert
	if clientID != "192.168.1.100" {
		t.Errorf("Expected X-Real-IP '192.168.1.100', got '%s'", clientID)
	}
}

// TestGetClientID_ShouldUseXForwardedForHeader tests X-Forwarded-For header
func TestGetClientID_ShouldUseXForwardedForHeader(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	clientID := getClientID(c)

	// Assert
	if clientID != "10.0.0.1" {
		t.Errorf("Expected first IP '10.0.0.1', got '%s'", clientID)
	}
}

// TestGetClientID_ShouldFallbackToRemoteAddr tests fallback to remote address
func TestGetClientID_ShouldFallbackToRemoteAddr(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	clientID := getClientID(c)

	// Assert
	// Should return some valid IP (from the test server)
	if clientID == "" {
		t.Error("Expected non-empty client ID")
	}
}
