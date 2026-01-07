package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/list"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ListRepository implements list.ListRepository with PostgreSQL
type ListRepository struct {
	pool *pgxpool.Pool
}

// NewListRepository creates a new PostgreSQL list repository
func NewListRepository(db *DB) *ListRepository {
	return &ListRepository{pool: db.Pool}
}

// Create inserts a new list into the database
func (r *ListRepository) Create(ctx context.Context, l *list.List) error {
	query := `
		INSERT INTO lists (id, group_id, name, description, is_archived, created_at, updated_at, updated_by, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		l.ID,
		l.GroupID,
		l.Name,
		l.Description,
		l.IsArchived,
		l.CreatedAt,
		l.UpdatedAt,
		l.UpdatedBy,
		l.Version,
	)

	return err
}

// GetByID retrieves a list by its ID
func (r *ListRepository) GetByID(ctx context.Context, id uuid.UUID) (*list.List, error) {
	query := `
		SELECT id, group_id, name, description, is_archived, created_at, updated_at, updated_by, version
		FROM lists
		WHERE id = $1
	`

	var l list.List
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&l.ID,
		&l.GroupID,
		&l.Name,
		&l.Description,
		&l.IsArchived,
		&l.CreatedAt,
		&l.UpdatedAt,
		&l.UpdatedBy,
		&l.Version,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, list.ErrListNotFound
		}
		return nil, err
	}

	return &l, nil
}

// ListByGroupID retrieves all lists in a group
func (r *ListRepository) ListByGroupID(ctx context.Context, groupID uuid.UUID, includeArchived bool, limit, offset int) ([]*list.List, int, error) {
	var query string
	var args []interface{}

	if includeArchived {
		query = `
			SELECT id, group_id, name, description, is_archived, created_at, updated_at, updated_by, version
			FROM lists
			WHERE group_id = $1
			ORDER BY updated_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{groupID, limit, offset}
	} else {
		query = `
			SELECT id, group_id, name, description, is_archived, created_at, updated_at, updated_by, version
			FROM lists
			WHERE group_id = $1 AND is_archived = false
			ORDER BY updated_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{groupID, limit, offset}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lists []*list.List
	for rows.Next() {
		var l list.List
		if err := rows.Scan(
			&l.ID,
			&l.GroupID,
			&l.Name,
			&l.Description,
			&l.IsArchived,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.UpdatedBy,
			&l.Version,
		); err != nil {
			return nil, 0, err
		}
		lists = append(lists, &l)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Get total count
	totalCount, err := r.CountByGroupID(ctx, groupID, includeArchived)
	if err != nil {
		return nil, 0, err
	}

	return lists, totalCount, nil
}

// Update updates an existing list with version check
func (r *ListRepository) Update(ctx context.Context, l *list.List, expectedVersion int64) error {
	query := `
		UPDATE lists
		SET name = $2, description = $3, updated_at = $4, updated_by = $5, version = version + 1
		WHERE id = $1 AND version = $6
	`

	result, err := r.pool.Exec(ctx, query,
		l.ID,
		l.Name,
		l.Description,
		time.Now(),
		l.UpdatedBy,
		expectedVersion,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		// Check if list exists
		_, err := r.GetByID(ctx, l.ID)
		if err != nil {
			return err
		}
		return list.ErrVersionMismatch
	}

	return nil
}

// Archive sets the archived status of a list
func (r *ListRepository) Archive(ctx context.Context, id uuid.UUID, archive bool, updatedBy uuid.UUID) (*list.List, error) {
	query := `
		UPDATE lists
		SET is_archived = $2, updated_at = $3, updated_by = $4, version = version + 1
		WHERE id = $1
		RETURNING id, group_id, name, description, is_archived, created_at, updated_at, updated_by, version
	`

	var l list.List
	err := r.pool.QueryRow(ctx, query, id, archive, time.Now(), updatedBy).Scan(
		&l.ID,
		&l.GroupID,
		&l.Name,
		&l.Description,
		&l.IsArchived,
		&l.CreatedAt,
		&l.UpdatedAt,
		&l.UpdatedBy,
		&l.Version,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, list.ErrListNotFound
		}
		return nil, err
	}

	return &l, nil
}

// CountByGroupID counts the total number of lists in a group
func (r *ListRepository) CountByGroupID(ctx context.Context, groupID uuid.UUID, includeArchived bool) (int, error) {
	var query string
	if includeArchived {
		query = `SELECT COUNT(*) FROM lists WHERE group_id = $1`
	} else {
		query = `SELECT COUNT(*) FROM lists WHERE group_id = $1 AND is_archived = false`
	}

	var count int
	err := r.pool.QueryRow(ctx, query, groupID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetGroupIDByListID retrieves the group ID for a list
func (r *ListRepository) GetGroupIDByListID(ctx context.Context, listID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT group_id FROM lists WHERE id = $1`

	var groupID uuid.UUID
	err := r.pool.QueryRow(ctx, query, listID).Scan(&groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, list.ErrListNotFound
		}
		return uuid.Nil, err
	}

	return groupID, nil
}

// ItemRepository implements list.ItemRepository with PostgreSQL
type ItemRepository struct {
	pool *pgxpool.Pool
}

// NewItemRepository creates a new PostgreSQL item repository
func NewItemRepository(db *DB) *ItemRepository {
	return &ItemRepository{pool: db.Pool}
}

