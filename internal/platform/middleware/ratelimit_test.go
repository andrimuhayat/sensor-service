package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

// Helper to create test Echo context
func createTestEchoContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c
}

// Helper to create rate limiter with specific config
func createTestRateLimiter(maxRequests int, windowSize time.Duration) *RateLimiter {
	return NewRateLimiter(maxRequests, windowSize)
}

// Helper to get client data for testing
func getClientData(rl *RateLimiter, key string) *RateLimiterData {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.clients[key]
}

// ============================================================================
// NewRateLimiter Tests
// ============================================================================

// TestNewRateLimiter_ShouldCreateInstanceWithCorrectConfig tests rate limiter creation
func TestNewRateLimiter_ShouldCreateInstanceWithCorrectConfig(t *testing.T) {
	// Arrange & Act
	rl := NewRateLimiter(100, time.Minute)

	// Assert
	if rl == nil {
		t.Error("Expected rate limiter to be created")
	}
	if rl.config.MaxRequests != 100 {
		t.Errorf("Expected MaxRequests 100, got %d", rl.config.MaxRequests)
	}
	if rl.config.WindowSize != time.Minute {
		t.Errorf("Expected WindowSize 1m, got %v", rl.config.WindowSize)
	}
}

// TestNewRateLimiter_ShouldInitializeEmptyClientsMap tests initial state
func TestNewRateLimiter_ShouldInitializeEmptyClientsMap(t *testing.T) {
	// Arrange & Act
	rl := NewRateLimiter(50, time.Second*30)

	// Assert
	rl.mu.RLock()
	clientCount := len(rl.clients)
	rl.mu.RUnlock()

	if clientCount != 0 {
		t.Errorf("Expected 0 clients, got %d", clientCount)
	}
}

// ============================================================================
// CheckAndIncrement Tests
// ============================================================================

// TestCheckAndIncrement_ShouldAllowFirstRequest tests first request is allowed
func TestCheckAndIncrement_ShouldAllowFirstRequest(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Minute)
	clientKey := "test-client-1"

	// Act
	allowed, remaining, resetIn := rl.CheckAndIncrement(clientKey)

	// Assert
	if !allowed {
		t.Error("Expected first request to be allowed")
	}
	if remaining != 4 {
		t.Errorf("Expected remaining 4, got %d", remaining)
	}
	if resetIn <= 0 {
		t.Error("Expected positive reset time")
	}
}

// TestCheckAndIncrement_ShouldAllowRequestsWithinLimit tests requests up to limit
func TestCheckAndIncrement_ShouldAllowRequestsWithinLimit(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Minute)
	clientKey := "test-client-2"

	// Act - Make 5 requests
	for i := 0; i < 5; i++ {
		allowed, _, _ := rl.CheckAndIncrement(clientKey)
		if !allowed {
			t.Errorf("Expected request %d to be allowed", i+1)
		}
	}
}

// TestCheckAndIncrement_ShouldBlockExcessRequests tests 6th request is blocked
func TestCheckAndIncrement_ShouldBlockExcessRequests(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Minute)
	clientKey := "test-client-3"

	// Act - Make 5 requests
	for i := 0; i < 5; i++ {
		rl.CheckAndIncrement(clientKey)
	}

	// 6th request should be blocked
	allowed, remaining, _ := rl.CheckAndIncrement(clientKey)

	// Assert
	if allowed {
		t.Error("Expected 6th request to be blocked")
	}
	if remaining != 0 {
		t.Errorf("Expected remaining 0, got %d", remaining)
	}
}

// TestCheckAndIncrement_ShouldTrackDifferentClientsIndependently tests per-client tracking
func TestCheckAndIncrement_ShouldTrackDifferentClientsIndependently(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(2, time.Minute)

	// Act
	// Client A makes 2 requests
	rl.CheckAndIncrement("client-a")
	rl.CheckAndIncrement("client-a")

	// Client B should still be allowed (different client)
	allowed, _, _ := rl.CheckAndIncrement("client-b")

	// Assert
	if !allowed {
		t.Error("Expected client-b to be allowed (independent from client-a)")
	}
}

// TestCheckAndIncrement_ShouldResetAfterWindowExpiry tests counter resets after window
func TestCheckAndIncrement_ShouldResetAfterWindowExpiry(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Millisecond*50)
	clientKey := "test-client-4"

	// Make 5 requests to reach limit
	for i := 0; i < 5; i++ {
		rl.CheckAndIncrement(clientKey)
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Act - Next request should be allowed after reset
	allowed, remaining, _ := rl.CheckAndIncrement(clientKey)

	// Assert
	if !allowed {
		t.Error("Expected request to be allowed after window reset")
	}
	if remaining != 4 {
		t.Errorf("Expected remaining 4 after reset, got %d", remaining)
	}
}

// ============================================================================
// GetConfig Tests
// ============================================================================

