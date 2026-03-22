package usecase

import (
	"errors"
	"net/http"
	"sensor-service/config"
	"sensor-service/internal/module/auth/dto"
	"sensor-service/internal/module/auth/entity"
	"sensor-service/internal/platform/app"
	"sensor-service/internal/platform/httpengine/httpresponse"
	"testing"
	"time"
)

// MockGenericRepository implements IGenericRepository for testing
type MockGenericRepository struct {
	FindByEmailFunc func(T any, email string) (d *any, err error)
	CreateFunc      func(T any) error
	UpdateFunc      func(T any) error
	FindByIDFunc    func(T any, id int) (d *any, err error)
	FindByFunc      func(T any, R any) (d *any, err error)
	FindAllFunc     func(T any) ([]*any, error)
	FindAllByFunc   func(T any, R any) ([]*any, error)
	DeleteByIDFunc  func(T any, id int) error
}

func (m *MockGenericRepository) FindByEmail(T any, email string) (d *any, err error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(T, email)
	}
	return nil, nil
}

func (m *MockGenericRepository) Create(T any) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(T)
	}
	return nil
}

func (m *MockGenericRepository) Update(T any) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(T)
	}
	return nil
}

func (m *MockGenericRepository) FindByID(T any, id int) (d *any, err error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(T, id)
	}
	return nil, nil
}

func (m *MockGenericRepository) FindBy(T any, R any) (d *any, err error) {
	if m.FindByFunc != nil {
		return m.FindByFunc(T, R)
	}
	return nil, nil
}

func (m *MockGenericRepository) FindAll(T any) ([]*any, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(T)
	}
	return nil, nil
}

func (m *MockGenericRepository) FindAllBy(T any, R any) ([]*any, error) {
	if m.FindAllByFunc != nil {
		return m.FindAllByFunc(T, R)
	}
	return nil, nil
}

func (m *MockGenericRepository) DeleteByID(T any, id int) error {
	if m.DeleteByIDFunc != nil {
		return m.DeleteByIDFunc(T, id)
	}
	return nil
}

// Helper function to create test app config
func createTestAppConfig() app.App {
	return app.App{
		SecretKey: "test-secret-key-for-jwt-generation",
	}
}

// Helper function to create test user
func createTestUser(email, password, role string) entity.User {
	return entity.User{
		Email:    email,
		Password: password,
		Role:     role,
	}
}

// Helper function to create test authentication request
func createTestAuthRequest(email, password string) config.HTTPRequest {
	return config.HTTPRequest{
		Body: map[string]interface{}{
			"email":    email,
			"password": password,
		},
	}
}

// ============================================================================
// SignIn Tests
// ============================================================================

// TestUseCase_SignIn_ShouldReturnTokenWhenCredentialsValid tests successful sign in
func TestUseCase_SignIn_ShouldReturnTokenWhenCredentialsValid(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	testUser := createTestUser("user@example.com", "hashedPassword123", "user")
	userPtr := any(testUser)

	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		if email == "user@example.com" {
			return &userPtr, nil
		}
		return nil, nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)
	request := createTestAuthRequest("user@example.com", "password123")

	// Act
	token, err := useCase.SignIn(request)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if token.Email != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", token.Email)
	}
	if token.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", token.Role)
	}
	if token.TokenString == "" {
		t.Error("Expected non-empty token string")
	}
}

// TestUseCase_SignIn_ShouldReturnErrorWhenUserNotFound tests sign in with non-existent user
func TestUseCase_SignIn_ShouldReturnErrorWhenUserNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		return nil, nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)
	request := createTestAuthRequest("nonexistent@example.com", "password123")

	// Act
	token, err := useCase.SignIn(request)

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
	if token.TokenString != "" {
		t.Error("Expected empty token for failed sign in")
	}
}

// TestUseCase_SignIn_ShouldReturnErrorWhenPasswordIncorrect tests sign in with wrong password
func TestUseCase_SignIn_ShouldReturnErrorWhenPasswordIncorrect(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	testUser := createTestUser("user@example.com", "$2a$10$wronghash", "user")
	userPtr := any(testUser)

	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		if email == "user@example.com" {
			return &userPtr, nil
		}
		return nil, nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)
	request := createTestAuthRequest("user@example.com", "wrongpassword")

	// Act
	token, err := useCase.SignIn(request)

	// Assert
	if err == nil {
		t.Error("Expected error for incorrect password")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
}

