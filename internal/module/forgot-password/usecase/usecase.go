package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"sensor-service/internal/module/forgot-password/repository"
	"sensor-service/internal/module/auth/entity"
	"sensor-service/internal/platform/httpengine/httpresponse"
	"sync"
	"time"
)

// PasswordResetToken stores reset tokens with expiry info - O(1) lookup
type PasswordResetToken struct {
	Token     string
	Email     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// ResetTokenCache stores active reset tokens with expiry - O(1) lookup
type ResetTokenCache struct {
	mu     sync.RWMutex
	tokens map[string]*PasswordResetToken
}

type IUseCase interface {
	InitiateForgotPassword(email string) (resetToken string, err *httpresponse.HTTPError)
	ValidateResetToken(resetToken string) (email string, err *httpresponse.HTTPError)
	CompleteForgotPassword(resetToken string, newPassword string) *httpresponse.HTTPError
}

type UseCase struct {
	GenericRepository repository.IGenericRepository
	resetTokenCache   *ResetTokenCache
}

// InitiateForgotPassword generates secure reset token for password recovery
// O(n) DB query where n = 1 (indexed email) + O(1) token generation + O(1) cache insertion
func (u UseCase) InitiateForgotPassword(email string) (string, *httpresponse.HTTPError) {
	httpError := httpresponse.HTTPError{}

	// Validate email input - O(1)
	if email == "" {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Email cannot be empty"
		return "", &httpError
	}

	// Verify email exists in DB - O(n) where n = 1 (indexed email lookup)
	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, email)
	if err != nil {
		log.Println("{InitiateForgotPassword}{FindByEmail}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return "", &httpError
	}

	if authUser == nil {
		// Return generic message for security (don't reveal if email exists)
		httpError.Code = http.StatusBadRequest
		httpError.Message = "If email exists, reset link will be sent"
		return "", &httpError
	}

	// Generate cryptographically secure random token - O(1) operation
	resetToken := u.generateSecureToken(32)

	// Store reset token with 15min expiry - O(1) insertion
	u.resetTokenCache.mu.Lock()
	u.resetTokenCache.tokens[resetToken] = &PasswordResetToken{
		Token:     resetToken,
		Email:     email,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
	}
	u.resetTokenCache.mu.Unlock()

	return resetToken, nil
}

// ValidateResetToken checks if reset token is valid and not expired
// O(1) token cache lookup
func (u UseCase) ValidateResetToken(resetToken string) (string, *httpresponse.HTTPError) {
	httpError := httpresponse.HTTPError{}

	// Validate token input - O(1)
	if resetToken == "" {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Reset token cannot be empty"
		return "", &httpError
	}

	// O(1) token cache lookup
	u.resetTokenCache.mu.RLock()
	pwdReset, exists := u.resetTokenCache.tokens[resetToken]
	u.resetTokenCache.mu.RUnlock()

	if !exists {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Invalid reset token"
		return "", &httpError
	}

	// Check token expiry - O(1) time comparison
	if time.Now().After(pwdReset.ExpiresAt) {
		// O(1) token deletion
		u.resetTokenCache.mu.Lock()
		delete(u.resetTokenCache.tokens, resetToken)
		u.resetTokenCache.mu.Unlock()

		httpError.Code = http.StatusBadRequest
		httpError.Message = "Reset token expired"
		return "", &httpError
	}

	return pwdReset.Email, nil
}

// CompleteForgotPassword validates reset token and updates user password
// O(1) token lookup + O(n) DB update where n = 1 (indexed email)
func (u UseCase) CompleteForgotPassword(resetToken string, newPassword string) *httpresponse.HTTPError {
	httpError := httpresponse.HTTPError{}

	// Validate inputs - O(1)
	if resetToken == "" {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Reset token cannot be empty"
		return &httpError
	}

	if newPassword == "" {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "New password cannot be empty"
		return &httpError
	}

	// O(1) token cache lookup
	u.resetTokenCache.mu.RLock()
	pwdReset, exists := u.resetTokenCache.tokens[resetToken]
	u.resetTokenCache.mu.RUnlock()

	if !exists {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Invalid reset token"
		return &httpError
	}

	// Check token expiry - O(1) time comparison
	if time.Now().After(pwdReset.ExpiresAt) {
		// O(1) token deletion
		u.resetTokenCache.mu.Lock()
		delete(u.resetTokenCache.tokens, resetToken)
		u.resetTokenCache.mu.Unlock()

		httpError.Code = http.StatusBadRequest
		httpError.Message = "Reset token expired"
		return &httpError
	}

	// Retrieve user from DB - O(n) where n = 1 (indexed email lookup)
	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, pwdReset.Email)
	if err != nil {
		log.Println("{CompleteForgotPassword}{FindByEmail}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return &httpError
	}

	if authUser == nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = "User not found"
		return &httpError
	}

	// Update user password in DB - O(n) where n = 1 (indexed email update)
	if err := u.GenericRepository.UpdatePassword(pwdReset.Email, newPassword); err != nil {
		log.Println("{CompleteForgotPassword}{UpdatePassword}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = "Failed to update password"
		return &httpError
	}

	// Invalidate reset token - O(1) deletion
	u.resetTokenCache.mu.Lock()
	delete(u.resetTokenCache.tokens, resetToken)
	u.resetTokenCache.mu.Unlock()

	return nil
}

// generateSecureToken creates cryptographically secure random token - O(1) operation
func (u UseCase) generateSecureToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Println("Error generating secure token:", err)
		return ""
	}
	return hex.EncodeToString(bytes)
}

func NewUseCase(genericRepository repository.IGenericRepository) IUseCase {
	return UseCase{
		GenericRepository: genericRepository,
		resetTokenCache: &ResetTokenCache{
			tokens: make(map[string]*PasswordResetToken),
		},
	}
}
