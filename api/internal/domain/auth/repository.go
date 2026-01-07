package auth

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create inserts a new user into the database
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by their ID
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail retrieves a user by their email address
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByPhone retrieves a user by their phone number
	GetByPhone(ctx context.Context, phone string) (*User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// ExistsByEmail checks if an email is already in use
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByPhone checks if a phone number is already in use
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
}
