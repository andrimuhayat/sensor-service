package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/exp/slices"
	"log"
	"net/http"
	"sensor-service/config"
	"sensor-service/internal/module/auth/dto"
	"sensor-service/internal/module/auth/entity"
	auth "sensor-service/internal/module/auth/repository"
	"sensor-service/internal/platform/app"
	"sensor-service/internal/platform/helper"
	"sensor-service/internal/platform/httpengine/httpresponse"
	"sync"
	"time"
)

// TokenCache stores active tokens with expiry info - O(1) lookup
type TokenCache struct {
	mu     sync.RWMutex
	tokens map[string]*TokenMetadata
}

// TokenMetadata holds token expiry and user info
type TokenMetadata struct {
	Email     string
	Role      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// SessionStore tracks active sessions with rate limiting - O(1) lookup
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*SessionData
}

// SessionData holds session info and rate limit counters
type SessionData struct {
	UserEmail      string
	LoginAttempts  int
	LastAttemptAt  time.Time
	SessionExpiry  time.Time
	RefreshTokens  map[string]time.Time // O(1) refresh token lookup
}

// PasswordResetToken stores reset tokens with expiry
type PasswordResetToken struct {
	Token     string
	Email     string
	ExpiresAt time.Time
}

type IUseCase interface {
	SignIn(request config.HTTPRequest) (dto.Token, *httpresponse.HTTPError)
	SignUp(request config.HTTPRequest) (entity.User, *httpresponse.HTTPError)
	RefreshToken(oldToken string) (dto.Token, *httpresponse.HTTPError)
	InitiatePasswordReset(email string) (string, *httpresponse.HTTPError)
	ResetPassword(resetToken string, newPassword string) *httpresponse.HTTPError
	ValidateSession(email string) (bool, *httpresponse.HTTPError)
	CheckRateLimit(email string) (bool, *httpresponse.HTTPError)
	Logout(token string) *httpresponse.HTTPError
	ChangePassword(email string, oldPassword string, newPassword string) *httpresponse.HTTPError
}

type UseCase struct {
	GenericRepository auth.IGenericRepository
	AppCfg            app.App
	tokenCache        *TokenCache
	sessionStore      *SessionStore
	resetTokens       map[string]*PasswordResetToken
	resetTokenMu      sync.RWMutex
}

// SignIn authenticates user and returns JWT token
// O(1) token cache lookup + O(n) DB query where n = 1 (indexed email)
func (u UseCase) SignIn(request config.HTTPRequest) (dto.Token, *httpresponse.HTTPError) {
	var err error
	httpError := httpresponse.HTTPError{}

	var requestAuth dto.Authentication

	config := helper.DecoderConfig(&requestAuth)
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return dto.Token{}, &httpError
	}
	if err = decoder.Decode(request.Body); err != nil {
		log.Println("{SignIn}{Decode}{Error} : ", err)
	}

	// Check rate limit before DB query - O(1) session lookup
	if allowed, rateLimitErr := u.CheckRateLimit(requestAuth.Email); !allowed {
		return dto.Token{}, rateLimitErr
	}

	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, requestAuth.Email)
	if err != nil {
		log.Println("{SignIn}{FindByEmail}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return dto.Token{}, &httpError
	}

	if authUser == nil {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Username or Password is incorrect"
		return dto.Token{}, &httpError
	}

	user, _ := helper.TypeConverter[entity.User](&authUser)

	check := helper.CheckPasswordHash(requestAuth.Password, user.Password)

	if !check {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Username or Password is incorrect"
		return dto.Token{}, &httpError
	}

	validToken, err := helper.GenerateJWT(user.Email, user.Role, u.AppCfg.SecretKey)
	if err != nil {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Failed to generate token"
		return dto.Token{}, &httpError
	}

	// Cache token metadata - O(1) insertion
	u.cacheToken(validToken, user.Email, user.Role)

	// Create session - O(1) insertion
	u.createSession(user.Email)

	var token dto.Token
	token.Email = user.Email
	token.Role = user.Role
	token.TokenString = validToken

	return token, nil
}

// SignUp creates new user account
// O(1) role validation + O(n) DB query where n = 1 (indexed email)
func (u UseCase) SignUp(request config.HTTPRequest) (entity.User, *httpresponse.HTTPError) {
	var err error
	httpError := httpresponse.HTTPError{}

	var user entity.User

	config := helper.DecoderConfig(&user)
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return entity.User{}, &httpError
	}
	if err = decoder.Decode(request.Body); err != nil {
		log.Println("{SignIn}{Decode}{Error} : ", err)
	}

	if !slices.Contains(helper.Privileges, user.Role) {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Role doesn't exists, please use admin or user"
		return entity.User{}, &httpError
	}

	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, user.Email)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return entity.User{}, &httpError
	}

	if authUser != nil {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "email cannot be same"
		return entity.User{}, &httpError
	}
	user.Password, err = helper.GeneratehashPassword(user.Password)
	if err != nil {
		log.Fatalln("Error in password hashing.")
	}

	err = u.GenericRepository.Create(user)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return entity.User{}, &httpError
	}
	return user, nil
}

