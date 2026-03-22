package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sensor-service/internal/module/user/dto"
	userUsecase "sensor-service/internal/module/user/usecase"
	"sensor-service/internal/platform/helper"
)

// MockUserUseCase implements userUsecase.IUserUseCase for testing
type MockUserUseCase struct {
	GetUserByEmailFunc    func(email string) (*userEntity.User, *httpresponse.HTTPError)
	UpdateUserFunc        func(email string, request config.HTTPRequest, callerRole string) (*userEntity.User, *httpresponse.HTTPError)
	UpdateUserStatusFunc  func(email string, callerRole string, newStatus string) *httpresponse.HTTPError
	GetAllUsersFunc       func() ([]userEntity.User, *httpresponse.HTTPError)
}

func (m *MockUserUseCase) GetUserByEmail(email string) (*userEntity.User, *httpresponse.HTTPError) {
	return m.GetUserByEmailFunc(email)
}

func (m *MockUserUseCase) UpdateUser(email string, request config.HTTPRequest, callerRole string) (*userEntity.User, *httpresponse.HTTPError) {
	return m.UpdateUserFunc(email, request, callerRole)
}

func (m *MockUserUseCase) UpdateUserStatus(email string, callerRole string, newStatus string) *httpresponse.HTTPError {
	return m.UpdateUserStatusFunc(email, callerRole, newStatus)
}

func (m *MockUserUseCase) GetAllUsers() ([]userEntity.User, *httpresponse.HTTPError) {
	return m.GetAllUsersFunc()
}

// Test helper to create request
func createRequest(method, path string, body map[string]interface{}) *http.Request {
	var bodyStr string
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyStr = string(bodyBytes)
	}
	req := httptest.NewRequest(method, path, strings.NewReader(bodyStr))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// Test helper to create response recorder
func createRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// ===================== GetUserByEmail Tests =====================

