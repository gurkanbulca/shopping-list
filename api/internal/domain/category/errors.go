package category

import "errors"

var (
	// ErrCategoryNotFound is returned when a category is not found
	ErrCategoryNotFound = errors.New("category not found")

	// ErrDuplicateCategoryName is returned when a category name already exists in the group
	ErrDuplicateCategoryName = errors.New("category name already exists in this group")

	// ErrVersionMismatch is returned when version check fails (optimistic locking)
	ErrVersionMismatch = errors.New("version mismatch - category was modified by another user")

	// ErrNotGroupMember is returned when user is not a member of the group
	ErrNotGroupMember = errors.New("user is not a member of this group")

	// ErrInvalidCategoryName is returned when category name is invalid
	ErrInvalidCategoryName = errors.New("category name must be 1-50 characters")
)