// RefreshToken generates new token from existing valid token
// O(1) token cache lookup + O(1) token generation
func (u UseCase) RefreshToken(oldToken string) (dto.Token, *httpresponse.HTTPError) {
	httpError := httpresponse.HTTPError{}

	// O(1) cache lookup
	u.tokenCache.mu.RLock()
	metadata, exists := u.tokenCache.tokens[oldToken]
	u.tokenCache.mu.RUnlock()

	if !exists || time.Now().After(metadata.ExpiresAt) {
		httpError.Code = http.StatusUnauthorized
		httpError.Message = "Token expired or invalid"
		return dto.Token{}, &httpError
	}

	// Generate new token - O(1) operation
	newToken, err := helper.GenerateJWT(metadata.Email, metadata.Role, u.AppCfg.SecretKey)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = "Failed to generate new token"
		return dto.Token{}, &httpError
	}

	// Cache new token - O(1) insertion
	u.cacheToken(newToken, metadata.Email, metadata.Role)

	// Invalidate old token - O(1) deletion
	u.tokenCache.mu.Lock()
	delete(u.tokenCache.tokens, oldToken)
	u.tokenCache.mu.Unlock()

	return dto.Token{
		Email:       metadata.Email,
		Role:        metadata.Role,
		TokenString: newToken,
	}, nil
}

// InitiatePasswordReset generates secure reset token
// O(1) token generation + O(1) storage
func (u UseCase) InitiatePasswordReset(email string) (string, *httpresponse.HTTPError) {
	httpError := httpresponse.HTTPError{}

	// Verify email exists - O(n) DB query where n = 1 (indexed email)
	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, email)
	if err != nil || authUser == nil {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Email not found"
		return "", &httpError
	}

	// Generate secure random token - O(1) operation
	resetToken := u.generateSecureToken(32)

	// Store reset token with 15min expiry - O(1) insertion
	u.resetTokenMu.Lock()
	u.resetTokens[resetToken] = &PasswordResetToken{
		Token:     resetToken,
		Email:     email,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	u.resetTokenMu.Unlock()

	return resetToken, nil
}

// ResetPassword validates reset token and updates password
// O(1) token lookup + O(n) DB update where n = 1 (indexed email)
func (u UseCase) ResetPassword(resetToken string, newPassword string) *httpresponse.HTTPError {
	httpError := httpresponse.HTTPError{}

	// O(1) token lookup
	u.resetTokenMu.RLock()
	pwdReset, exists := u.resetTokens[resetToken]
	u.resetTokenMu.RUnlock()

	if !exists {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Invalid reset token"
		return &httpError
	}

	if time.Now().After(pwdReset.ExpiresAt) {
		u.resetTokenMu.Lock()
		delete(u.resetTokens, resetToken)
		u.resetTokenMu.Unlock()

		httpError.Code = http.StatusBadRequest
		httpError.Message = "Reset token expired"
		return &httpError
	}

	// Hash new password
	hashedPassword, err := helper.GeneratehashPassword(newPassword)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = "Failed to hash password"
		return &httpError
	}

	// Update user password in DB
	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, pwdReset.Email)
	if err != nil || authUser == nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = "User not found"
		return &httpError
	}

	user, _ := helper.TypeConverter[entity.User](&authUser)
	user.Password = hashedPassword

	if err := u.GenericRepository.Update(user); err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = "Failed to update password"
		return &httpError
	}

	// Invalidate reset token - O(1) deletion
	u.resetTokenMu.Lock()
	delete(u.resetTokens, resetToken)
	u.resetTokenMu.Unlock()

	return nil
}

