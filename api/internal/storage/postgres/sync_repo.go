package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/sync"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SyncRepository implements sync.Repository with PostgreSQL
type SyncRepository struct {
	pool *pgxpool.Pool
}

// NewSyncRepository creates a new PostgreSQL sync repository
func NewSyncRepository(db *DB) *SyncRepository {
	return &SyncRepository{pool: db.Pool}
}

// GetDelta returns changes since the given cursor for a group
func (r *SyncRepository) GetDelta(ctx context.Context, groupID uuid.UUID, cursor int64, maxChanges int) ([]*sync.Change, error) {
	query := `
		SELECT id, sequence, entity_type, entity_id, operation, group_id, changed_at, changed_by
		FROM mutation_log
		WHERE group_id = $1 AND sequence > $2
		ORDER BY sequence ASC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, groupID, cursor, maxChanges)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []*sync.Change
	for rows.Next() {
		var c sync.Change
		var entityType, operation string

		if err := rows.Scan(
			&c.ID,
			&c.Sequence,
			&entityType,
			&c.EntityID,
			&operation,
			&c.GroupID,
			&c.ChangedAt,
			&c.ChangedBy,
		); err != nil {
			return nil, err
		}

		c.EntityType = sync.EntityType(entityType)
		c.Operation = sync.ChangeOperation(operation)
		changes = append(changes, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return changes, nil
}

// RecordChange records a change in the mutation log
func (r *SyncRepository) RecordChange(ctx context.Context, change *sync.Change) error {
	query := `
		INSERT INTO mutation_log (id, entity_type, entity_id, operation, group_id, changed_at, changed_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING sequence
	`

	return r.pool.QueryRow(ctx, query,
		change.ID,
		string(change.EntityType),
		change.EntityID,
		string(change.Operation),
		change.GroupID,
		change.ChangedAt,
		change.ChangedBy,
	).Scan(&change.Sequence)
}

// GetProcessedMutation checks if a mutation was already processed
func (r *SyncRepository) GetProcessedMutation(ctx context.Context, mutationID uuid.UUID) (*sync.ProcessedMutation, error) {
	query := `
		SELECT mutation_id, entity_id, version, processed_at
		FROM processed_mutations
		WHERE mutation_id = $1
	`

	var pm sync.ProcessedMutation
	err := r.pool.QueryRow(ctx, query, mutationID).Scan(
		&pm.MutationID,
		&pm.EntityID,
		&pm.Version,
		&pm.ProcessedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sync.ErrMutationNotFound
		}
		return nil, err
	}

	return &pm, nil
}

// RecordProcessedMutation records that a mutation was processed
func (r *SyncRepository) RecordProcessedMutation(ctx context.Context, mutation *sync.ProcessedMutation) error {
	query := `
		INSERT INTO processed_mutations (mutation_id, entity_id, version, processed_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (mutation_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query,
		mutation.MutationID,
		mutation.EntityID,
		mutation.Version,
		mutation.ProcessedAt,
	)

	return err
}

// GetEntityByID retrieves an entity's current state for delta response
func (r *SyncRepository) GetListByID(ctx context.Context, id uuid.UUID) (interface{}, error) {
	query := `
		SELECT id, group_id, name, description, is_archived, created_at, updated_at, updated_by, version
		FROM lists
		WHERE id = $1
	`

	var result struct {
		ID          uuid.UUID
		GroupID     uuid.UUID
		Name        string
		Description *string
		IsArchived  bool
		CreatedAt   interface{}
		UpdatedAt   interface{}
		UpdatedBy   *uuid.UUID
		Version     int64
	}

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.GroupID,
		&result.Name,
		&result.Description,
		&result.IsArchived,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.UpdatedBy,
		&result.Version,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetItemByID retrieves an item by ID
func (r *SyncRepository) GetItemByID(ctx context.Context, id uuid.UUID) (interface{}, error) {
	query := `
		SELECT id, list_id, category_id, name, priority, sort_order, is_purchased, quantity, notes, created_at, updated_at, updated_by, version
		FROM items
		WHERE id = $1
	`

	var result struct {
		ID          uuid.UUID
		ListID      uuid.UUID
		CategoryID  *uuid.UUID
		Name        string
		Priority    string
		SortOrder   int32
		IsPurchased bool
		Quantity    *string
		Notes       *string
		CreatedAt   interface{}
		UpdatedAt   interface{}
		UpdatedBy   *uuid.UUID
		Version     int64
	}

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.ListID,
		&result.CategoryID,
		&result.Name,
		&result.Priority,
		&result.SortOrder,
		&result.IsPurchased,
		&result.Quantity,
		&result.Notes,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.UpdatedBy,
		&result.Version,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetCategoryByID retrieves a category by ID
func (r *SyncRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (interface{}, error) {
	query := `
		SELECT id, group_id, name, created_at, updated_at, updated_by, version
		FROM categories
		WHERE id = $1
	`

	var result struct {
		ID        uuid.UUID
		GroupID   uuid.UUID
		Name      string
		CreatedAt interface{}
		UpdatedAt interface{}
		UpdatedBy *uuid.UUID
		Version   int64
	}

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.GroupID,
		&result.Name,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.UpdatedBy,
		&result.Version,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
