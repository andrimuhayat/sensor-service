package usecase

import (
	"errors"
	"net/http"
	"sensor-service/internal/module/auth/entity"
	"sensor-service/internal/platform/app"
	"sensor-service/internal/platform/httpengine/httpresponse"
	"testing"
	"time"
)

// MockGenericRepository implements IGenericRepository for testing
type MockGenericRepository struct {
	FindByEmailFunc    func(T any, email string) (d *any, err error)
	UpdatePasswordFunc func(email string, newPassword string) error
	FindByIDFunc       func(T any, id int) (d *any, err error)
	CreateFunc         func(T any) error
	UpdateFunc         func(T any) error
	FindByFunc         func(T any, R any) (d *any, err error)
	FindAllFunc        func(T any) ([]*any, error)
	FindAllByFunc      func(T any, R any) ([]*any, error)
	DeleteByIDFunc     func(T any, id int) error
}

func (m *MockGenericRepository) FindByEmail(T any, email string) (d *any, err error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(T, email)
	}
	return nil, nil
}

func (m *MockGenericRepository) UpdatePassword(email string, newPassword string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(email, newPassword)
	}
	return nil
}

func (m *MockGenericRepository) FindByID(T any, id int) (d *any, err error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(T, id)
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

// Helper function to create test user
func createTestUser(email, password, role string) entity.User {
	return entity.User{
		Email:    email,
		Password: password,
		Role:     role,
	}
}

// Helper function to create test app config
func createTestAppConfig() app.App {
	return app.App{
		SecretKey: "test-secret-key-for-jwt-generation",
	}
}

// ============================================================================
// InitiateForgotPassword Tests
// ============================================================================

// TestUseCase_InitiateForgotPassword_ShouldReturnResetTokenWhenEmailValid tests successful reset token generation
func TestUseCase_InitiateForgotPassword_ShouldReturnResetTokenWhenEmailValid(t *testing.T) {
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

	useCase := NewUseCase(mockRepo).(UseCase)

	// Act
	resetToken, err := useCase.InitiateForgotPassword("user@example.com")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resetToken == "" {
		t.Error("Expected non-empty reset token")
	}
	if len(resetToken) != 64 { // 32 bytes = 64 hex characters
		t.Errorf("Expected token length 64, got %d", len(resetToken))
	}
}

// TestUseCase_InitiateForgotPassword_ShouldReturnErrorWhenEmailNotFound tests reset initiation with non-existent email
func TestUseCase_InitiateForgotPassword_ShouldReturnErrorWhenEmailNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		return nil, nil
	}

	useCase := NewUseCase(mockRepo).(UseCase)

	// Act
	resetToken, err := useCase.InitiateForgotPassword("nonexistent@example.com")

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

// TestUseCase_InitiateForgotPassword_ShouldReturnErrorWhenEmailEmpty tests reset initiation with empty email
func TestUseCase_InitiateForgotPassword_ShouldReturnErrorWhenEmailEmpty(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo).(UseCase)

	// Act
	resetToken, err := useCase.InitiateForgotPassword("")

	// Assert
	if err == nil {
		t.Error("Expected error for empty email")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
	if resetToken != "" {
		t.Error("Expected empty reset token on error")
	}
}

// ============================================================================
// ValidateResetToken Tests
// ============================================================================

// TestUseCase_ValidateResetToken_ShouldReturnEmailWhenTokenValid tests successful token validation
func TestUseCase_ValidateResetToken_ShouldReturnEmailWhenTokenValid(t *testing.T) {
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

	useCase := NewUseCase(mockRepo).(UseCase)

	// Create a valid reset token
	resetToken := "valid-reset-token-123"
	useCase.resetTokenCache.mu.Lock()
	useCase.resetTokenCache.tokens[resetToken] = &PasswordResetToken{
		Token:     resetToken,
		Email:     "user@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
	}
	useCase.resetTokenCache.mu.Unlock()

	// Act
	email, err := useCase.ValidateResetToken(resetToken)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if email != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", email)
	}
}

// TestUseCase_ValidateResetToken_ShouldReturnErrorWhenTokenInvalid tests validation with invalid token
func TestUseCase_ValidateResetToken_ShouldReturnErrorWhenTokenInvalid(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo).(UseCase)

	// Act
	email, err := useCase.ValidateResetToken("invalid-token")

	// Assert
	if err == nil {
		t.Error("Expected error for invalid token")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
	if email != "" {
		t.Error("Expected empty email on error")
	}
}

