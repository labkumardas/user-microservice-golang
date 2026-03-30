package interfaces

import (
	"context"

	"user-microservice-golang/model"
)

// UserService defines the business-logic contract for user operations.
// Controllers depend only on this interface, never on the concrete service.
type UserService interface {
	// Register validates the request, hashes the password, and creates the user
	Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error)
	//checkEmail(ctx context.Context, req *model.User) (*model.RegisterRequest, error)

	// Login authenticates credentials and returns a signed JWT
	Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error)

	// GetByID returns a sanitised user response for the given ID
	GetByID(ctx context.Context, id string) (*model.UserResponse, error)

	// GetAll returns a paginated list of users (admin)
	GetAll(ctx context.Context, page, limit int) (*model.PaginatedUsersResponse, error)

	// UpdateProfile applies allowed profile field updates
	UpdateProfile(ctx context.Context, id string, req *model.UpdateUserRequest) (*model.UserResponse, error)

	// UpdatePassword validates the current password and sets the new one
	UpdatePassword(ctx context.Context, id string, req *model.UpdatePasswordRequest) error

	// UpdateStatus changes a user's account status (admin)
	UpdateStatus(ctx context.Context, id string, req *model.UpdateStatusRequest) (*model.UserResponse, error)

	// DeleteUser soft-deletes a user account
	DeleteUser(ctx context.Context, id string) error
}
