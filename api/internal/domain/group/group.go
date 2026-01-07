package group

import (
	"time"

	"github.com/google/uuid"
)

// Group represents a group entity (tenant boundary) in the domain layer
type Group struct {
	ID          uuid.UUID
	Name        string
	Description *string
	OwnerID     uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	UpdatedBy   *uuid.UUID
	Version     int64
}

// NewGroup creates a new group with generated ID and timestamps
func NewGroup(name string, description *string, ownerID uuid.UUID) *Group {
	now := time.Now()
	return &Group{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
		UpdatedBy:   &ownerID,
		Version:     1,
	}
}

// Update updates group fields
func (g *Group) Update(name string, description *string, updatedBy uuid.UUID) {
	g.Name = name
	g.Description = description
	g.UpdatedAt = time.Now()
	g.UpdatedBy = &updatedBy
}
