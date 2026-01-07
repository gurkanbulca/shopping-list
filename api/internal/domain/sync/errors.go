package sync

import "errors"

var (
	// ErrMutationNotFound is returned when a processed mutation is not found
	ErrMutationNotFound = errors.New("mutation not found")

	// ErrVersionMismatch is returned when version check fails (optimistic locking)
	ErrVersionMismatch = errors.New("version mismatch - entity was modified by another user")

	// ErrNotGroupMember is returned when user is not a member of the group
	ErrNotGroupMember = errors.New("user is not a member of this group")

	// ErrInvalidMutation is returned when mutation data is invalid
	ErrInvalidMutation = errors.New("invalid mutation data")

	// ErrInvalidEntityType is returned when entity type is not recognized
	ErrInvalidEntityType = errors.New("invalid entity type")

	// ErrEntityNotFound is returned when the entity to mutate is not found
	ErrEntityNotFound = errors.New("entity not found")

	// ErrInvalidMutationType is returned when mutation type is not recognized
	ErrInvalidMutationType = errors.New("invalid mutation type")
)
