package list

import (
	"time"

	"github.com/google/uuid"
)

// List represents a shopping list entity in the domain layer
type List struct {
	ID          uuid.UUID
	GroupID     uuid.UUID
	Name        string
	Description *string
	IsArchived  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	UpdatedBy   *uuid.UUID
	Version     int64
}

// NewList creates a new list with generated ID and timestamps
func NewList(groupID uuid.UUID, name string, description *string, createdBy uuid.UUID) *List {
	now := time.Now()
	return &List{
		ID:          uuid.New(),
		GroupID:     groupID,
		Name:        name,
		Description: description,
		IsArchived:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
		UpdatedBy:   &createdBy,
		Version:     1,
	}
}

// Update updates list fields with version check
func (l *List) Update(name string, description *string, updatedBy uuid.UUID) {
	l.Name = name
	l.Description = description
	l.UpdatedAt = time.Now()
	l.UpdatedBy = &updatedBy
}

// Archive sets the archived status
func (l *List) Archive(archive bool, updatedBy uuid.UUID) {
	l.IsArchived = archive
	l.UpdatedAt = time.Now()
	l.UpdatedBy = &updatedBy
}
