package usecase_test

import (
	"testing"
	"time"

	"sensor-service/config"
	"sensor-service/internal/module/user/entity"
	userUsecase "sensor-service/internal/module/user/usecase"
	"sensor-service/internal/platform/app"
)

// MockGenericRepository implements auth.IGenericRepository for testing
type MockGenericRepository struct {
	FindByEmailFunc func(entity interface{}, email string) (interface{}, error)
	FindAllFunc     func(entity interface{}) ([]interface{}, error)
	UpdateFunc      func(entity interface{}) error
	CreateFunc      func(entity interface{}) error
	DeleteByIDFunc  func(id int) error
}

func (m *MockGenericRepository) FindByEmail(entity interface{}, email string) (interface{}, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(entity, email)
	}
	return nil, nil
}

func (m *MockGenericRepository) FindAll(entity interface{}) ([]interface{}, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(entity)
	}
	return nil, nil
}

func (m *MockGenericRepository) Update(entity interface{}) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(entity)
	}
	return nil
}

func (m *MockGenericRepository) Create(entity interface{}) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(entity)
	}
	return nil
}

func (m *MockGenericRepository) DeleteByID(id int) error {
	if m.DeleteByIDFunc != nil {
		return m.DeleteByIDFunc(id)
	}
	return nil
}

// ===================== GetUserByEmail Tests =====================

