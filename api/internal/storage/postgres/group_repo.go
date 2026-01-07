package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/group"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GroupRepository implements group.GroupRepository with PostgreSQL
type GroupRepository struct {
	pool *pgxpool.Pool
}

// NewGroupRepository creates a new PostgreSQL group repository
func NewGroupRepository(db *DB) *GroupRepository {
	return &GroupRepository{pool: db.Pool}
}

// Create inserts a new group into the database
func (r *GroupRepository) Create(ctx context.Context, g *group.Group) error {
	query := `
		INSERT INTO groups (id, name, description, owner_id, created_at, updated_at, updated_by, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		g.ID,
		g.Name,
		g.Description,
		g.OwnerID,
		g.CreatedAt,
		g.UpdatedAt,
		g.UpdatedBy,
		g.Version,
	)

	return err
}

// GetByID retrieves a group by its ID
func (r *GroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*group.Group, error) {
	query := `
		SELECT id, name, description, owner_id, created_at, updated_at, updated_by, version
		FROM groups
		WHERE id = $1
	`

	var g group.Group
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&g.ID,
		&g.Name,
		&g.Description,
		&g.OwnerID,
		&g.CreatedAt,
		&g.UpdatedAt,
		&g.UpdatedBy,
		&g.Version,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, group.ErrGroupNotFound
		}
		return nil, err
	}

	return &g, nil
}

// ListByUserID retrieves all groups a user is a member of
func (r *GroupRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*group.Group, int, error) {
	query := `
		SELECT g.id, g.name, g.description, g.owner_id, g.created_at, g.updated_at, g.updated_by, g.version
		FROM groups g
		INNER JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1 AND gm.status = 'ACTIVE'
		ORDER BY g.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var groups []*group.Group
	for rows.Next() {
		var g group.Group
		if err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.Description,
			&g.OwnerID,
			&g.CreatedAt,
			&g.UpdatedAt,
			&g.UpdatedBy,
			&g.Version,
		); err != nil {
			return nil, 0, err
		}
		groups = append(groups, &g)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Get total count
	totalCount, err := r.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return groups, totalCount, nil
}

// Update updates an existing group
func (r *GroupRepository) Update(ctx context.Context, g *group.Group) error {
	query := `
		UPDATE groups
		SET name = $2, description = $3, updated_at = $4, updated_by = $5, version = version + 1
		WHERE id = $1 AND version = $6
	`

	result, err := r.pool.Exec(ctx, query,
		g.ID,
		g.Name,
		g.Description,
		g.UpdatedAt,
		g.UpdatedBy,
		g.Version,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return group.ErrGroupNotFound
	}

	return nil
}

