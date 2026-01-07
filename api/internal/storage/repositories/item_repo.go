package repositories

import (
	"context"

	"github.com/google/uuid"
)

// Item represents an item entity
type Item struct {
	ID          uuid.UUID
	ListID      uuid.UUID
	CategoryID  *uuid.UUID
	Name        string
	Priority    string // LOW, MEDIUM, HIGH, URGENT
	SortOrder   int32
	IsPurchased bool
	Quantity    *string
	Notes       *string
	CreatedAt   int64
	UpdatedAt   int64
	UpdatedBy   *uuid.UUID
	Version     int64
}

// ItemRepository defines the interface for item data access
type ItemRepository interface {
	Create(ctx context.Context, item *Item) error
	GetByID(ctx context.Context, id uuid.UUID) (*Item, error)
	ListByListID(ctx context.Context, listID uuid.UUID) ([]*Item, error)
	Update(ctx context.Context, item *Item) error
	Delete(ctx context.Context, id uuid.UUID) error
	ReorderItems(ctx context.Context, listID uuid.UUID, itemIDs []uuid.UUID) error
}
