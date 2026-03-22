package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"sensor-service/internal/module/auth/dto"
	"sensor-service/internal/module/auth/entity"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

// MockUseCase implements auth.IUseCase for testing
type MockUseCase struct {
	SignInFunc               func(request interface{}) (dto.Token, *httpresponse.HTTPError)
	SignUpFunc                func(request interface{}) (entity.User, *httpresponse.HTTPError)
	RefreshTokenFunc         func(oldToken string) (dto.Token, *httpresponse.HTTPError)
	InitiatePasswordResetFunc func(email string) (string, *httpresponse.HTTPError)
	ResetPasswordFunc         func(resetToken string, newPassword string) *httpresponse.HTTPError
	ValidateSessionFunc       func(email string) (bool, *httpresponse.HTTPError)
	CheckRateLimitFunc        func(email string) (bool, *httpresponse.HTTPError)
	LogoutFunc                func(token string) *httpresponse.HTTPError
	ChangePasswordFunc        func(email string, oldPassword string, newPassword string) *httpresponse.HTTPError
	RemoveUserFunc            func(email string) *httpresponse.HTTPError
}

func (m *MockUseCase) SignIn(request interface{}) (dto.Token, *httpresponse.HTTPError) {
	if m.SignInFunc != nil {
		return m.SignInFunc(request)
	}
	return dto.Token{}, nil
}

func (m *MockUseCase) SignUp(request interface{}) (entity.User, *httpresponse.HTTPError) {
	if m.SignUpFunc != nil {
		return m.SignUpFunc(request)
	}
	return entity.User{}, nil
}

func (m *MockUseCase) RefreshToken(oldToken string) (dto.Token, *httpresponse.HTTPError) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(oldToken)
	}
	return dto.Token{}, nil
}

func (m *MockUseCase) InitiatePasswordReset(email string) (string, *httpresponse.HTTPError) {
	if m.InitiatePasswordResetFunc != nil {
		return m.InitiatePasswordResetFunc(email)
	}
	return "", nil
}

func (m *MockUseCase) ResetPassword(resetToken string, newPassword string) *httpresponse.HTTPError {
	if m.ResetPasswordFunc != nil {
		return m.ResetPasswordFunc(resetToken, newPassword)
	}
	return nil
}

func (m *MockUseCase) ValidateSession(email string) (bool, *httpresponse.HTTPError) {
	if m.ValidateSessionFunc != nil {
		return m.ValidateSessionFunc(email)
	}
	return false, nil
}

func (m *MockUseCase) CheckRateLimit(email string) (bool, *httpresponse.HTTPError) {
	if m.CheckRateLimitFunc != nil {
		return m.CheckRateLimitFunc(email)
	}
	return true, nil
}

func (m *MockUseCase) Logout(token string) *httpresponse.HTTPError {
	if m.LogoutFunc != nil {
		return m.LogoutFunc(token)
	}
	return nil
}

func (m *MockUseCase) ChangePassword(email string, oldPassword string, newPassword string) *httpresponse.HTTPError {
	if m.ChangePasswordFunc != nil {
		return m.ChangePasswordFunc(email, oldPassword, newPassword)
	}
	return nil
}

func (m *MockUseCase) RemoveUser(email string) *httpresponse.HTTPError {
	if m.RemoveUserFunc != nil {
		return m.RemoveUserFunc(email)
	}
	return nil
}

// Helper function to create test request
func createRemoveUserRequest(email string) *httptest.Request {
	reqBody := `{"email": "` + email + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/removeuser", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req
}

// ============================================================================
// RemoveUser Tests
// ============================================================================

