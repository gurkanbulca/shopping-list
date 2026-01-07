package repositories

import (
	"context"

	"github.com/google/uuid"
)

// User represents a user entity
type User struct {
	ID           uuid.UUID
	Email        *string
	Phone        *string
	PasswordHash string
	Name         *string
	CreatedAt    int64
	UpdatedAt    int64
	Version      int64
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	Update(ctx context.Context, user *User) error
}