// TestUseCase_SignIn_ShouldEnforceRateLimitOn5Attempts tests rate limiting
func TestUseCase_SignIn_ShouldEnforceRateLimitOn5Attempts(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		return nil, nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)
	request := createTestAuthRequest("user@example.com", "wrongpassword")

	// Act - Make 5 failed attempts
	var lastErr *httpresponse.HTTPError
	for i := 0; i < 5; i++ {
		_, lastErr = useCase.SignIn(request)
	}

	// 6th attempt should be rate limited
	_, rateLimitErr := useCase.SignIn(request)

	// Assert
	if rateLimitErr == nil {
		t.Error("Expected rate limit error on 6th attempt")
	}
	if rateLimitErr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, rateLimitErr.Code)
	}
}

// ============================================================================
// SignUp Tests
// ============================================================================

// TestUseCase_SignUp_ShouldCreateUserWhenValidInput tests successful user creation
func TestUseCase_SignUp_ShouldCreateUserWhenValidInput(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		return nil, nil
	}
	mockRepo.CreateFunc = func(T any) error {
		return nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)
	request := config.HTTPRequest{
		Body: map[string]interface{}{
			"email":    "newuser@example.com",
			"password": "password123",
			"roles":    "user",
		},
	}

	// Act
	user, err := useCase.SignUp(request)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user.Email != "newuser@example.com" {
		t.Errorf("Expected email 'newuser@example.com', got '%s'", user.Email)
	}
	if user.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", user.Role)
	}
}

// TestUseCase_SignUp_ShouldReturnErrorWhenEmailExists tests sign up with duplicate email
func TestUseCase_SignUp_ShouldReturnErrorWhenEmailExists(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	existingUser := createTestUser("existing@example.com", "hashedPassword", "user")
	userPtr := any(existingUser)

	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		if email == "existing@example.com" {
			return &userPtr, nil
		}
		return nil, nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)
	request := config.HTTPRequest{
		Body: map[string]interface{}{
			"email":    "existing@example.com",
			"password": "password123",
			"roles":    "user",
		},
	}

	// Act
	user, err := useCase.SignUp(request)

	// Assert
	if err == nil {
		t.Error("Expected error for duplicate email")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
	if user.Email != "" {
		t.Error("Expected empty user on error")
	}
}

// TestUseCase_SignUp_ShouldReturnErrorWhenInvalidRole tests sign up with invalid role
func TestUseCase_SignUp_ShouldReturnErrorWhenInvalidRole(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		return nil, nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)
	request := config.HTTPRequest{
		Body: map[string]interface{}{
			"email":    "newuser@example.com",
			"password": "password123",
			"roles":    "superadmin", // Invalid role
		},
	}

	// Act
	user, err := useCase.SignUp(request)

	// Assert
	if err == nil {
		t.Error("Expected error for invalid role")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
}

// ============================================================================
// RefreshToken Tests
// ============================================================================

// TestUseCase_RefreshToken_ShouldReturnNewTokenWhenValid tests token refresh with valid token
func TestUseCase_RefreshToken_ShouldReturnNewTokenWhenValid(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Cache a valid token
	oldToken := "valid-token-123"
	useCase.cacheToken(oldToken, "user@example.com", "user")

	// Act
	newToken, err := useCase.RefreshToken(oldToken)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if newToken.Email != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", newToken.Email)
	}
	if newToken.TokenString == "" {
		t.Error("Expected non-empty new token")
	}
	if newToken.TokenString == oldToken {
		t.Error("Expected new token to be different from old token")
	}
}

// TestUseCase_RefreshToken_ShouldReturnErrorWhenTokenNotFound tests refresh with invalid token
func TestUseCase_RefreshToken_ShouldReturnErrorWhenTokenNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	_, err := useCase.RefreshToken("nonexistent-token")

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent token")
	}
	if err.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, err.Code)
	}
}

// TestUseCase_RefreshToken_ShouldReturnErrorWhenTokenExpired tests refresh with expired token
func TestUseCase_RefreshToken_ShouldReturnErrorWhenTokenExpired(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Cache an expired token
	expiredToken := "expired-token-123"
	useCase.tokenCache.mu.Lock()
	useCase.tokenCache.tokens[expiredToken] = &TokenMetadata{
		Email:     "user@example.com",
		Role:      "user",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-25 * time.Hour),
	}
	useCase.tokenCache.mu.Unlock()

	// Act
	_, err := useCase.RefreshToken(expiredToken)

	// Assert
	if err == nil {
		t.Error("Expected error for expired token")
	}
	if err.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, err.Code)
	}
}