// TestHandler_RemoveUser_ShouldReturnSuccessWhenUserExists tests successful user removal
func TestHandler_RemoveUser_ShouldReturnSuccessWhenUserExists(t *testing.T) {
	// Arrange
	mockUseCase := &MockUseCase{}
	mockUseCase.RemoveUserFunc = func(email string) *httpresponse.HTTPError {
		if email == "user@example.com" {
			return nil
		}
		return &httpresponse.HTTPError{Code: http.StatusBadRequest, Message: "User not found"}
	}

	handler := NewHandler(mockUseCase)
	e := echo.New()
	req := createRemoveUserRequest("user@example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.RemoveUser(c)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response httpresponse.ResponseHandler
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Status != http.StatusOK {
		t.Errorf("Expected response status %d, got %d", http.StatusOK, response.Status)
	}
}

// TestHandler_RemoveUser_ShouldReturnErrorWhenUserNotFound tests removal of non-existent user
func TestHandler_RemoveUser_ShouldReturnErrorWhenUserNotFound(t *testing.T) {
	// Arrange
	mockUseCase := &MockUseCase{}
	mockUseCase.RemoveUserFunc = func(email string) *httpresponse.HTTPError {
		return &httpresponse.HTTPError{Code: http.StatusBadRequest, Message: "User not found"}
	}

	handler := NewHandler(mockUseCase)
	e := echo.New()
	req := createRemoveUserRequest("nonexistent@example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.RemoveUser(c)

	// Assert
	if err != nil {
		t.Errorf("Expected no error (error is returned via response), got %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response httpresponse.ResponseHandler
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Status != http.StatusBadRequest {
		t.Errorf("Expected response status %d, got %d", http.StatusBadRequest, response.Status)
	}
}

// TestHandler_RemoveUser_ShouldReturnErrorWhenInternalServerError tests internal server error
func TestHandler_RemoveUser_ShouldReturnErrorWhenInternalServerError(t *testing.T) {
	// Arrange
	mockUseCase := &MockUseCase{}
	mockUseCase.RemoveUserFunc = func(email string) *httpresponse.HTTPError {
		return &httpresponse.HTTPError{Code: http.StatusInternalServerError, Message: "Internal server error"}
	}

	handler := NewHandler(mockUseCase)
	e := echo.New()
	req := createRemoveUserRequest("user@example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.RemoveUser(c)

	// Assert
	if err != nil {
		t.Errorf("Expected no error (error is returned via response), got %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// TestHandler_RemoveUser_ShouldReturnErrorWhenEmailEmpty tests empty email request
func TestHandler_RemoveUser_ShouldReturnErrorWhenEmailEmpty(t *testing.T) {
	// Arrange
	mockUseCase := &MockUseCase{}
	// When email is empty, the use case should be called with empty string
	mockUseCase.RemoveUserFunc = func(email string) *httpresponse.HTTPError {
		if email == "" {
			return &httpresponse.HTTPError{Code: http.StatusBadRequest, Message: "Email is required"}
		}
		return nil
	}

	handler := NewHandler(mockUseCase)
	e := echo.New()
	reqBody := `{"email": ""}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/removeuser", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.RemoveUser(c)

	// Assert
	if err != nil {
		t.Errorf("Expected no error (error is returned via response), got %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// TestHandler_RemoveUser_ShouldReturnErrorWhenInvalidJSON tests malformed JSON request
func TestHandler_RemoveUser_ShouldReturnErrorWhenInvalidJSON(t *testing.T) {
	// Arrange
	mockUseCase := &MockUseCase{}
	handler := NewHandler(mockUseCase)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/user/removeuser", strings.NewReader(`{invalid json}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.RemoveUser(c)

	// Assert
	// Error should be returned because JSON is invalid
	if err == nil && rec.Code == http.StatusOK {
		t.Error("Expected error or bad request for invalid JSON")
	}
}

// TestHandler_RemoveUser_ShouldReturnSuccessForDifferentEmailFormats tests various email formats
func TestHandler_RemoveUser_ShouldReturnSuccessForDifferentEmailFormats(t *testing.T) {
	// Arrange
	emails := []string{"user@example.com", "test.user@domain.org", "admin@company.co.id"}

	for _, email := range emails {
		mockUseCase := &MockUseCase{}
		mockUseCase.RemoveUserFunc = func(e string) *httpresponse.HTTPError {
			if e == email {
				return nil
			}
			return &httpresponse.HTTPError{Code: http.StatusBadRequest, Message: "User not found"}
		}

		handler := NewHandler(mockUseCase)
		e := echo.New()
		req := createRemoveUserRequest(email)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Act
		err := handler.RemoveUser(c)

		// Assert
		if err != nil {
			t.Errorf("For email %s: Expected no error, got %v", email, err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("For email %s: Expected status %d, got %d", email, http.StatusOK, rec.Code)
		}
	}
}

// TestHandler_RemoveUser_ShouldCallUseCaseWithCorrectEmail tests that handler passes correct email to use case
func TestHandler_RemoveUser_ShouldCallUseCaseWithCorrectEmail(t *testing.T) {
	// Arrange
	expectedEmail := "specific.user@test.org"
	actualEmail := ""

	mockUseCase := &MockUseCase{}
	mockUseCase.RemoveUserFunc = func(email string) *httpresponse.HTTPError {
		actualEmail = email
		return nil
	}

	handler := NewHandler(mockUseCase)
	e := echo.New()
	req := createRemoveUserRequest(expectedEmail)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	_ = handler.RemoveUser(c)

	// Assert
	if actualEmail != expectedEmail {
		t.Errorf("Expected use case to be called with email '%s', got '%s'", expectedEmail, actualEmail)
	}
}