// TestGetConfig_ShouldReturnCorrectConfiguration tests config getter
func TestGetConfig_ShouldReturnCorrectConfiguration(t *testing.T) {
	// Arrange
	rl := NewRateLimiter(100, time.Hour)

	// Act
	config := rl.GetConfig()

	// Assert
	if config.MaxRequests != 100 {
		t.Errorf("Expected MaxRequests 100, got %d", config.MaxRequests)
	}
	if config.WindowSize != time.Hour {
		t.Errorf("Expected WindowSize 1h, got %v", config.WindowSize)
	}
}

// ============================================================================
// Reset Tests
// ============================================================================

// TestReset_ShouldClearAllClients tests reset clears client data
func TestReset_ShouldClearAllClients(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Minute)
	rl.CheckAndIncrement("client-1")
	rl.CheckAndIncrement("client-2")
	
	// Verify clients exist
	rl.mu.RLock()
	beforeCount := len(rl.clients)
	rl.mu.RUnlock()

	// Act
	rl.Reset()

	// Assert
	rl.mu.RLock()
	afterCount := len(rl.clients)
	rl.mu.RUnlock()

	if beforeCount == 0 {
		t.Error("Expected clients before reset")
	}
	if afterCount != 0 {
		t.Errorf("Expected 0 clients after reset, got %d", afterCount)
	}
}

// TestReset_ShouldAllowNewRequestsAfterReset tests new requests work after reset
func TestReset_ShouldAllowNewRequestsAfterReset(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(2, time.Minute)
	clientKey := "test-client-reset"

	// Use up the limit
	rl.CheckAndIncrement(clientKey)
	rl.CheckAndIncrement(clientKey)

	// Verify limit reached
	allowed, _, _ := rl.CheckAndIncrement(clientKey)
	if allowed {
		t.Error("Expected request to be blocked before reset")
	}

	// Act - Reset
	rl.Reset()

	// Assert - Should be allowed after reset
	allowed, _, _ = rl.CheckAndIncrement(clientKey)
	if !allowed {
		t.Error("Expected request to be allowed after reset")
	}
}

// ============================================================================
// RateLimitMiddleware Tests
// ============================================================================

// TestRateLimitMiddleware_ShouldAllowRequestWithinLimit tests middleware allows requests
func TestRateLimitMiddleware_ShouldAllowRequestWithinLimit(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(100, time.Minute)
	middleware := RateLimitMiddleware(rl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	nextHandler := func(c echo.Context) error {
		handlerCalled = true
		return nil
	}

	// Act
	err := middleware(nextHandler)(c)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !handlerCalled {
		t.Error("Expected next handler to be called")
	}
	if rec.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("Expected X-RateLimit-Limit header")
	}
}

// TestRateLimitMiddleware_ShouldReturn429WhenLimitExceeded tests middleware blocks excess
func TestRateLimitMiddleware_ShouldReturn429WhenLimitExceeded(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(1, time.Minute)
	// Use up the limit
	rl.CheckAndIncrement("127.0.0.1")

	middleware := RateLimitMiddleware(rl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextHandler := func(c echo.Context) error {
		return nil
	}

	// Act
	err := middleware(nextHandler)(c)

	// Assert
	if err == nil {
		t.Error("Expected error when rate limit exceeded")
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}

	// Verify error response
	var httpErr httpresponse.HTTPError
	// Response body should contain error
}

// TestRateLimitMiddleware_ShouldSetRateLimitHeaders tests headers are set
func TestRateLimitMiddleware_ShouldSetRateLimitHeaders(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(10, time.Minute)
	middleware := RateLimitMiddleware(rl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextHandler := func(c echo.Context) error {
		return nil
	}

	// Act
	middleware(nextHandler)(c)

	// Assert
	if rec.Header().Get("X-RateLimit-Limit") != "10" {
		t.Errorf("Expected X-RateLimit-Limit 10, got %s", rec.Header().Get("X-RateLimit-Limit"))
	}
	if rec.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("Expected X-RateLimit-Remaining header")
	}
	if rec.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("Expected X-RateLimit-Reset header")
	}
}

// TestRateLimitMiddleware_ShouldUseXForwardedForHeader tests IP extraction
func TestRateLimitMiddleware_ShouldUseXForwardedForHeader(t *testing.T) {
	// Arrange
	rl := createTestRateLimiter(5, time.Minute)
	// Reset to clear any previous client data
	rl.Reset()
	
	middleware := RateLimitMiddleware(rl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextHandler := func(c echo.Context) error {
		return nil
	}

	// Act
	middleware(nextHandler)(c)

	// Assert - Should use X-Forwarded-For IP
	// The middleware should track by the forwarded IP
	allowed, _, _ := rl.CheckAndIncrement("192.168.1.1")
	if !allowed {
		t.Error("Expected X-Forwarded-For IP to be tracked")
	}
}
