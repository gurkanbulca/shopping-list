package repositories

import (
	"context"

	"github.com/google/uuid"
)

// List represents a shopping list entity
type List struct {
	ID          uuid.UUID
	GroupID     uuid.UUID
	Name        string
	Description *string
	IsArchived  bool
	CreatedAt   int64
	UpdatedAt   int64
	UpdatedBy   *uuid.UUID
	Version     int64
}

// ListRepository defines the interface for list data access
type ListRepository interface {
	Create(ctx context.Context, list *List) error
	GetByID(ctx context.Context, id uuid.UUID) (*List, error)
	ListByGroupID(ctx context.Context, groupID uuid.UUID, includeArchived bool, limit, offset int) ([]*List, error)
	Update(ctx context.Context, list *List) error
	Archive(ctx context.Context, id uuid.UUID, archive bool) error
}
