package sync

import (
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

// MutationType represents the type of mutation
type MutationType string

const (
	MutationTypeCreate MutationType = "CREATE"
	MutationTypeUpdate MutationType = "UPDATE"
	MutationTypeDelete MutationType = "DELETE"
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

// NewChange creates a new change record
func NewChange(entityType EntityType, entityID uuid.UUID, operation ChangeOperation, groupID *uuid.UUID, changedBy *uuid.UUID) *Change {
	return &Change{
		ID:         uuid.New(),
		EntityType: entityType,
		EntityID:   entityID,
		Operation:  operation,
		GroupID:    groupID,
		ChangedAt:  time.Now(),
		ChangedBy:  changedBy,
	}
}

// Mutation represents a local mutation to push
type Mutation struct {
	MutationID      uuid.UUID
	Type            MutationType
	EntityType      string
	EntityID        *uuid.UUID // nil for creates
	EntityData      []byte     // JSON encoded entity data
	ExpectedVersion int64
}

// MutationResult represents the result of processing a mutation
type MutationResult struct {
	MutationID   uuid.UUID
	Success      bool
	ErrorCode    string
	ErrorMessage string
	EntityID     uuid.UUID
	Version      int64
}

// ProcessedMutation represents a processed mutation for idempotency
type ProcessedMutation struct {
	MutationID  uuid.UUID
	EntityID    uuid.UUID
	Version     int64
	ProcessedAt time.Time
}

// NewProcessedMutation creates a new processed mutation record
func NewProcessedMutation(mutationID, entityID uuid.UUID, version int64) *ProcessedMutation {
	return &ProcessedMutation{
		MutationID:  mutationID,
		EntityID:    entityID,
		Version:     version,
		ProcessedAt: time.Now(),
	}
}
