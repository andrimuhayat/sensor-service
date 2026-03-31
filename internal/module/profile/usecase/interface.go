package usecase

import (
	"sensor-service/config"
	"sensor-service/internal/module/profile/entity"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

// IProfileUseCase defines the interface for profile business logic
type IProfileUseCase interface {
	// GetUserProfile retrieves a user's profile by email
	// O(1) database lookup
	GetUserProfile(email string) (*entity.Profile, *httpresponse.HTTPError)

	// UpdateUserProfile updates a user's profile with validation
	// O(1) database update
	UpdateUserProfile(email string, request config.HTTPRequest, callerRole string) (*entity.Profile, *httpresponse.HTTPError)
}
