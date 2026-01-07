package repositories

import (
	"context"

	"github.com/google/uuid"
)

// Change represents a change in delta sync
type Change struct {
	Sequence   int64
	EntityType string
	EntityID   uuid.UUID
	Operation  string // CREATE, UPDATE, DELETE
	ChangedAt  int64
}

// SyncRepository defines the interface for sync data access
type SyncRepository interface {
	GetDelta(ctx context.Context, groupID uuid.UUID, cursor int64, maxChanges int) ([]*Change, int64, error)
	RecordChange(ctx context.Context, change *Change) error
}
