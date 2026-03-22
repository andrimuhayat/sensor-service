# Auth Module

## Overview

The Auth module provides comprehensive authentication and authorization functionality for the sensor-service application. It implements secure user authentication, token management, session handling, and rate limiting.

## Architecture

### Components

1. **UseCase Layer** (`usecase.go`)
   - Core business logic for authentication operations
   - Manages token lifecycle and session state
   - Implements rate limiting and security policies

2. **Repository Layer** (`repository/`)
   - Generic repository pattern for database operations
   - Handles user data persistence
   - Supports CRUD operations with type safety

3. **Entity Layer** (`entity/`)
   - `User` struct: Represents user data (email, password, role)

4. **DTO Layer** (`dto/`)
   - `Authentication`: Request payload for sign in/up
   - `Token`: Response payload with JWT token

## Authentication Flow

### Sign In Flow
```
1. User submits email + password
2. Rate limit check (max 5 attempts per 15 minutes)
3. Find user by email in database
4. Verify password hash
5. Generate JWT token
6. Cache token metadata (O(1) lookup)
7. Create session for user
8. Return token to client
```

### Sign Up Flow
```
1. User submits email + password + role
2. Validate role (must be 'admin' or 'user')
3. Check if email already exists
4. Hash password using bcrypt
5. Create user in database
6. Return created user
```

### Token Refresh Flow
```
1. Client sends old token
2. Lookup token in cache (O(1))
3. Verify token not expired
4. Generate new JWT token
5. Cache new token metadata
6. Invalidate old token
7. Return new token
```

### Password Reset Flow
```
1. User requests password reset with email
2. Verify email exists in database
3. Generate secure random reset token (32 bytes)
4. Store reset token with 15-minute expiry
5. Return reset token to user (typically sent via email)
6. User submits reset token + new password
7. Validate reset token exists and not expired
8. Hash new password
9. Update user password in database
10. Invalidate reset token
```

## Key Features

### 1. Token Management
- **JWT Tokens**: Secure token generation with configurable expiry (24 hours)
- **Token Cache**: O(1) lookup for token validation
- **Token Metadata**: Stores email, role, creation time, expiry
- **Token Invalidation**: Automatic cleanup on logout or refresh

### 2. Session Management
- **Session Store**: O(1) session lookup by email
- **Session Expiry**: 24-hour session lifetime
- **Session Cleanup**: Automatic removal of expired sessions
- **Refresh Token Tracking**: Maintains refresh token history per session

### 3. Rate Limiting
- **Login Attempt Limiting**: Maximum 5 attempts per 15 minutes
- **Per-User Tracking**: Separate counters for each email
- **Automatic Reset**: Counter resets after 15-minute window
- **HTTP 429 Response**: Returns "Too Many Requests" when limit exceeded

### 4. Password Security
- **Bcrypt Hashing**: Industry-standard password hashing
- **Reset Token Security**: Cryptographically secure random tokens
- **Token Expiry**: Reset tokens expire after 15 minutes
- **One-Time Use**: Reset tokens invalidated after use

### 5. Role-Based Access
- **Supported Roles**: 'admin', 'user'
- **Role Validation**: Enforced during sign up
- **Role Persistence**: Stored in JWT token and database

## Performance Characteristics

### Time Complexity
- **SignIn**: O(n) where n=1 (indexed email lookup) + O(1) token cache
- **SignUp**: O(1) role validation + O(n) where n=1 (indexed email lookup)
- **RefreshToken**: O(1) token cache lookup + O(1) token generation
- **InitiatePasswordReset**: O(n) where n=1 (indexed email lookup) + O(1) token storage
- **ResetPassword**: O(1) token lookup + O(n) where n=1 (indexed email lookup)
- **ValidateSession**: O(1) session lookup
- **CheckRateLimit**: O(1) session lookup + O(1) counter update
- **Logout**: O(1) token deletion + O(1) session deletion

### Space Complexity
- **TokenCache**: O(m) where m = number of active tokens
- **SessionStore**: O(u) where u = number of active users
- **ResetTokens**: O(r) where r = number of pending password resets

## Data Structures

### TokenCache
```go
type TokenCache struct {
    mu     sync.RWMutex
    tokens map[string]*TokenMetadata  // O(1) lookup
}

type TokenMetadata struct {
    Email     string
    Role      string
    ExpiresAt time.Time
    CreatedAt time.Time
}
```

### SessionStore
```go
type SessionStore struct {
    mu       sync.RWMutex
    sessions map[string]*SessionData  // O(1) lookup
}

type SessionData struct {
    UserEmail      string
    LoginAttempts  int
    LastAttemptAt  time.Time
    SessionExpiry  time.Time
    RefreshTokens  map[string]time.Time  // O(1) refresh token lookup
}
```

### PasswordResetToken
```go
type PasswordResetToken struct {
    Token     string
    Email     string
    ExpiresAt time.Time
}
```

## Error Handling

