package list

import "errors"

var (
	// ErrListNotFound is returned when a list is not found
	ErrListNotFound = errors.New("list not found")

	// ErrItemNotFound is returned when an item is not found
	ErrItemNotFound = errors.New("item not found")

	// ErrVersionMismatch is returned when version check fails (optimistic locking)
	ErrVersionMismatch = errors.New("version mismatch - item was modified by another user")

	// ErrNotGroupMember is returned when user is not a member of the group
	ErrNotGroupMember = errors.New("user is not a member of this group")

	// ErrInvalidListName is returned when list name is invalid
	ErrInvalidListName = errors.New("list name must be 1-100 characters")

	// ErrInvalidItemName is returned when item name is invalid
	ErrInvalidItemName = errors.New("item name must be 1-200 characters")

	// ErrInvalidPriority is returned when priority is invalid
	ErrInvalidPriority = errors.New("invalid priority")

	// ErrListArchived is returned when trying to modify an archived list
	ErrListArchived = errors.New("cannot modify archived list")
)