// TestUseCase_ValidateResetToken_ShouldReturnErrorWhenTokenExpired tests validation with expired token
func TestUseCase_ValidateResetToken_ShouldReturnErrorWhenTokenExpired(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo).(UseCase)

	// Create an expired reset token
	expiredToken := "expired-reset-token"
	useCase.resetTokenCache.mu.Lock()
	useCase.resetTokenCache.tokens[expiredToken] = &PasswordResetToken{
		Token:     expiredToken,
		Email:     "user@example.com",
		ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired 1 minute ago
		CreatedAt: time.Now().Add(-20 * time.Minute),
	}
	useCase.resetTokenCache.mu.Unlock()

	// Act
	email, err := useCase.ValidateResetToken(expiredToken)

	// Assert
	if err == nil {
		t.Error("Expected error for expired token")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
	if email != "" {
		t.Error("Expected empty email on error")
	}
}

// ============================================================================
// CompleteForgotPassword Tests
// ============================================================================

// TestUseCase_CompleteForgotPassword_ShouldUpdatePasswordWhenTokenValid tests successful password update
func TestUseCase_CompleteForgotPassword_ShouldUpdatePasswordWhenTokenValid(t *testing.T) {
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
	mockRepo.UpdatePasswordFunc = func(email string, newPassword string) error {
		return nil
	}

	useCase := NewUseCase(mockRepo).(UseCase)

	// Create a valid reset token
	resetToken := "valid-reset-token-123"
	useCase.resetTokenCache.mu.Lock()
	useCase.resetTokenCache.tokens[resetToken] = &PasswordResetToken{
		Token:     resetToken,
		Email:     "user@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
	}
	useCase.resetTokenCache.mu.Unlock()

	// Act
	err := useCase.CompleteForgotPassword(resetToken, "newPassword123")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify token is invalidated after use
	useCase.resetTokenCache.mu.RLock()
	_, exists := useCase.resetTokenCache.tokens[resetToken]
	useCase.resetTokenCache.mu.RUnlock()

	if exists {
		t.Error("Expected token to be invalidated after password reset")
	}
}

// TestUseCase_CompleteForgotPassword_ShouldReturnErrorWhenTokenInvalid tests password reset with invalid token
func TestUseCase_CompleteForgotPassword_ShouldReturnErrorWhenTokenInvalid(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo).(UseCase)

	// Act
	err := useCase.CompleteForgotPassword("invalid-token", "newPassword123")

	// Assert
	if err == nil {
		t.Error("Expected error for invalid token")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
}

// TestUseCase_CompleteForgotPassword_ShouldReturnErrorWhenPasswordEmpty tests password reset with empty password
func TestUseCase_CompleteForgotPassword_ShouldReturnErrorWhenPasswordEmpty(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo).(UseCase)

	// Create a valid reset token
	resetToken := "valid-reset-token-123"
	useCase.resetTokenCache.mu.Lock()
	useCase.resetTokenCache.tokens[resetToken] = &PasswordResetToken{
		Token:     resetToken,
		Email:     "user@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
	}
	useCase.resetTokenCache.mu.Unlock()

	// Act
	err := useCase.CompleteForgotPassword(resetToken, "")

	// Assert
	if err == nil {
		t.Error("Expected error for empty password")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
}

// TestUseCase_CompleteForgotPassword_ShouldReturnErrorWhenTokenExpired tests password reset with expired token
func TestUseCase_CompleteForgotPassword_ShouldReturnErrorWhenTokenExpired(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo).(UseCase)

	// Create an expired reset token
	expiredToken := "expired-reset-token"
	useCase.resetTokenCache.mu.Lock()
	useCase.resetTokenCache.tokens[expiredToken] = &PasswordResetToken{
		Token:     expiredToken,
		Email:     "user@example.com",
		ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired 1 minute ago
		CreatedAt: time.Now().Add(-20 * time.Minute),
	}
	useCase.resetTokenCache.mu.Unlock()

	// Act
	err := useCase.CompleteForgotPassword(expiredToken, "newPassword123")

	// Assert
	if err == nil {
		t.Error("Expected error for expired token")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
}

// TestUseCase_CompleteForgotPassword_ShouldReturnErrorWhenResetTokenEmpty tests password reset with empty token
func TestUseCase_CompleteForgotPassword_ShouldReturnErrorWhenResetTokenEmpty(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo).(UseCase)

	// Act
	err := useCase.CompleteForgotPassword("", "newPassword123")

	// Assert
	if err == nil {
		t.Error("Expected error for empty reset token")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.Code)
	}
}

