package validation

import (
	"strings"

	"github.com/google/uuid"
)

// ValidateGroupName validates a group name
func ValidateGroupName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return &AuthValidationError{
			Field:   "name",
			Message: "group name is required",
		}
	}
	if len(name) > 100 {
		return &AuthValidationError{
			Field:   "name",
			Message: "group name must be 100 characters or less",
		}
	}
	return nil
}

// ValidateGroupDescription validates a group description
func ValidateGroupDescription(description string) error {
	if len(description) > 500 {
		return &AuthValidationError{
			Field:   "description",
			Message: "group description must be 500 characters or less",
		}
	}
	return nil
}

// ValidateCreateGroupRequest validates a create group request
func ValidateCreateGroupRequest(name, description string) error {
	if err := ValidateGroupName(name); err != nil {
		return err
	}
	if err := ValidateGroupDescription(description); err != nil {
		return err
	}
	return nil
}

// ValidateGroupID validates a group ID
func ValidateGroupID(groupID string) (uuid.UUID, error) {
	if groupID == "" {
		return uuid.Nil, &AuthValidationError{
			Field:   "group_id",
			Message: "group ID is required",
		}
	}
	id, err := uuid.Parse(groupID)
	if err != nil {
		return uuid.Nil, &AuthValidationError{
			Field:   "group_id",
			Message: "invalid group ID format",
		}
	}
	return id, nil
}

// ValidateUserID validates a user ID
func ValidateUserID(userID string) (uuid.UUID, error) {
	if userID == "" {
		return uuid.Nil, &AuthValidationError{
			Field:   "user_id",
			Message: "user ID is required",
		}
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, &AuthValidationError{
			Field:   "user_id",
			Message: "invalid user ID format",
		}
	}
	return id, nil
}

// ValidateInviteMemberRequest validates an invite member request
func ValidateInviteMemberRequest(groupID, email, phone string, role int32) error {
	if _, err := ValidateGroupID(groupID); err != nil {
		return err
	}

	email = strings.TrimSpace(email)
	phone = strings.TrimSpace(phone)

	if email == "" && phone == "" {
		return &AuthValidationError{
			Field:   "email/phone",
			Message: "email or phone is required to invite a member",
		}
	}

	if email != "" {
		if err := ValidateEmail(email); err != nil {
			return err
		}
	}

	if phone != "" {
		if err := ValidatePhone(phone); err != nil {
			return err
		}
	}

	// Role 0 is unspecified, 1 is owner (not allowed for invites)
	if role == 0 {
		return &AuthValidationError{
			Field:   "role",
			Message: "role is required",
		}
	}
	if role == 1 {
		return &AuthValidationError{
			Field:   "role",
			Message: "cannot invite as owner",
		}
	}

	return nil
}

// ValidateUpdateMemberRoleRequest validates an update member role request
func ValidateUpdateMemberRoleRequest(groupID, userID string, role int32) error {
	if _, err := ValidateGroupID(groupID); err != nil {
		return err
	}
	if _, err := ValidateUserID(userID); err != nil {
		return err
	}

	// Role 0 is unspecified, 1 is owner (not allowed via this method)
	if role == 0 {
		return &AuthValidationError{
			Field:   "role",
			Message: "role is required",
		}
	}
	if role == 1 {
		return &AuthValidationError{
			Field:   "role",
			Message: "cannot assign owner role; use ownership transfer instead",
		}
	}

	return nil
}

// ValidatePagination validates pagination parameters
func ValidatePagination(pageSize int32) (int, error) {
	if pageSize <= 0 {
		return 20, nil // Default page size
	}
	if pageSize > 100 {
		return 100, nil // Max page size
	}
	return int(pageSize), nil
}
