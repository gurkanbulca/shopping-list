package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity in the domain layer
type User struct {
	ID           uuid.UUID
	Email        *string
	Phone        *string
	PasswordHash string
	Name         *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Version      int64
}

// NewUser creates a new user with generated ID and timestamps
func NewUser(email, phone *string, passwordHash string, name *string) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		Phone:        phone,
		PasswordHash: passwordHash,
		Name:         name,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      1,
	}
}

// HasEmail returns true if user has an email
func (u *User) HasEmail() bool {
	return u.Email != nil && *u.Email != ""
}

// HasPhone returns true if user has a phone number
func (u *User) HasPhone() bool {
	return u.Phone != nil && *u.Phone != ""
}

// GetDisplayName returns the display name or email/phone as fallback
func (u *User) GetDisplayName() string {
	if u.Name != nil && *u.Name != "" {
		return *u.Name
	}
	if u.HasEmail() {
		return *u.Email
	}
	if u.HasPhone() {
		return *u.Phone
	}
	return "User"
}