func TestGetUserByEmail_HappyPath(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		GetUserByEmailFunc: func(email string) (*userEntity.User, *httpresponse.HTTPError) {
			return &userEntity.User{
				ID:       1,
				Email:    "test@example.com",
				Role:     "user",
				Status:   "activated",
				Password: "hashedpassword",
			}, nil
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	req := createRequest("GET", "/api/user/test@example.com", nil)
	rec := createRecorder()

	handler.GetUserByEmail(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["message"] != "User retrieved successfully" {
		t.Errorf("Unexpected message: %v", response["message"])
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		GetUserByEmailFunc: func(email string) (*userEntity.User, *httpresponse.HTTPError) {
			return nil, &httpresponse.HTTPError{
				Code:    http.StatusNotFound,
				Message: "User not found",
			}
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	req := createRequest("GET", "/api/user/nonexistent@example.com", nil)
	rec := createRecorder()

	handler.GetUserByEmail(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetUserByEmail_InternalError(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		GetUserByEmailFunc: func(email string) (*userEntity.User, *httpresponse.HTTPError) {
			return nil, &httpresponse.HTTPError{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
			}
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	req := createRequest("GET", "/api/user/test@example.com", nil)
	rec := createRecorder()

	handler.GetUserByEmail(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// ===================== UpdateUser Tests =====================

func TestUpdateUser_HappyPath(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		UpdateUserFunc: func(email string, request config.HTTPRequest, callerRole string) (*userEntity.User, *httpresponse.HTTPError) {
			return &userEntity.User{
				ID:     1,
				Email:  "test@example.com",
				Role:   "admin",
				Status: "activated",
			}, nil
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	reqBody := map[string]interface{}{
		"role": "admin",
	}
	req := createRequest("PUT", "/api/user/test@example.com", reqBody)
	// Note: In real scenario, JWT token would be set in header

	rec := createRecorder()

	handler.UpdateUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["message"] != "User updated successfully" {
		t.Errorf("Unexpected message: %v", response["message"])
	}
}

func TestUpdateUser_Forbidden(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		UpdateUserFunc: func(email string, request config.HTTPRequest, callerRole string) (*userEntity.User, *httpresponse.HTTPError) {
			return nil, &httpresponse.HTTPError{
				Code:    http.StatusForbidden,
				Message: "Forbidden: only admin can update user data",
			}
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	reqBody := map[string]interface{}{
		"role": "user",
	}
	req := createRequest("PUT", "/api/user/test@example.com", reqBody)
	rec := createRecorder()

	handler.UpdateUser(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		UpdateUserFunc: func(email string, request config.HTTPRequest, callerRole string) (*userEntity.User, *httpresponse.HTTPError) {
			return nil, &httpresponse.HTTPError{
				Code:    http.StatusNotFound,
				Message: "User not found",
			}
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	reqBody := map[string]interface{}{
		"role": "admin",
	}
	req := createRequest("PUT", "/api/user/nonexistent@example.com", reqBody)
	rec := createRecorder()

	handler.UpdateUser(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUpdateUser_InvalidRole(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		UpdateUserFunc: func(email string, request config.HTTPRequest, callerRole string) (*userEntity.User, *httpresponse.HTTPError) {
			return nil, &httpresponse.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "Invalid role: must be 'admin' or 'user'",
			}
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	reqBody := map[string]interface{}{
		"role": "invalidrole",
	}
	req := createRequest("PUT", "/api/user/test@example.com", reqBody)
	rec := createRecorder()

	handler.UpdateUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// ===================== UpdateUserStatus Tests =====================

func TestUpdateUserStatus_HappyPath(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		UpdateUserStatusFunc: func(email string, callerRole string, newStatus string) *httpresponse.HTTPError {
			return nil
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	reqBody := map[string]interface{}{
		"status": "deactivated",
	}
	req := createRequest("PATCH", "/api/user/test@example.com/status", reqBody)
	rec := createRecorder()

	handler.UpdateUserStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["message"] != "User status updated successfully" {
		t.Errorf("Unexpected message: %v", response["message"])
	}
}

func TestUpdateUserStatus_Forbidden(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		UpdateUserStatusFunc: func(email string, callerRole string, newStatus string) *httpresponse.HTTPError {
			return &httpresponse.HTTPError{
				Code:    http.StatusForbidden,
				Message: "Forbidden: only admin can change user status",
			}
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	reqBody := map[string]interface{}{
		"status": "deactivated",
	}
	req := createRequest("PATCH", "/api/user/test@example.com/status", reqBody)
	rec := createRecorder()

	handler.UpdateUserStatus(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestUpdateUserStatus_InvalidStatus(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		UpdateUserStatusFunc: func(email string, callerRole string, newStatus string) *httpresponse.HTTPError {
			return &httpresponse.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "Invalid status: must be 'activated' or 'deactivated'",
			}
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	reqBody := map[string]interface{}{
		"status": "invalidstatus",
	}
	req := createRequest("PATCH", "/api/user/test@example.com/status", reqBody)
	rec := createRecorder()

	handler.UpdateUserStatus(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateUserStatus_NotFound(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		UpdateUserStatusFunc: func(email string, callerRole string, newStatus string) *httpresponse.HTTPError {
			return &httpresponse.HTTPError{
				Code:    http.StatusNotFound,
				Message: "User not found",
			}
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	reqBody := map[string]interface{}{
		"status": "deactivated",
	}
	req := createRequest("PATCH", "/api/user/nonexistent@example.com/status", reqBody)
	rec := createRecorder()

	handler.UpdateUserStatus(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

// ===================== GetAllUsers Tests =====================

func TestGetAllUsers_HappyPath(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		GetAllUsersFunc: func() ([]userEntity.User, *httpresponse.HTTPError) {
			return []userEntity.User{
				{ID: 1, Email: "user1@example.com", Role: "user", Status: "activated"},
				{ID: 2, Email: "user2@example.com", Role: "admin", Status: "activated"},
			}, nil
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	req := createRequest("GET", "/api/users", nil)
	rec := createRecorder()

	handler.GetAllUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["message"] != "Users retrieved successfully" {
		t.Errorf("Unexpected message: %v", response["message"])
	}
}

func TestGetAllUsers_EmptyList(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		GetAllUsersFunc: func() ([]userEntity.User, *httpresponse.HTTPError) {
			return []userEntity.User{}, nil
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	req := createRequest("GET", "/api/users", nil)
	rec := createRecorder()

	handler.GetAllUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetAllUsers_InternalError(t *testing.T) {
	mockUseCase := &MockUserUseCase{
		GetAllUsersFunc: func() ([]userEntity.User, *httpresponse.HTTPError) {
			return nil, &httpresponse.HTTPError{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
			}
		},
	}

	handler := &UserHandler{UserUseCase: mockUseCase}

	req := createRequest("GET", "/api/users", nil)
	rec := createRecorder()

	handler.GetAllUsers(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