// Create inserts a new item into the database
func (r *ItemRepository) Create(ctx context.Context, i *list.Item) error {
	query := `
		INSERT INTO items (id, list_id, category_id, name, priority, sort_order, is_purchased, quantity, notes, created_at, updated_at, updated_by, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.pool.Exec(ctx, query,
		i.ID,
		i.ListID,
		i.CategoryID,
		i.Name,
		string(i.Priority),
		i.SortOrder,
		i.IsPurchased,
		i.Quantity,
		i.Notes,
		i.CreatedAt,
		i.UpdatedAt,
		i.UpdatedBy,
		i.Version,
	)

	return err
}

// GetByID retrieves an item by its ID
func (r *ItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*list.Item, error) {
	query := `
		SELECT id, list_id, category_id, name, priority, sort_order, is_purchased, quantity, notes, created_at, updated_at, updated_by, version
		FROM items
		WHERE id = $1
	`

	return r.scanItem(ctx, query, id)
}

// ListByListID retrieves all items in a list with default sorting
func (r *ItemRepository) ListByListID(ctx context.Context, listID uuid.UUID) ([]*list.Item, error) {
	query := `
		SELECT id, list_id, category_id, name, priority, sort_order, is_purchased, quantity, notes, created_at, updated_at, updated_by, version
		FROM items
		WHERE list_id = $1
		ORDER BY is_purchased ASC, 
			CASE priority 
				WHEN 'URGENT' THEN 1 
				WHEN 'HIGH' THEN 2 
				WHEN 'MEDIUM' THEN 3 
				WHEN 'LOW' THEN 4 
				ELSE 5 
			END ASC, 
			sort_order ASC
	`

	rows, err := r.pool.Query(ctx, query, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*list.Item
	for rows.Next() {
		i, err := r.scanItemFromRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// Update updates an existing item with version check
func (r *ItemRepository) Update(ctx context.Context, i *list.Item, expectedVersion int64) error {
	query := `
		UPDATE items
		SET name = $2, priority = $3, category_id = $4, quantity = $5, notes = $6, is_purchased = $7, updated_at = $8, updated_by = $9, version = version + 1
		WHERE id = $1 AND version = $10
	`

	result, err := r.pool.Exec(ctx, query,
		i.ID,
		i.Name,
		string(i.Priority),
		i.CategoryID,
		i.Quantity,
		i.Notes,
		i.IsPurchased,
		time.Now(),
		i.UpdatedBy,
		expectedVersion,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		// Check if item exists
		_, err := r.GetByID(ctx, i.ID)
		if err != nil {
			return err
		}
		return list.ErrVersionMismatch
	}

	return nil
}

// Delete removes an item from the database
func (r *ItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM items WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return list.ErrItemNotFound
	}

	return nil
}

// GetMaxSortOrder returns the maximum sort order in a list
func (r *ItemRepository) GetMaxSortOrder(ctx context.Context, listID uuid.UUID) (int32, error) {
	query := `SELECT COALESCE(MAX(sort_order), 0) FROM items WHERE list_id = $1`

	var maxOrder int32
	err := r.pool.QueryRow(ctx, query, listID).Scan(&maxOrder)
	if err != nil {
		return 0, err
	}

	return maxOrder, nil
}

// ReorderItems updates sort_order for multiple items
func (r *ItemRepository) ReorderItems(ctx context.Context, listID uuid.UUID, itemIDs []uuid.UUID) error {
	// Update sort_order based on position in the array
	for i, itemID := range itemIDs {
		query := `UPDATE items SET sort_order = $1, updated_at = $2 WHERE id = $3 AND list_id = $4`
		_, err := r.pool.Exec(ctx, query, i, time.Now(), itemID, listID)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetListIDByItemID retrieves the list ID for an item
func (r *ItemRepository) GetListIDByItemID(ctx context.Context, itemID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT list_id FROM items WHERE id = $1`

	var listID uuid.UUID
	err := r.pool.QueryRow(ctx, query, itemID).Scan(&listID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, list.ErrItemNotFound
		}
		return uuid.Nil, err
	}

	return listID, nil
}

// scanItem is a helper to scan an item from a query result
func (r *ItemRepository) scanItem(ctx context.Context, query string, args ...interface{}) (*list.Item, error) {
	var i list.Item
	var priority string

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&i.ID,
		&i.ListID,
		&i.CategoryID,
		&i.Name,
		&priority,
		&i.SortOrder,
		&i.IsPurchased,
		&i.Quantity,
		&i.Notes,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UpdatedBy,
		&i.Version,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, list.ErrItemNotFound
		}
		return nil, err
	}

	i.Priority = list.Priority(priority)

	return &i, nil
}

// scanItemFromRow scans an item from a rows result
func (r *ItemRepository) scanItemFromRow(rows pgx.Rows) (*list.Item, error) {
	var i list.Item
	var priority string

	err := rows.Scan(
		&i.ID,
		&i.ListID,
		&i.CategoryID,
		&i.Name,
		&priority,
		&i.SortOrder,
		&i.IsPurchased,
		&i.Quantity,
		&i.Notes,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UpdatedBy,
		&i.Version,
	)

	if err != nil {
		return nil, err
	}

	i.Priority = list.Priority(priority)

	return &i, nil
}