// ChangePassword verifies old password and updates to new password
// O(n) DB query where n = 1 (indexed email) + O(n) DB update where n = 1
func (u UseCase) ChangePassword(email string, oldPassword string, newPassword string) *httpresponse.HTTPError {
	httpError := httpresponse.HTTPError{}

	// Find user by email - O(n) DB query where n = 1 (indexed email)
	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, email)
	if err != nil {
		log.Println("{ChangePassword}{FindByEmail}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return &httpError
	}

	if authUser == nil {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "User not found"
		return &httpError
	}

	user, _ := helper.TypeConverter[entity.User](&authUser)

	// Verify old password matches - O(1) hash comparison
	if !helper.CheckPasswordHash(oldPassword, user.Password) {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Old password is incorrect"
		return &httpError
	}

	// Hash new password - O(1) operation
	hashedPassword, err := helper.GeneratehashPassword(newPassword)
	if err != nil {
		log.Println("{ChangePassword}{GeneratehashPassword}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = "Failed to hash password"
		return &httpError
	}

	// Update user password in DB - O(n) DB update where n = 1 (indexed email)
	user.Password = hashedPassword

	if err := u.GenericRepository.Update(user); err != nil {
		log.Println("{ChangePassword}{Update}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return &httpError
	}

	return nil
}

// ValidateSession checks if user session is still active
// O(1) session lookup
func (u UseCase) ValidateSession(email string) (bool, *httpresponse.HTTPError) {
	httpError := httpresponse.HTTPError{}

	u.sessionStore.mu.RLock()
	session, exists := u.sessionStore.sessions[email]
	u.sessionStore.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if time.Now().After(session.SessionExpiry) {
		u.sessionStore.mu.Lock()
		delete(u.sessionStore.sessions, email)
		u.sessionStore.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// CheckRateLimit enforces login attempt rate limiting
// O(1) session lookup + O(1) counter update
func (u UseCase) CheckRateLimit(email string) (bool, *httpresponse.HTTPError) {
	httpError := httpresponse.HTTPError{}

	u.sessionStore.mu.Lock()
	session, exists := u.sessionStore.sessions[email]

	if !exists {
		// First attempt for this email
		u.sessionStore.sessions[email] = &SessionData{
			UserEmail:     email,
			LoginAttempts: 1,
			LastAttemptAt: time.Now(),
		}
		u.sessionStore.mu.Unlock()
		return true, nil
	}

	// Reset counter if 15 minutes have passed
	if time.Since(session.LastAttemptAt) > 15*time.Minute {
		session.LoginAttempts = 1
		session.LastAttemptAt = time.Now()
		u.sessionStore.mu.Unlock()
		return true, nil
	}

	// Allow max 5 attempts per 15 minutes
	if session.LoginAttempts >= 5 {
		u.sessionStore.mu.Unlock()
		httpError.Code = http.StatusTooManyRequests
		httpError.Message = "Too many login attempts. Try again in 15 minutes."
		return false, &httpError
	}

	session.LoginAttempts++
	session.LastAttemptAt = time.Now()
	u.sessionStore.mu.Unlock()

	return true, nil
}

// Logout invalidates user session and tokens
// O(1) token deletion + O(1) session deletion
func (u UseCase) Logout(token string) *httpresponse.HTTPError {
	httpError := httpresponse.HTTPError{}

	// O(1) token lookup and deletion
	u.tokenCache.mu.Lock()
	metadata, exists := u.tokenCache.tokens[token]
	if exists {
		delete(u.tokenCache.tokens, token)
	}
	u.tokenCache.mu.Unlock()

	if !exists {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Invalid token"
		return &httpError
	}

	// O(1) session deletion
	u.sessionStore.mu.Lock()
	delete(u.sessionStore.sessions, metadata.Email)
	u.sessionStore.mu.Unlock()

	return nil
}

// cacheToken stores token metadata with expiry - O(1) operation
func (u UseCase) cacheToken(token string, email string, role string) {
	u.tokenCache.mu.Lock()
	defer u.tokenCache.mu.Unlock()

	u.tokenCache.tokens[token] = &TokenMetadata{
		Email:     email,
		Role:      role,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
}

// createSession initializes new session for user - O(1) operation
func (u UseCase) createSession(email string) {
	u.sessionStore.mu.Lock()
	defer u.sessionStore.mu.Unlock()

	u.sessionStore.sessions[email] = &SessionData{
		UserEmail:     email,
		LoginAttempts: 0,
		LastAttemptAt: time.Now(),
		SessionExpiry: time.Now().Add(24 * time.Hour),
		RefreshTokens: make(map[string]time.Time),
	}
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

func NewUseCase(genericRepository auth.IGenericRepository, app app.App) IUseCase {
	return UseCase{
		GenericRepository: genericRepository,
		AppCfg:            app,
		tokenCache: &TokenCache{
			tokens: make(map[string]*TokenMetadata),
		},
		sessionStore: &SessionStore{
			sessions: make(map[string]*SessionData),
		},
		resetTokens: make(map[string]*PasswordResetToken),
	}
}
