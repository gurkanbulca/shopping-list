package list

import (
	"context"

	"github.com/google/uuid"
)

// ListRepository defines the interface for list data access
type ListRepository interface {
	// Create inserts a new list into the database
	Create(ctx context.Context, list *List) error

	// GetByID retrieves a list by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*List, error)

	// ListByGroupID retrieves all lists in a group
	ListByGroupID(ctx context.Context, groupID uuid.UUID, includeArchived bool, limit, offset int) ([]*List, int, error)

	// Update updates an existing list with version check
	Update(ctx context.Context, list *List, expectedVersion int64) error

	// Archive sets the archived status of a list
	Archive(ctx context.Context, id uuid.UUID, archive bool, updatedBy uuid.UUID) (*List, error)

	// CountByGroupID counts the total number of lists in a group
	CountByGroupID(ctx context.Context, groupID uuid.UUID, includeArchived bool) (int, error)

	// GetGroupIDByListID retrieves the group ID for a list
	GetGroupIDByListID(ctx context.Context, listID uuid.UUID) (uuid.UUID, error)
}

// ItemRepository defines the interface for item data access
type ItemRepository interface {
	// Create inserts a new item into the database
	Create(ctx context.Context, item *Item) error

	// GetByID retrieves an item by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Item, error)

	// ListByListID retrieves all items in a list with default sorting
	// Sorting: is_purchased=false first, priority DESC, sort_order ASC
	ListByListID(ctx context.Context, listID uuid.UUID) ([]*Item, error)

	// Update updates an existing item with version check
	Update(ctx context.Context, item *Item, expectedVersion int64) error

	// Delete removes an item from the database
	Delete(ctx context.Context, id uuid.UUID) error

	// GetMaxSortOrder returns the maximum sort order in a list
	GetMaxSortOrder(ctx context.Context, listID uuid.UUID) (int32, error)

	// ReorderItems updates sort_order for multiple items
	ReorderItems(ctx context.Context, listID uuid.UUID, itemIDs []uuid.UUID) error

	// GetListIDByItemID retrieves the list ID for an item
	GetListIDByItemID(ctx context.Context, itemID uuid.UUID) (uuid.UUID, error)
}

// MembershipChecker defines the interface for checking group membership
type MembershipChecker interface {
	// IsMember checks if a user is an active member of a group
	IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
}
