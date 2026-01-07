package validation

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// emailRegex is a simple email validation regex
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// phoneRegex validates E.164 phone number format (e.g., +14155552671)
	phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
)

// AuthValidationError represents a validation error for auth operations
type AuthValidationError struct {
	Field   string
	Message string
}

func (e *AuthValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return nil // Optional field
	}

	email = strings.TrimSpace(email)
	if !emailRegex.MatchString(email) {
		return &AuthValidationError{
			Field:   "email",
			Message: "invalid email format",
		}
	}

	return nil
}

// ValidatePhone validates a phone number in E.164 format
func ValidatePhone(phone string) error {
	if phone == "" {
		return nil // Optional field
	}

	phone = strings.TrimSpace(phone)
	if !phoneRegex.MatchString(phone) {
		return &AuthValidationError{
			Field:   "phone",
			Message: "invalid phone format, must be E.164 format (e.g., +14155552671)",
		}
	}

	return nil
}

// ValidatePassword validates a password meets security requirements
// Requirements: at least 8 characters, at least one letter and one number
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return &AuthValidationError{
			Field:   "password",
			Message: "password must be at least 8 characters",
		}
	}

	hasLetter := false
	hasNumber := false

	for _, c := range password {
		if unicode.IsLetter(c) {
			hasLetter = true
		}
		if unicode.IsDigit(c) {
			hasNumber = true
		}
	}

	if !hasLetter || !hasNumber {
		return &AuthValidationError{
			Field:   "password",
			Message: "password must contain at least one letter and one number",
		}
	}

	return nil
}

// ValidateName validates a user's display name
func ValidateName(name string) error {
	if name == "" {
		return nil // Optional field
	}

	name = strings.TrimSpace(name)
	if len(name) > 100 {
		return &AuthValidationError{
			Field:   "name",
			Message: "name must be 100 characters or less",
		}
	}

	return nil
}

// ValidateRegisterRequest validates a registration request
func ValidateRegisterRequest(email, phone, password, name string) error {
	// At least one identifier is required
	email = strings.TrimSpace(email)
	phone = strings.TrimSpace(phone)

	if email == "" && phone == "" {
		return &AuthValidationError{
			Field:   "email/phone",
			Message: "email or phone is required",
		}
	}

	if err := ValidateEmail(email); err != nil {
		return err
	}

	if err := ValidatePhone(phone); err != nil {
		return err
	}

	if password == "" {
		return &AuthValidationError{
			Field:   "password",
			Message: "password is required",
		}
	}

	if err := ValidatePassword(password); err != nil {
		return err
	}

	if err := ValidateName(name); err != nil {
		return err
	}

	return nil
}

// ValidateLoginRequest validates a login request
func ValidateLoginRequest(email, phone, password string) error {
	email = strings.TrimSpace(email)
	phone = strings.TrimSpace(phone)

	if email == "" && phone == "" {
		return &AuthValidationError{
			Field:   "email/phone",
			Message: "email or phone is required",
		}
	}

	if password == "" {
		return &AuthValidationError{
			Field:   "password",
			Message: "password is required",
		}
	}

	return nil
}

// ValidateRefreshTokenRequest validates a refresh token request
func ValidateRefreshTokenRequest(refreshToken string) error {
	if refreshToken == "" {
		return &AuthValidationError{
			Field:   "refresh_token",
			Message: "refresh token is required",
		}
	}

	return nil
}
