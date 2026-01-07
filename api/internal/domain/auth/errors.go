package auth

import "errors"

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrDuplicateEmail is returned when an email is already in use
	ErrDuplicateEmail = errors.New("email already exists")

	// ErrDuplicatePhone is returned when a phone number is already in use
	ErrDuplicatePhone = errors.New("phone number already exists")

	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired is returned when a token has expired
	ErrTokenExpired = errors.New("token expired")

	// ErrMissingIdentifier is returned when neither email nor phone is provided
	ErrMissingIdentifier = errors.New("email or phone is required")

	// ErrWeakPassword is returned when a password doesn't meet requirements
	ErrWeakPassword = errors.New("password must be at least 8 characters with at least one letter and one number")
)
