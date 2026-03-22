package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/exp/slices"
	"sensor-service/config"
	"sensor-service/internal/module/auth/dto"
	"sensor-service/internal/module/auth/entity"
	auth "sensor-service/internal/module/auth/repository"
	"sensor-service/internal/module/user/dto"
	userEntity "sensor-service/internal/module/user/entity"
	"sensor-service/internal/platform/app"
	"sensor-service/internal/platform/helper"
	"sensor-service/internal/platform/httpengine/httpresponse"
	"sync"
	"time"
)

// UserCache stores cached user data with expiry - O(1) lookup
type UserCache struct {
	mu    sync.RWMutex
	users map[string]*UserCacheMetadata
}

// UserCacheMetadata holds cached user info with TTL
type UserCacheMetadata struct {
	UserID    int
	Email     string
	Role      string
	Status    string
	ExpiresAt time.Time
	CachedAt  time.Time
}

type IUserUseCase interface {
	GetUserByEmail(email string) (*userEntity.User, *httpresponse.HTTPError)
	UpdateUser(email string, request config.HTTPRequest, callerRole string) (*userEntity.User, *httpresponse.HTTPError)
	UpdateUserStatus(email string, callerRole string, newStatus string) *httpresponse.HTTPError
	GetAllUsers() ([]userEntity.User, *httpresponse.HTTPError)
}

type UserUseCase struct {
	GenericRepository auth.IGenericRepository
	AppCfg            app.App
	userCache         *UserCache
	mu                sync.RWMutex
}

// NewUserUseCase creates a new UserUseCase instance
// O(1) initialization
func NewUserUseCase(repo auth.IGenericRepository, cfg app.App) *UserUseCase {
	return &UserUseCase{
		GenericRepository: repo,
		AppCfg:            cfg,
		userCache: &UserCache{
			users: make(map[string]*UserCacheMetadata),
		},
	}
}

// GetUserByEmail retrieves user by email with caching
// O(1) cache lookup + O(n) DB query where n = 1 (indexed email)
func (u *UserUseCase) GetUserByEmail(email string) (*userEntity.User, *httpresponse.HTTPError) {
	httpError := httpresponse.HTTPError{}

	// Check cache first - O(1) lookup
	u.userCache.mu.RLock()
	if cached, exists := u.userCache.users[email]; exists && time.Now().Before(cached.ExpiresAt) {
		u.userCache.mu.RUnlock()
		// Fetch from DB to get full user data
		authUser, err := u.GenericRepository.FindByEmail(entity.User{}, email)
		if err != nil || authUser == nil {
			httpError.Code = http.StatusInternalServerError
			httpError.Message = httpresponse.ErrorInternalServerError.Message
			return nil, &httpError
		}
		user, _ := helper.TypeConverter[userEntity.User](&authUser)
		return user, nil
	}
	u.userCache.mu.RUnlock()

	// O(n) DB query where n = 1 (indexed email)
	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, email)
	if err != nil {
		log.Println("{GetUserByEmail}{FindByEmail}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return nil, &httpError
	}

	if authUser == nil {
		httpError.Code = http.StatusNotFound
		httpError.Message = "User not found"
		return nil, &httpError
	}

	user, _ := helper.TypeConverter[userEntity.User](&authUser)

	// Cache user metadata - O(1) insertion
	u.cacheUser(user)

	return user, nil
}