### HTTP Status Codes
- **200 OK**: Successful authentication/authorization
- **400 Bad Request**: Invalid credentials, duplicate email, invalid role, expired token
- **401 Unauthorized**: Invalid or expired token
- **429 Too Many Requests**: Rate limit exceeded
- **500 Internal Server Error**: Database or system errors

### Error Messages
- "Username or Password is incorrect": Invalid credentials
- "Role doesn't exists, please use admin or user": Invalid role
- "email cannot be same": Duplicate email
- "Token expired or invalid": Invalid or expired token
- "Too many login attempts. Try again in 15 minutes.": Rate limit exceeded
- "Invalid reset token": Reset token not found
- "Reset token expired": Reset token has expired

## Thread Safety

All shared data structures use `sync.RWMutex` for thread-safe access:
- **TokenCache**: Protected by `mu` for concurrent token operations
- **SessionStore**: Protected by `mu` for concurrent session operations
- **ResetTokens**: Protected by `resetTokenMu` for concurrent reset token operations

## Testing

Comprehensive test coverage in `usecase_test.go` includes:

### SignIn Tests (4 tests)
- Happy path: Valid credentials
- Error case: User not found
- Error case: Incorrect password
- Rate limiting: 5 attempts blocked

### SignUp Tests (3 tests)
- Happy path: Valid user creation
- Error case: Duplicate email
- Error case: Invalid role

### RefreshToken Tests (3 tests)
- Happy path: Valid token refresh
- Error case: Token not found
- Error case: Token expired

### InitiatePasswordReset Tests (2 tests)
- Happy path: Valid email
- Error case: Email not found

### ResetPassword Tests (3 tests)
- Happy path: Valid reset token
- Error case: Invalid token
- Error case: Expired token

### ValidateSession Tests (3 tests)
- Happy path: Active session
- Error case: Session not found
- Error case: Session expired

### CheckRateLimit Tests (4 tests)
- Happy path: First attempt allowed
- Happy path: 5 attempts allowed
- Error case: 6th attempt blocked
- Happy path: Counter reset after 15 minutes

### Logout Tests (2 tests)
- Happy path: Valid token logout
- Error case: Invalid token

### Helper Function Tests (4 tests)
- Secure token generation
- Token uniqueness
- Token caching
- Session creation

**Total: 32 comprehensive test cases**

## Usage Examples

### Initialize UseCase
```go
repo := repository.NewGenericRepository(db)
appConfig := app.App{SecretKey: "your-secret-key"}
useCase := usecase.NewUseCase(repo, appConfig)
```

### Sign In
```go
request := config.HTTPRequest{
    Body: map[string]interface{}{
        "email": "user@example.com",
        "password": "password123",
    },
}
token, err := useCase.SignIn(request)
if err != nil {
    // Handle error
}
// Use token.TokenString in Authorization header
```

### Sign Up
```go
request := config.HTTPRequest{
    Body: map[string]interface{}{
        "email": "newuser@example.com",
        "password": "password123",
        "roles": "user",
    },
}
user, err := useCase.SignUp(request)
if err != nil {
    // Handle error
}
```

### Refresh Token
```go
newToken, err := useCase.RefreshToken(oldToken)
if err != nil {
    // Handle error
}
```

### Initiate Password Reset
```go
resetToken, err := useCase.InitiatePasswordReset("user@example.com")
if err != nil {
    // Handle error
}
// Send resetToken to user via email
```

### Reset Password
```go
err := useCase.ResetPassword(resetToken, "newPassword123")
if err != nil {
    // Handle error
}
```

### Validate Session
```go
isValid, err := useCase.ValidateSession("user@example.com")
if !isValid {
    // Session expired or not found
}
```

### Check Rate Limit
```go
allowed, err := useCase.CheckRateLimit("user@example.com")
if !allowed {
    // Rate limit exceeded
}
```

### Logout
```go
err := useCase.Logout(token)
if err != nil {
    // Handle error
}
```

## Security Considerations

1. **Password Storage**: Always use bcrypt for password hashing
2. **Token Expiry**: Tokens expire after 24 hours
3. **Reset Token Expiry**: Reset tokens expire after 15 minutes
4. **Rate Limiting**: Prevents brute force attacks (5 attempts per 15 minutes)
5. **Secure Token Generation**: Uses `crypto/rand` for cryptographically secure tokens
6. **Thread Safety**: All shared state protected by mutexes
7. **Session Isolation**: Each user has isolated session data

## Future Enhancements

1. **Token Blacklist**: Implement persistent token blacklist for revocation
2. **Multi-Factor Authentication**: Add 2FA support
3. **OAuth2 Integration**: Support third-party authentication
4. **Audit Logging**: Track authentication events
5. **IP Whitelisting**: Restrict login by IP address
6. **Device Tracking**: Track and manage user devices
7. **Persistent Sessions**: Store sessions in Redis for scalability
8. **Email Verification**: Verify email during sign up
