package models

import "time"

// SignupRequest is the payload for POST /api/auth/signup
// @Description New account registration details
type SignupRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=255" example:"Kiran Kumar"`
	Email    string `json:"email" binding:"required,email,max=191" example:"kiran@example.com"`
	Password string `json:"password" binding:"required,min=8,max=72" example:"Str0ng!Passphrase"`
	Address  string `json:"address" binding:"omitempty,max=1000" example:"123 MG Road, Bengaluru"`
	Phone    string `json:"phone" binding:"omitempty,max=50" example:"+91-9876543210"`
}

// LoginRequest is the payload for POST /api/auth/login
// @Description Credentials used to obtain a token pair
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=191" example:"kiran@example.com"`
	Password string `json:"password" binding:"required,max=72" example:"Str0ng!Passphrase"`
}

// RefreshRequest is the payload for POST /api/auth/refresh and /api/auth/logout
// @Description A previously issued refresh token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"v2.V0hFTiBJTiBET1VCVCwgcmVhZCB0aGUgZG9jcw"`
}

// ChangePasswordRequest is the payload for POST /api/auth/change-password
// @Description Current and replacement password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,max=72" example:"Str0ng!Passphrase"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72" example:"Even!Str0nger1"`
}

// UserResponse is the public projection of a User. It exists so that adding a
// sensitive column to the User model can never accidentally widen the API.
// @Description Public user profile
type UserResponse struct {
	ID          uint       `json:"id" example:"1"`
	Name        string     `json:"name" example:"Kiran Kumar"`
	Email       string     `json:"email" example:"kiran@example.com"`
	Address     string     `json:"address" example:"123 MG Road, Bengaluru"`
	Phone       string     `json:"phone" example:"+91-9876543210"`
	Role        string     `json:"role" example:"customer"`
	IsActive    bool       `json:"is_active" example:"true"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// NewUserResponse projects a User into its public form.
func NewUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		Address:     u.Address,
		Phone:       u.Phone,
		Role:        u.Role,
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
}

// AuthResponse is returned by signup, login and refresh.
// @Description Issued token pair plus the authenticated user
type AuthResponse struct {
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string       `json:"refresh_token" example:"3Qk9m1s..."`
	TokenType    string       `json:"token_type" example:"Bearer"`
	ExpiresIn    int          `json:"expires_in" example:"900"`
	User         UserResponse `json:"user"`
}

// MessageResponse is a generic success envelope.
// @Description Generic success message
type MessageResponse struct {
	Message string `json:"message" example:"operation completed successfully"`
}

// ErrorResponse is a generic failure envelope.
// @Description Generic error message
type ErrorResponse struct {
	Error string `json:"error" example:"invalid email or password"`
}