// UpdateUser updates user data (email, role, status) - restricted to admin role only
// O(1) authorization check + O(n) DB query + O(n) DB update where n = 1 (indexed email)
func (u *UserUseCase) UpdateUser(email string, request config.HTTPRequest, callerRole string) (*userEntity.User, *httpresponse.HTTPError) {
	httpError := httpresponse.HTTPError{}

	// Authorization check: only admin role can update users - O(1) check
	if callerRole != "admin" {
		httpError.Code = http.StatusForbidden
		httpError.Message = "Forbidden: only admin can update user data"
		return nil, &httpError
	}

	var updateReq dto.UpdateUserRequest
	config := helper.DecoderConfig(&updateReq)
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return nil, &httpError
	}
	if err = decoder.Decode(request.Body); err != nil {
		log.Println("{UpdateUser}{Decode}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return nil, &httpError
	}

	// Find user by email - O(n) DB query where n = 1 (indexed email)
	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, email)
	if err != nil {
		log.Println("{UpdateUser}{FindByEmail}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return nil, &httpError
	}

	if authUser == nil {
		httpError.Code = http.StatusNotFound
		httpError.Message = "User not found"
		return nil, &httpError
	}

	user, _ := helper.TypeConverter[userEntity.User](&authUser)

	// Update fields if provided
	if updateReq.Email != "" {
		// Check if new email already exists - O(n) DB query
		existingUser, err := u.GenericRepository.FindByEmail(entity.User{}, updateReq.Email)
		if err != nil {
			log.Println("{UpdateUser}{FindByEmail}{Error} : ", err)
			httpError.Code = http.StatusInternalServerError
			httpError.Message = httpresponse.ErrorInternalServerError.Message
			return nil, &httpError
		}
		if existingUser != nil && updateReq.Email != email {
			httpError.Code = http.StatusBadRequest
			httpError.Message = "Email already exists"
			return nil, &httpError
		}
		user.Email = updateReq.Email
	}

	if updateReq.Role != "" {
		// Validate role - O(1) check
		if !slices.Contains(helper.Privileges, updateReq.Role) {
			httpError.Code = http.StatusBadRequest
			httpError.Message = "Invalid role: must be 'admin' or 'user'"
			return nil, &httpError
		}
		user.Role = updateReq.Role
	}

	if updateReq.Status != "" {
		// Validate status - O(1) check
		if updateReq.Status != "activated" && updateReq.Status != "deactivated" {
			httpError.Code = http.StatusBadRequest
			httpError.Message = "Invalid status: must be 'activated' or 'deactivated'"
			return nil, &httpError
		}
		user.Status = updateReq.Status
	}

	if updateReq.Password != "" {
		// Hash new password - O(1) operation
		hashedPassword, err := helper.GeneratehashPassword(updateReq.Password)
		if err != nil {
			log.Println("{UpdateUser}{GeneratehashPassword}{Error} : ", err)
			httpError.Code = http.StatusInternalServerError
			httpError.Message = "Failed to hash password"
			return nil, &httpError
		}
		user.Password = hashedPassword
	}

	// Update user in DB - O(n) DB update where n = 1 (indexed email)
	if err := u.GenericRepository.Update(user); err != nil {
		log.Println("{UpdateUser}{Update}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return nil, &httpError
	}

	// Invalidate cache - O(1) deletion
	u.userCache.mu.Lock()
	delete(u.userCache.users, email)
	u.userCache.mu.Unlock()

	return user, nil
}

// UpdateUserStatus activates or deactivates a user account - restricted to admin role only
// O(1) authorization check + O(n) DB query + O(n) DB update where n = 1 (indexed email)
func (u *UserUseCase) UpdateUserStatus(email string, callerRole string, newStatus string) *httpresponse.HTTPError {
	httpError := httpresponse.HTTPError{}

	// Authorization check: only admin role can change user status - O(1) check
	if callerRole != "admin" {
		httpError.Code = http.StatusForbidden
		httpError.Message = "Forbidden: only admin can change user status"
		return &httpError
	}

	// Validate status value - O(1) check
	if newStatus != "activated" && newStatus != "deactivated" {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Invalid status: must be 'activated' or 'deactivated'"
		return &httpError
	}

	// Find user by email - O(n) DB query where n = 1 (indexed email)
	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, email)
	if err != nil {
		log.Println("{UpdateUserStatus}{FindByEmail}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return &httpError
	}

	if authUser == nil {
		httpError.Code = http.StatusNotFound
		httpError.Message = "User not found"
		return &httpError
	}

	user, _ := helper.TypeConverter[userEntity.User](&authUser)

	// Update user status - O(n) DB update where n = 1 (indexed email)
	user.Status = newStatus

	if err := u.GenericRepository.Update(user); err != nil {
		log.Println("{UpdateUserStatus}{Update}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return &httpError
	}

	// Invalidate cache - O(1) deletion
	u.userCache.mu.Lock()
	delete(u.userCache.users, email)
	u.userCache.mu.Unlock()

	return nil
}

// GetAllUsers retrieves all users
// O(n) DB query where n = number of users
func (u *UserUseCase) GetAllUsers() ([]userEntity.User, *httpresponse.HTTPError) {
	httpError := httpresponse.HTTPError{}

	// O(n) DB query
	users, err := u.GenericRepository.FindAll(entity.User{})
	if err != nil {
		log.Println("{GetAllUsers}{FindAll}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return nil, &httpError
	}

	var result []userEntity.User
	for _, u := range users {
		user, _ := helper.TypeConverter[userEntity.User](&u)
		result = append(result, *user)
	}

	return result, nil
}

// cacheUser stores user metadata in cache with TTL
// O(1) insertion
func (u *UserUseCase) cacheUser(user *userEntity.User) {
	u.userCache.mu.Lock()
	u.userCache.users[user.Email] = &UserCacheMetadata{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CachedAt:  time.Now(),
	}
	u.userCache.mu.Unlock()
}

// generateSecureToken generates a cryptographically secure random token
// O(1) operation
func (u *UserUseCase) generateSecureToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// IUserRepository interface for user-specific repository operations
type IUserRepository interface {
	FindByEmail(email string) (*userEntity.User, error)
	FindAll() ([]userEntity.User, error)
	Update(user *userEntity.User) error
}