// ============================================================================
// InitiatePasswordReset Tests
// ============================================================================

// TestUseCase_InitiatePasswordReset_ShouldReturnResetTokenWhenEmailValid tests password reset initiation
func TestUseCase_InitiatePasswordReset_ShouldReturnResetTokenWhenEmailValid(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	testUser := createTestUser("user@example.com", "hashedPassword", "user")
	userPtr := any(testUser)

	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		if email == "user@example.com" {
			return &userPtr, nil
		}
		return nil, nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	resetToken, err := useCase.InitiatePasswordReset("user@example.com")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resetToken == "" {
		t.Error("Expected non-empty reset token")
	}
}

// TestUseCase_InitiatePasswordReset_ShouldReturnErrorWhenEmailNotFound tests reset with non-existent email
func TestUseCase_InitiatePasswordReset_ShouldReturnErrorWhenEmailNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		return nil, nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	resetToken, err := useCase.InitiatePasswordReset("nonexistent@example.com")

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
	if resetToken != "" {
		t.Error("Expected empty reset token on error")
	}
}

// ============================================================================
// ResetPassword Tests
// ============================================================================

// TestUseCase_ResetPassword_ShouldUpdatePasswordWhenTokenValid tests password reset with valid token
func TestUseCase_ResetPassword_ShouldUpdatePasswordWhenTokenValid(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	testUser := createTestUser("user@example.com", "oldHashedPassword", "user")
	userPtr := any(testUser)

	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		if email == "user@example.com" {
			return &userPtr, nil
		}
		return nil, nil
	}
	mockRepo.UpdateFunc = func(T any) error {
		return nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Create a valid reset token
	resetToken := "valid-reset-token"
	useCase.resetTokenMu.Lock()
	useCase.resetTokens[resetToken] = &PasswordResetToken{
		Token:     resetToken,
		Email:     "user@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	useCase.resetTokenMu.Unlock()

	// Act
	err := useCase.ResetPassword(resetToken, "newPassword123")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestUseCase_ResetPassword_ShouldReturnErrorWhenTokenInvalid tests reset with invalid token
func TestUseCase_ResetPassword_ShouldReturnErrorWhenTokenInvalid(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	err := useCase.ResetPassword("invalid-token", "newPassword123")

	// Assert
	if err == nil {
		t.Error("Expected error for invalid token")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
}

// TestUseCase_ResetPassword_ShouldReturnErrorWhenTokenExpired tests reset with expired token
func TestUseCase_ResetPassword_ShouldReturnErrorWhenTokenExpired(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Create an expired reset token
	expiredToken := "expired-reset-token"
	useCase.resetTokenMu.Lock()
	useCase.resetTokens[expiredToken] = &PasswordResetToken{
		Token:     expiredToken,
		Email:     "user@example.com",
		ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired 1 minute ago
	}
	useCase.resetTokenMu.Unlock()

	// Act
	err := useCase.ResetPassword(expiredToken, "newPassword123")

	// Assert
	if err == nil {
		t.Error("Expected error for expired token")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
}

// ============================================================================
// ChangePassword Tests
// ============================================================================

// TestUseCase_ChangePassword_ShouldUpdatePasswordWhenOldPasswordCorrect tests successful password change
func TestUseCase_ChangePassword_ShouldUpdatePasswordWhenOldPasswordCorrect(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	// Use bcrypt hash of "oldpassword123" for the stored password
	// $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.rsS/lW1pCP8q6fVZHG (example hash)
	testUser := createTestUser("user@example.com", "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.rsS/lW1pCP8q6fVZHG", "user")
	userPtr := any(testUser)

	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		if email == "user@example.com" {
			return &userPtr, nil
		}
		return nil, nil
	}
	mockRepo.UpdateFunc = func(T any) error {
		return nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act - Note: The actual password verification uses bcrypt, so we need to use "oldpassword123" which hashes to the stored value
	err := useCase.ChangePassword("user@example.com", "oldpassword123", "newpassword456")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestUseCase_ChangePassword_ShouldReturnErrorWhenUserNotFound tests password change for non-existent user
func TestUseCase_ChangePassword_ShouldReturnErrorWhenUserNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		return nil, nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	err := useCase.ChangePassword("nonexistent@example.com", "oldpassword123", "newpassword456")

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
	if err.Message != "User not found" {
		t.Errorf("Expected message 'User not found', got '%s'", err.Message)
	}
}

// TestUseCase_ChangePassword_ShouldReturnErrorWhenOldPasswordIncorrect tests password change with wrong old password
func TestUseCase_ChangePassword_ShouldReturnErrorWhenOldPasswordIncorrect(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	testUser := createTestUser("user@example.com", "$2a$10$hashedpassword", "user")
	userPtr := any(testUser)

	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		if email == "user@example.com" {
			return &userPtr, nil
		}
		return nil, nil
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act - Using wrong old password
	err := useCase.ChangePassword("user@example.com", "wrongpassword", "newpassword456")

	// Assert
	if err == nil {
		t.Error("Expected error for incorrect old password")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
	if err.Message != "Old password is incorrect" {
		t.Errorf("Expected message 'Old password is incorrect', got '%s'", err.Message)
	}
}

// TestUseCase_ChangePassword_ShouldReturnErrorWhenDatabaseUpdateFails tests password change when DB update fails
func TestUseCase_ChangePassword_ShouldReturnErrorWhenDatabaseUpdateFails(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	testUser := createTestUser("user@example.com", "$2a$10$hashedpassword", "user")
	userPtr := any(testUser)

	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		if email == "user@example.com" {
			return &userPtr, nil
		}
		return nil, nil
	}
	mockRepo.UpdateFunc = func(T any) error {
		return errors.New("database connection error")
	}

	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	err := useCase.ChangePassword("user@example.com", "oldpassword123", "newpassword456")

	// Assert
	if err == nil {
		t.Error("Expected error when database update fails")
	}
	if err.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, err.Code)
	}
}

// ============================================================================
// ValidateSession Tests
// ============================================================================

// TestUseCase_ValidateSession_ShouldReturnTrueWhenSessionValid tests session validation with active session
func TestUseCase_ValidateSession_ShouldReturnTrueWhenSessionValid(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Create an active session
	useCase.createSession("user@example.com")

	// Act
	isValid, err := useCase.ValidateSession("user@example.com")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !isValid {
		t.Error("Expected session to be valid")
	}
}

// TestUseCase_ValidateSession_ShouldReturnFalseWhenSessionNotFound tests validation with non-existent session
func TestUseCase_ValidateSession_ShouldReturnFalseWhenSessionNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	isValid, err := useCase.ValidateSession("nonexistent@example.com")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if isValid {
		t.Error("Expected session to be invalid")
	}
}

// TestUseCase_ValidateSession_ShouldReturnFalseWhenSessionExpired tests validation with expired session
func TestUseCase_ValidateSession_ShouldReturnFalseWhenSessionExpired(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Create an expired session
	useCase.sessionStore.mu.Lock()
	useCase.sessionStore.sessions["user@example.com"] = &SessionData{
		UserEmail:     "user@example.com",
		LoginAttempts: 0,
		LastAttemptAt: time.Now(),
		SessionExpiry: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		RefreshTokens: make(map[string]time.Time),
	}
	useCase.sessionStore.mu.Unlock()

	// Act
	isValid, err := useCase.ValidateSession("user@example.com")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if isValid {
		t.Error("Expected expired session to be invalid")
	}
}

// ============================================================================
// CheckRateLimit Tests
// ============================================================================

// TestUseCase_CheckRateLimit_ShouldAllowFirstAttempt tests rate limit allows first attempt
func TestUseCase_CheckRateLimit_ShouldAllowFirstAttempt(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	allowed, err := useCase.CheckRateLimit("user@example.com")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !allowed {
		t.Error("Expected first attempt to be allowed")
	}
}

// TestUseCase_CheckRateLimit_ShouldAllow5AttemptsWithin15Minutes tests rate limit allows up to 5 attempts
func TestUseCase_CheckRateLimit_ShouldAllow5AttemptsWithin15Minutes(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act - Make 5 attempts
	for i := 0; i < 5; i++ {
		allowed, err := useCase.CheckRateLimit("user@example.com")
		if !allowed || err != nil {
			t.Errorf("Attempt %d: expected to be allowed", i+1)
		}
	}
}

// TestUseCase_CheckRateLimit_ShouldBlockOn6thAttempt tests rate limit blocks 6th attempt
func TestUseCase_CheckRateLimit_ShouldBlockOn6thAttempt(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act - Make 5 attempts
	for i := 0; i < 5; i++ {
		useCase.CheckRateLimit("user@example.com")
	}

	// 6th attempt should be blocked
	allowed, err := useCase.CheckRateLimit("user@example.com")

	// Assert
	if allowed {
		t.Error("Expected 6th attempt to be blocked")
	}
	if err == nil {
		t.Error("Expected error for rate limit exceeded")
	}
	if err.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, err.Code)
	}
}

