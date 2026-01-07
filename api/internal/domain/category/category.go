package category

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a category entity in the domain layer
type Category struct {
	ID        uuid.UUID
	GroupID   uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	UpdatedBy *uuid.UUID
	Version   int64
}

// NewCategory creates a new category with generated ID and timestamps
func NewCategory(groupID uuid.UUID, name string, createdBy uuid.UUID) *Category {
	now := time.Now()
	return &Category{
		ID:        uuid.New(),
		GroupID:   groupID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		UpdatedBy: &createdBy,
		Version:   1,
	}
}

// Update updates category fields
func (c *Category) Update(name string, updatedBy uuid.UUID) {
	c.Name = name
	c.UpdatedAt = time.Now()
	c.UpdatedBy = &updatedBy
}
