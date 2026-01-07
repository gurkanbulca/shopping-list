package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EntityType represents the type of entity in sync
type EntityType string

const (
	EntityTypeUser        EntityType = "USER"
	EntityTypeGroup       EntityType = "GROUP"
	EntityTypeGroupMember EntityType = "GROUP_MEMBER"
	EntityTypeList        EntityType = "LIST"
	EntityTypeItem        EntityType = "ITEM"
	EntityTypeCategory    EntityType = "CATEGORY"
)

// ChangeOperation represents the type of change operation
type ChangeOperation string

const (
	ChangeOperationCreate ChangeOperation = "CREATE"
	ChangeOperationUpdate ChangeOperation = "UPDATE"
	ChangeOperationDelete ChangeOperation = "DELETE"
)

// Change represents a change in delta sync
type Change struct {
	ID         uuid.UUID
	Sequence   int64
	EntityType EntityType
	EntityID   uuid.UUID
	Operation  ChangeOperation
	GroupID    *uuid.UUID
	ChangedAt  time.Time
	ChangedBy  *uuid.UUID
}

// ProcessedMutation represents a processed mutation for idempotency
type ProcessedMutation struct {
	MutationID  uuid.UUID
	EntityID    uuid.UUID
	Version     int64
	ProcessedAt time.Time
}

// SyncRepository defines the interface for sync data access
type SyncRepository interface {
	// GetDelta returns changes since the given cursor for a group
	GetDelta(ctx context.Context, groupID uuid.UUID, cursor int64, maxChanges int) ([]*Change, error)

	// RecordChange records a change in the mutation log
	RecordChange(ctx context.Context, change *Change) error

	// GetProcessedMutation checks if a mutation was already processed (for idempotency)
	GetProcessedMutation(ctx context.Context, mutationID uuid.UUID) (*ProcessedMutation, error)

	// RecordProcessedMutation records that a mutation was processed
	RecordProcessedMutation(ctx context.Context, mutation *ProcessedMutation) error
}