// TestUseCase_CheckRateLimit_ShouldResetCounterAfter15Minutes tests rate limit counter resets after 15 minutes
func TestUseCase_CheckRateLimit_ShouldResetCounterAfter15Minutes(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Make 5 attempts
	for i := 0; i < 5; i++ {
		useCase.CheckRateLimit("user@example.com")
	}

	// Manually set last attempt to 16 minutes ago to simulate time passing
	useCase.sessionStore.mu.Lock()
	useCase.sessionStore.sessions["user@example.com"].LastAttemptAt = time.Now().Add(-16 * time.Minute)
	useCase.sessionStore.mu.Unlock()

	// Act - Next attempt should be allowed (counter reset)
	allowed, err := useCase.CheckRateLimit("user@example.com")

	// Assert
	if err != nil {
		t.Errorf("Expected no error after reset, got %v", err)
	}
	if !allowed {
		t.Error("Expected attempt to be allowed after 15 minute window")
	}
}

// ============================================================================
// Logout Tests
// ============================================================================

// TestUseCase_Logout_ShouldInvalidateTokenWhenValid tests logout with valid token
func TestUseCase_Logout_ShouldInvalidateTokenWhenValid(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Cache a token and create session
	token := "valid-token-123"
	useCase.cacheToken(token, "user@example.com", "user")
	useCase.createSession("user@example.com")

	// Act
	err := useCase.Logout(token)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify token is invalidated
	useCase.tokenCache.mu.RLock()
	_, exists := useCase.tokenCache.tokens[token]
	useCase.tokenCache.mu.RUnlock()

	if exists {
		t.Error("Expected token to be invalidated after logout")
	}

	// Verify session is invalidated
	isValid, _ := useCase.ValidateSession("user@example.com")
	if isValid {
		t.Error("Expected session to be invalidated after logout")
	}
}

