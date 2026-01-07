package sync

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for sync data access
type Repository interface {
	// GetDelta returns changes since the given cursor for a group
	GetDelta(ctx context.Context, groupID uuid.UUID, cursor int64, maxChanges int) ([]*Change, error)

	// RecordChange records a change in the mutation log
	RecordChange(ctx context.Context, change *Change) error

	// GetProcessedMutation checks if a mutation was already processed (for idempotency)
	GetProcessedMutation(ctx context.Context, mutationID uuid.UUID) (*ProcessedMutation, error)

	// RecordProcessedMutation records that a mutation was processed
	RecordProcessedMutation(ctx context.Context, mutation *ProcessedMutation) error
}

// MembershipChecker defines the interface for checking group membership
type MembershipChecker interface {
	IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
}

// ListRepository provides access to list operations for mutations
type ListRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (interface{}, error)
}

// ItemRepository provides access to item operations for mutations
type ItemRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (interface{}, error)
}

// CategoryRepository provides access to category operations for mutations
type CategoryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (interface{}, error)
}