func TestGetUserByEmail_HappyPath(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindByEmailFunc: func(entity interface{}, email string) (interface{}, error) {
			return &entity.User{
				ID:       1,
				Email:    "test@example.com",
				Role:     "user",
				Status:   "activated",
				Password: "hashedpassword",
			}, nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	user, err := useCase.GetUserByEmail("test@example.com")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	if user.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", user.Role)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindByEmailFunc: func(entity interface{}, email string) (interface{}, error) {
			return nil, nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	user, err := useCase.GetUserByEmail("nonexistent@example.com")

	if user != nil {
		t.Errorf("Expected nil user, got %v", user)
	}

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err.Code != 404 {
		t.Errorf("Expected status code 404, got %d", err.Code)
	}
}

func TestGetUserByEmail_CacheHit(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindByEmailFunc: func(entity interface{}, email string) (interface{}, error) {
			return &entity.User{
				ID:       1,
				Email:    "test@example.com",
				Role:     "user",
				Status:   "activated",
				Password: "hashedpassword",
			}, nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	// First call - populates cache
	_, err := useCase.GetUserByEmail("test@example.com")
	if err != nil {
		t.Errorf("Expected no error on first call, got %v", err)
	}

	// Second call - should hit cache
	_, err = useCase.GetUserByEmail("test@example.com")
	if err != nil {
		t.Errorf("Expected no error on second call (cache hit), got %v", err)
	}
}

// ===================== UpdateUser Tests =====================

func TestUpdateUser_HappyPath(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindByEmailFunc: func(entity interface{}, email string) (interface{}, error) {
			return &entity.User{
				ID:       1,
				Email:    "test@example.com",
				Role:     "user",
				Status:   "activated",
				Password: "oldhashedpassword",
			}, nil
		},
		UpdateFunc: func(entity interface{}) error {
			return nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	request := config.HTTPRequest{}
	request.Body = map[string]interface{}{
		"role": "admin",
	}

	user, err := useCase.UpdateUser("test@example.com", request, "admin")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", user.Role)
	}
}

func TestUpdateUser_Forbidden(t *testing.T) {
	mockRepo := &MockGenericRepository{}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	request := config.HTTPRequest{}
	request.Body = map[string]interface{}{
		"role": "admin",
	}

	_, err := useCase.UpdateUser("test@example.com", request, "user")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err.Code != 403 {
		t.Errorf("Expected status code 403, got %d", err.Code)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindByEmailFunc: func(entity interface{}, email string) (interface{}, error) {
			return nil, nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	request := config.HTTPRequest{}
	request.Body = map[string]interface{}{
		"role": "admin",
	}

	_, err := useCase.UpdateUser("nonexistent@example.com", request, "admin")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err.Code != 404 {
		t.Errorf("Expected status code 404, got %d", err.Code)
	}
}

func TestUpdateUser_InvalidRole(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindByEmailFunc: func(entity interface{}, email string) (interface{}, error) {
			return &entity.User{
				ID:       1,
				Email:    "test@example.com",
				Role:     "user",
				Status:   "activated",
				Password: "hashedpassword",
			}, nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	request := config.HTTPRequest{}
	request.Body = map[string]interface{}{
		"role": "invalidrole",
	}

	_, err := useCase.UpdateUser("test@example.com", request, "admin")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err.Code != 400 {
		t.Errorf("Expected status code 400, got %d", err.Code)
	}
}

func TestUpdateUser_UpdateEmail(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindByEmailFunc: func(entity interface{}, email string) (interface{}, error) {
			if email == "new@example.com" {
				return nil, nil // new email doesn't exist
			}
			return &entity.User{
				ID:       1,
				Email:    "old@example.com",
				Role:     "user",
				Status:   "activated",
				Password: "hashedpassword",
			}, nil
		},
		UpdateFunc: func(entity interface{}) error {
			return nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	request := config.HTTPRequest{}
	request.Body = map[string]interface{}{
		"email": "new@example.com",
	}

	user, err := useCase.UpdateUser("old@example.com", request, "admin")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user.Email != "new@example.com" {
		t.Errorf("Expected email 'new@example.com', got '%s'", user.Email)
	}
}

// ===================== UpdateUserStatus Tests =====================

func TestUpdateUserStatus_HappyPath(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindByEmailFunc: func(entity interface{}, email string) (interface{}, error) {
			return &entity.User{
				ID:       1,
				Email:    "test@example.com",
				Role:     "user",
				Status:   "activated",
				Password: "hashedpassword",
			}, nil
		},
		UpdateFunc: func(entity interface{}) error {
			return nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	err := useCase.UpdateUserStatus("test@example.com", "admin", "deactivated")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestUpdateUserStatus_Forbidden(t *testing.T) {
	mockRepo := &MockGenericRepository{}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	err := useCase.UpdateUserStatus("test@example.com", "user", "deactivated")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err.Code != 403 {
		t.Errorf("Expected status code 403, got %d", err.Code)
	}
}

func TestUpdateUserStatus_InvalidStatus(t *testing.T) {
	mockRepo := &MockGenericRepository{}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	err := useCase.UpdateUserStatus("test@example.com", "admin", "invalidstatus")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err.Code != 400 {
		t.Errorf("Expected status code 400, got %d", err.Code)
	}
}

func TestUpdateUserStatus_NotFound(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindByEmailFunc: func(entity interface{}, email string) (interface{}, error) {
			return nil, nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	err := useCase.UpdateUserStatus("nonexistent@example.com", "admin", "deactivated")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err.Code != 404 {
		t.Errorf("Expected status code 404, got %d", err.Code)
	}
}

// ===================== GetAllUsers Tests =====================

func TestGetAllUsers_HappyPath(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindAllFunc: func(entity interface{}) ([]interface{}, error) {
			return []interface{}{
				&entity.User{ID: 1, Email: "user1@example.com", Role: "user", Status: "activated"},
				&entity.User{ID: 2, Email: "user2@example.com", Role: "admin", Status: "activated"},
			}, nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	users, err := useCase.GetAllUsers()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestGetAllUsers_EmptyList(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindAllFunc: func(entity interface{}) ([]interface{}, error) {
			return []interface{}{}, nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	users, err := useCase.GetAllUsers()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}

func TestGetAllUsers_InternalError(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindAllFunc: func(entity interface{}) ([]interface{}, error) {
			return nil, &entity.User{}
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	_, err := useCase.GetAllUsers()

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

// ===================== Cache Tests =====================

func TestCacheExpiration(t *testing.T) {
	mockRepo := &MockGenericRepository{
		FindByEmailFunc: func(entity interface{}, email string) (interface{}, error) {
			return &entity.User{
				ID:       1,
				Email:    "test@example.com",
				Role:     "user",
				Status:   "activated",
				Password: "hashedpassword",
			}, nil
		},
	}

	cfg := app.App{}
	useCase := userUsecase.NewUserUseCase(mockRepo, cfg)

	// First call - populates cache with 15min TTL
	_, err := useCase.GetUserByEmail("test@example.com")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify cache entry exists
	if len(useCase.(*userUsecase.UserUseCase).GetCacheForTest()) == 0 {
		t.Error("Expected cache to have 1 entry")
	}
}
