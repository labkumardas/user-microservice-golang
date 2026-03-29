package model

import "time"

// ─── Request DTOs ────────────────────────────────────────────────────────────

// RegisterRequest is the payload for creating a new user
type RegisterRequest struct {
	FirstName string `json:"first_name" binding:"required,min=2,max=50"`
	LastName  string `json:"last_name"  binding:"required,min=2,max=50"`
	Email     string `json:"email"      binding:"required,email"`
	Password  string `json:"password"   binding:"required,min=8"`
	Phone     string `json:"phone"      binding:"omitempty,e164"`
}

// LoginRequest is the payload for user authentication
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateUserRequest is the payload for updating user profile
type UpdateUserRequest struct {
	FirstName string `json:"first_name" binding:"omitempty,min=2,max=50"`
	LastName  string `json:"last_name"  binding:"omitempty,min=2,max=50"`
	Phone     string `json:"phone"      binding:"omitempty,e164"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,url"`
}

// UpdatePasswordRequest handles password changes
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password"     binding:"required,min=8"`
}

// UpdateStatusRequest allows admins to change user status
type UpdateStatusRequest struct {
	Status UserStatus `json:"status" binding:"required,oneof=active inactive banned"`
}

// ─── Response DTOs ───────────────────────────────────────────────────────────

// UserResponse is the sanitised user payload returned to callers
type UserResponse struct {
	ID        string     `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone,omitempty"`
	Role      Role       `json:"role"`
	Status    UserStatus `json:"status"`
	AvatarURL string     `json:"avatar_url,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// AuthResponse is returned after successful login / register
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// PaginatedUsersResponse wraps a list of users with pagination metadata
type PaginatedUsersResponse struct {
	Data       []UserResponse `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int64          `json:"total_pages"`
}

// ToUserResponse converts a User document to the safe response DTO
func ToUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:        u.ID.Hex(),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Phone:     u.Phone,
		Role:      u.Role,
		Status:    u.Status,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