// ============================================================================
// Token Cache Tests
// ============================================================================

// TestUseCase_ResetTokenCache_ShouldStoreAndRetrieveToken tests token cache storage and retrieval
func TestUseCase_ResetTokenCache_ShouldStoreAndRetrieveToken(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo).(UseCase)

	resetToken := "test-token-123"
	testEmail := "user@example.com"

	// Act - Store token
	useCase.resetTokenCache.mu.Lock()
	useCase.resetTokenCache.tokens[resetToken] = &PasswordResetToken{
		Token:     resetToken,
		Email:     testEmail,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
	}
	useCase.resetTokenCache.mu.Unlock()

	// Retrieve token
	useCase.resetTokenCache.mu.RLock()
	pwdReset, exists := useCase.resetTokenCache.tokens[resetToken]
	useCase.resetTokenCache.mu.RUnlock()

	// Assert
	if !exists {
		t.Error("Expected token to be stored in cache")
	}
	if pwdReset.Email != testEmail {
		t.Errorf("Expected email '%s', got '%s'", testEmail, pwdReset.Email)
	}
}

// TestUseCase_ResetTokenCache_ShouldHandleConcurrentAccess tests concurrent token cache access
func TestUseCase_ResetTokenCache_ShouldHandleConcurrentAccess(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	useCase := NewUseCase(mockRepo).(UseCase)

	// Act - Simulate concurrent writes and reads
	done := make(chan bool)

	// Write goroutine
	go func() {
		for i := 0; i < 10; i++ {
			token := "token-" + string(rune(i))
			useCase.resetTokenCache.mu.Lock()
			useCase.resetTokenCache.tokens[token] = &PasswordResetToken{
				Token:     token,
				Email:     "user@example.com",
				ExpiresAt: time.Now().Add(15 * time.Minute),
			}
			useCase.resetTokenCache.mu.Unlock()
		}
		done <- true
	}()

	// Read goroutine
	go func() {
		for i := 0; i < 10; i++ {
			useCase.resetTokenCache.mu.RLock()
			_ = len(useCase.resetTokenCache.tokens)
			useCase.resetTokenCache.mu.RUnlock()
		}
		done <- true
	}()

	// Wait for goroutines
	<-done
	<-done

	// Assert - No panic means test passed
	useCase.resetTokenCache.mu.RLock()
	if len(useCase.resetTokenCache.tokens) == 0 {
		t.Error("Expected tokens to be stored during concurrent access")
	}
	useCase.resetTokenCache.mu.RUnlock()
}

// ============================================================================
// Error Handling Tests
// ============================================================================

// TestUseCase_InitiateForgotPassword_ShouldReturnInternalErrorWhenDBFails tests DB error handling
func TestUseCase_InitiateForgotPassword_ShouldReturnInternalErrorWhenDBFails(t *testing.T) {
	// Arrange
	mockRepo := &MockGenericRepository{}
	mockRepo.FindByEmailFunc = func(T any, email string) (d *any, err error) {
		return nil, errors.New("database connection failed")
	}

	useCase := NewUseCase(mockRepo).(UseCase)

	// Act
	resetToken, err := useCase.InitiateForgotPassword("user@example.com")

	// Assert
	if err == nil {
		t.Error("Expected error for database failure")
	}
	if err.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, err.Code)
	}
	if resetToken != "" {
		t.Error("Expected empty reset token on error")
	}
}

// TestUseCase_CompleteForgotPassword_ShouldReturnErrorWhenUpdateFails tests password update failure
func TestUseCase_CompleteForgotPassword_ShouldReturnErrorWhenUpdateFails(t *testing.T) {
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
	mockRepo.UpdatePasswordFunc = func(email string, newPassword string) error {
		return errors.New("database update failed")
	}

	useCase := NewUseCase(mockRepo).(UseCase)

	// Create a valid reset token
	resetToken := "valid-reset-token-123"
	useCase.resetTokenCache.mu.Lock()
	useCase.resetTokenCache.tokens[resetToken] = &PasswordResetToken{
		Token:     resetToken,
		Email:     "user@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
	}
	useCase.resetTokenCache.mu.Unlock()

	// Act
	err := useCase.CompleteForgotPassword(resetToken, "newPassword123")

	// Assert
	if err == nil {
		t.Error("Expected error for password update failure")
	}
	if err.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, err.Code)
	}
}
