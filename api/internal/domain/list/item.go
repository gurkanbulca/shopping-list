package list

import (
	"time"

	"github.com/google/uuid"
)

// Priority represents the priority level of an item
type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityMedium Priority = "MEDIUM"
	PriorityHigh   Priority = "HIGH"
	PriorityUrgent Priority = "URGENT"
)

// IsValid checks if the priority is valid
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

// Item represents a product/item in a shopping list
type Item struct {
	ID          uuid.UUID
	ListID      uuid.UUID
	CategoryID  *uuid.UUID
	Name        string
	Priority    Priority
	SortOrder   int32
	IsPurchased bool
	Quantity    *string
	Notes       *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	UpdatedBy   *uuid.UUID
	Version     int64
}

// NewItem creates a new item with generated ID and timestamps
func NewItem(listID uuid.UUID, name string, priority Priority, sortOrder int32, createdBy uuid.UUID) *Item {
	now := time.Now()
	return &Item{
		ID:          uuid.New(),
		ListID:      listID,
		Name:        name,
		Priority:    priority,
		SortOrder:   sortOrder,
		IsPurchased: false,
		CreatedAt:   now,
		UpdatedAt:   now,
		UpdatedBy:   &createdBy,
		Version:     1,
	}
}

// Update updates item fields
func (i *Item) Update(name string, priority Priority, categoryID *uuid.UUID, quantity, notes *string, updatedBy uuid.UUID) {
	i.Name = name
	i.Priority = priority
	i.CategoryID = categoryID
	i.Quantity = quantity
	i.Notes = notes
	i.UpdatedAt = time.Now()
	i.UpdatedBy = &updatedBy
}

// SetPurchased sets the purchased status
func (i *Item) SetPurchased(purchased bool, updatedBy uuid.UUID) {
	i.IsPurchased = purchased
	i.UpdatedAt = time.Now()
	i.UpdatedBy = &updatedBy
}

// SetSortOrder updates the sort order
func (i *Item) SetSortOrder(order int32) {
	i.SortOrder = order
	i.UpdatedAt = time.Now()
}