// TestUseCase_Logout_ShouldReturnErrorWhenTokenInvalid tests logout with invalid token
func TestUseCase_Logout_ShouldReturnErrorWhenTokenInvalid(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	err := useCase.Logout("invalid-token")

	// Assert
	if err == nil {
		t.Error("Expected error for invalid token")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

// TestUseCase_GenerateSecureToken_ShouldReturnNonEmptyToken tests secure token generation
func TestUseCase_GenerateSecureToken_ShouldReturnNonEmptyToken(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	token := useCase.generateSecureToken(32)

	// Assert
	if token == "" {
		t.Error("Expected non-empty token")
	}
	if len(token) != 64 { // 32 bytes = 64 hex characters
		t.Errorf("Expected token length 64, got %d", len(token))
	}
}

// TestUseCase_GenerateSecureToken_ShouldGenerateUniqueTokens tests token uniqueness
func TestUseCase_GenerateSecureToken_ShouldGenerateUniqueTokens(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	token1 := useCase.generateSecureToken(32)
	token2 := useCase.generateSecureToken(32)

	// Assert
	if token1 == token2 {
		t.Error("Expected tokens to be unique")
	}
}

// TestUseCase_CacheToken_ShouldStoreTokenMetadata tests token caching
func TestUseCase_CacheToken_ShouldStoreTokenMetadata(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	token := "test-token"
	useCase.cacheToken(token, "user@example.com", "user")

	// Assert
	useCase.tokenCache.mu.RLock()
	metadata, exists := useCase.tokenCache.tokens[token]
	useCase.tokenCache.mu.RUnlock()

	if !exists {
		t.Error("Expected token to be cached")
	}
	if metadata.Email != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", metadata.Email)
	}
	if metadata.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", metadata.Role)
	}
}

// TestUseCase_CreateSession_ShouldInitializeSessionData tests session creation
func TestUseCase_CreateSession_ShouldInitializeSessionData(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo, createTestAppConfig()).(UseCase)

	// Act
	useCase.createSession("user@example.com")

	// Assert
	useCase.sessionStore.mu.RLock()
	session, exists := useCase.sessionStore.sessions["user@example.com"]
	useCase.sessionStore.mu.RUnlock()

	if !exists {
		t.Error("Expected session to be created")
	}
	if session.UserEmail != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", session.UserEmail)
	}
	if session.LoginAttempts != 0 {
		t.Errorf("Expected login attempts 0, got %d", session.LoginAttempts)
	}
}
