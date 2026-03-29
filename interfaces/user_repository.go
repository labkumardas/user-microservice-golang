package interfaces

import (
	"context"

	"user-microservice-golang/model"
)

// UserRepository defines the persistence contract for User entities.
// All repository implementations must satisfy this interface.
type UserRepository interface {
	// Create inserts a new user document and returns the created user
	Create(ctx context.Context, user *model.User) (*model.User, error)

	// FindByID retrieves a user by its MongoDB ObjectID string
	FindByID(ctx context.Context, id string) (*model.User, error)

	// FindByEmail retrieves a user by email address
	FindByEmail(ctx context.Context, email string) (*model.User, error)

	// FindAll retrieves a paginated list of non-deleted users
	FindAll(ctx context.Context, page, limit int) ([]*model.User, int64, error)

	// Update applies partial updates to a user document
	Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error)

	// UpdatePassword replaces the hashed password field
	UpdatePassword(ctx context.Context, id, hashedPassword string) error

	// UpdateStatus changes the account status (admin operation)
	UpdateStatus(ctx context.Context, id string, status model.UserStatus) error

	// SoftDelete marks the user as deleted without removing the document
	SoftDelete(ctx context.Context, id string) error

	// ExistsByEmail returns true if any non-deleted user owns that email
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