// CountByUserID counts the total number of groups a user is a member of
func (r *GroupRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM groups g
		INNER JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1 AND gm.status = 'ACTIVE'
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// MemberRepository implements group.MemberRepository with PostgreSQL
type MemberRepository struct {
	pool *pgxpool.Pool
}

// NewMemberRepository creates a new PostgreSQL member repository
func NewMemberRepository(db *DB) *MemberRepository {
	return &MemberRepository{pool: db.Pool}
}

// Create inserts a new group member
func (r *MemberRepository) Create(ctx context.Context, m *group.GroupMember) error {
	query := `
		INSERT INTO group_members (id, group_id, user_id, role, status, invited_by, invited_at, accepted_at, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.pool.Exec(ctx, query,
		m.ID,
		m.GroupID,
		m.UserID,
		string(m.Role),
		string(m.Status),
		m.InvitedBy,
		m.InvitedAt,
		m.AcceptedAt,
		m.CreatedAt,
		m.UpdatedAt,
		m.Version,
	)

	return err
}

// GetByGroupAndUser retrieves a member by group and user ID
func (r *MemberRepository) GetByGroupAndUser(ctx context.Context, groupID, userID uuid.UUID) (*group.GroupMember, error) {
	query := `
		SELECT id, group_id, user_id, role, status, invited_by, invited_at, accepted_at, created_at, updated_at, version
		FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`

	return r.scanMember(ctx, query, groupID, userID)
}

// ListByGroupID retrieves all members of a group
func (r *MemberRepository) ListByGroupID(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*group.GroupMember, int, error) {
	query := `
		SELECT id, group_id, user_id, role, status, invited_by, invited_at, accepted_at, created_at, updated_at, version
		FROM group_members
		WHERE group_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var members []*group.GroupMember
	for rows.Next() {
		m, err := r.scanMemberFromRow(rows)
		if err != nil {
			return nil, 0, err
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Get total count
	totalCount, err := r.CountByGroupID(ctx, groupID)
	if err != nil {
		return nil, 0, err
	}

	return members, totalCount, nil
}

// Update updates an existing group member
func (r *MemberRepository) Update(ctx context.Context, m *group.GroupMember) error {
	query := `
		UPDATE group_members
		SET role = $2, status = $3, accepted_at = $4, updated_at = $5, version = version + 1
		WHERE id = $1 AND version = $6
	`

	result, err := r.pool.Exec(ctx, query,
		m.ID,
		string(m.Role),
		string(m.Status),
		m.AcceptedAt,
		m.UpdatedAt,
		m.Version,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return group.ErrMemberNotFound
	}

	return nil
}

// Delete removes a member from a group
func (r *MemberRepository) Delete(ctx context.Context, groupID, userID uuid.UUID) error {
	query := `DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, groupID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return group.ErrMemberNotFound
	}

	return nil
}

// ExistsByGroupAndUser checks if a membership exists
func (r *MemberRepository) ExistsByGroupAndUser(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, groupID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetByGroupAndEmail retrieves a member by group and email
func (r *MemberRepository) GetByGroupAndEmail(ctx context.Context, groupID uuid.UUID, email string) (*group.GroupMember, error) {
	query := `
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.invited_by, gm.invited_at, gm.accepted_at, gm.created_at, gm.updated_at, gm.version
		FROM group_members gm
		INNER JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = $1 AND u.email = $2
	`

	return r.scanMember(ctx, query, groupID, email)
}

// GetByGroupAndPhone retrieves a member by group and phone
func (r *MemberRepository) GetByGroupAndPhone(ctx context.Context, groupID uuid.UUID, phone string) (*group.GroupMember, error) {
	query := `
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.invited_by, gm.invited_at, gm.accepted_at, gm.created_at, gm.updated_at, gm.version
		FROM group_members gm
		INNER JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = $1 AND u.phone = $2
	`

	return r.scanMember(ctx, query, groupID, phone)
}

// CountByGroupID counts the total number of members in a group
func (r *MemberRepository) CountByGroupID(ctx context.Context, groupID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM group_members WHERE group_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, groupID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// scanMember is a helper to scan a member from a query result
func (r *MemberRepository) scanMember(ctx context.Context, query string, args ...interface{}) (*group.GroupMember, error) {
	var m group.GroupMember
	var role, status string

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&m.ID,
		&m.GroupID,
		&m.UserID,
		&role,
		&status,
		&m.InvitedBy,
		&m.InvitedAt,
		&m.AcceptedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.Version,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, group.ErrMemberNotFound
		}
		return nil, err
	}

	m.Role = group.MemberRole(role)
	m.Status = group.MemberStatus(status)

	return &m, nil
}

// scanMemberFromRow scans a member from a rows result
func (r *MemberRepository) scanMemberFromRow(rows pgx.Rows) (*group.GroupMember, error) {
	var m group.GroupMember
	var role, status string

	err := rows.Scan(
		&m.ID,
		&m.GroupID,
		&m.UserID,
		&role,
		&status,
		&m.InvitedBy,
		&m.InvitedAt,
		&m.AcceptedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.Version,
	)

	if err != nil {
		return nil, err
	}

	m.Role = group.MemberRole(role)
	m.Status = group.MemberStatus(status)

	return &m, nil
}

// UserLookupRepository implements group.UserLookupRepository with PostgreSQL
type UserLookupRepository struct {
	pool *pgxpool.Pool
}

// NewUserLookupRepository creates a new PostgreSQL user lookup repository
func NewUserLookupRepository(db *DB) *UserLookupRepository {
	return &UserLookupRepository{pool: db.Pool}
}

// GetUserIDByEmail retrieves a user's ID by email
func (r *UserLookupRepository) GetUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	query := `SELECT id FROM users WHERE email = $1`

	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, email).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, group.ErrUserNotFound
		}
		return uuid.Nil, err
	}

	return id, nil
}

// GetUserIDByPhone retrieves a user's ID by phone
func (r *UserLookupRepository) GetUserIDByPhone(ctx context.Context, phone string) (uuid.UUID, error) {
	query := `SELECT id FROM users WHERE phone = $1`

	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, phone).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, group.ErrUserNotFound
		}
		return uuid.Nil, err
	}

	return id, nil
}
