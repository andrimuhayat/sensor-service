package dto

type Authentication struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Token struct {
	Role        string `json:"role"`
	Email       string `json:"email"`
	TokenString string `json:"token"`
}

// ChangePasswordRequest for changing password while logged in
type ChangePasswordRequest struct {
	Email       string `json:"email"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangeUserStatusRequest for activating or deactivating user account (admin only)
type ChangeUserStatusRequest struct {
	Email     string `json:"email"`
	NewStatus string `json:"new_status"`
}
