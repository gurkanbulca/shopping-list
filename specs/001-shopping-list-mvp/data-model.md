# Data Model: Shopping List SaaS MVP

**Created**: 2026-01-07  
**Feature**: Shopping List SaaS MVP

## Overview

The data model implements a multi-tenant shopping list system where Groups serve as tenant boundaries. All resources (Lists, Items, Categories) belong to Groups, and access is controlled through GroupMembership.

## Core Entities

### User

Represents a registered user account.

**Fields**:
- `id` (UUID, PRIMARY KEY)
- `email` (STRING, UNIQUE, NOT NULL) - or phone number
- `phone` (STRING, UNIQUE, NULLABLE) - alternative to email
- `password_hash` (STRING, NOT NULL) - bcrypt hash
- `name` (STRING, NULLABLE)
- `created_at` (TIMESTAMP, NOT NULL)
- `updated_at` (TIMESTAMP, NOT NULL)
- `version` (BIGINT, NOT NULL, DEFAULT 1) - for optimistic locking

**Constraints**:
- Either email or phone must be provided (not both null)
- Email must be valid format if provided
- Password hash uses bcrypt with cost 10+

**Relationships**:
- One-to-many with GroupMember (user can belong to multiple groups)

**Indexes**:
- `idx_users_email` on `email` (UNIQUE)
- `idx_users_phone` on `phone` (UNIQUE, partial where phone IS NOT NULL)

### Group

Represents a tenant boundary - a collection of users who share lists and items.

**Fields**:
- `id` (UUID, PRIMARY KEY)
- `name` (STRING, NOT NULL)
- `description` (STRING, NULLABLE)
- `owner_id` (UUID, NOT NULL, FOREIGN KEY → users.id)
- `created_at` (TIMESTAMP, NOT NULL)
- `updated_at` (TIMESTAMP, NOT NULL)
- `updated_by` (UUID, NULLABLE, FOREIGN KEY → users.id)
- `version` (BIGINT, NOT NULL, DEFAULT 1)

**Constraints**:
- Name must be non-empty (trimmed length > 0)
- Owner must exist in users table

**Relationships**:
- Many-to-one with User (owner)
- One-to-many with GroupMember
- One-to-many with List
- One-to-many with Category

**Indexes**:
- `idx_groups_owner_id` on `owner_id`

### GroupMember

Represents a user's membership in a group with a role.

**Fields**:
- `id` (UUID, PRIMARY KEY)
- `group_id` (UUID, NOT NULL, FOREIGN KEY → groups.id)
- `user_id` (UUID, NOT NULL, FOREIGN KEY → users.id)
- `role` (ENUM, NOT NULL) - OWNER, ADMIN, MEMBER
- `status` (ENUM, NOT NULL) - ACTIVE, INVITED
- `invited_by` (UUID, NULLABLE, FOREIGN KEY → users.id)
- `invited_at` (TIMESTAMP, NULLABLE)
- `accepted_at` (TIMESTAMP, NULLABLE)
- `created_at` (TIMESTAMP, NOT NULL)
- `updated_at` (TIMESTAMP, NOT NULL)
- `version` (BIGINT, NOT NULL, DEFAULT 1)

**Constraints**:
- Unique constraint on (group_id, user_id) - user can only be member once
- Role must be valid enum value
- Status must be valid enum value
- If status is INVITED, invited_by and invited_at must be set
- If status is ACTIVE, accepted_at must be set
- Owner role can only be assigned to one member per group

**Relationships**:
- Many-to-one with Group
- Many-to-one with User (member)
- Many-to-one with User (inviter)

**Indexes**:
- `idx_group_members_group_user` on `(group_id, user_id)` (UNIQUE)
- `idx_group_members_user_id` on `user_id`
- `idx_group_members_status` on `status` (for filtering invitations)

### List

Represents a shopping list within a group.

**Fields**:
- `id` (UUID, PRIMARY KEY)
- `group_id` (UUID, NOT NULL, FOREIGN KEY → groups.id)
- `name` (STRING, NOT NULL)
- `description` (STRING, NULLABLE)
- `is_archived` (BOOLEAN, NOT NULL, DEFAULT false)
- `created_at` (TIMESTAMP, NOT NULL)
- `updated_at` (TIMESTAMP, NOT NULL)
- `updated_by` (UUID, NULLABLE, FOREIGN KEY → users.id)
- `version` (BIGINT, NOT NULL, DEFAULT 1)

**Constraints**:
- Name must be non-empty (trimmed length > 0)
- Group must exist

**Relationships**:
- Many-to-one with Group
- One-to-many with Item

**Indexes**:
- `idx_lists_group_id` on `group_id`
- `idx_lists_group_archived` on `(group_id, is_archived)` (for filtering active lists)
- `idx_lists_updated_at` on `updated_at` (for sync cursor)

### Item

Represents a product/item in a shopping list.

**Fields**:
- `id` (UUID, PRIMARY KEY)
- `list_id` (UUID, NOT NULL, FOREIGN KEY → lists.id)
- `category_id` (UUID, NULLABLE, FOREIGN KEY → categories.id)
- `name` (STRING, NOT NULL)
- `priority` (ENUM, NOT NULL, DEFAULT MEDIUM) - LOW, MEDIUM, HIGH, URGENT
- `sort_order` (INTEGER, NOT NULL, DEFAULT 0) - manual ordering
- `is_purchased` (BOOLEAN, NOT NULL, DEFAULT false)
- `quantity` (STRING, NULLABLE) - e.g., "2", "1kg", "500ml"
- `notes` (TEXT, NULLABLE)
- `created_at` (TIMESTAMP, NOT NULL)
- `updated_at` (TIMESTAMP, NOT NULL)
- `updated_by` (UUID, NULLABLE, FOREIGN KEY → users.id)
- `version` (BIGINT, NOT NULL, DEFAULT 1)

**Constraints**:
- Name must be non-empty (trimmed length > 0)
- Priority must be valid enum value
- List must exist
- Category can be NULL (soft reference - category may be deleted)
- Sort order must be >= 0

**Relationships**:
- Many-to-one with List
- Many-to-one with Category (nullable, soft reference)

**Indexes**:
- `idx_items_list_id` on `list_id`
- `idx_items_list_purchased_priority_sort` on `(list_id, is_purchased, priority DESC, sort_order ASC)` (for default sorting)
- `idx_items_category_id` on `category_id` (nullable, for filtering)
- `idx_items_updated_at` on `updated_at` (for sync cursor)

**Sorting Rules** (enforced in queries, not DB):
1. `is_purchased = false` first
2. Then `priority DESC` (URGENT > HIGH > MEDIUM > LOW)
3. Then `sort_order ASC`

### Category

Represents an organizational category within a group.

**Fields**:
- `id` (UUID, PRIMARY KEY)
- `group_id` (UUID, NOT NULL, FOREIGN KEY → groups.id)
- `name` (STRING, NOT NULL)
- `created_at` (TIMESTAMP, NOT NULL)
- `updated_at` (TIMESTAMP, NOT NULL)
- `updated_by` (UUID, NULLABLE, FOREIGN KEY → users.id)
- `version` (BIGINT, NOT NULL, DEFAULT 1)

**Constraints**:
- Name must be non-empty (trimmed length > 0)
- Unique constraint on (group_id, name) - no duplicate names per group
- Group must exist

**Relationships**:
- Many-to-one with Group
- One-to-many with Item (soft reference - items retain category_id even if category deleted)

**Indexes**:
- `idx_categories_group_id` on `group_id`
- `idx_categories_group_name` on `(group_id, name)` (UNIQUE)
- `idx_categories_updated_at` on `updated_at` (for sync cursor)

### Invite (Optional)

Represents a pending group invitation.

**Fields**:
- `id` (UUID, PRIMARY KEY)
- `group_id` (UUID, NOT NULL, FOREIGN KEY → groups.id)
- `email` (STRING, NOT NULL) - or phone
- `invited_by` (UUID, NOT NULL, FOREIGN KEY → users.id)
- `role` (ENUM, NOT NULL) - ADMIN, MEMBER (OWNER not allowed)
- `status` (ENUM, NOT NULL) - PENDING, ACCEPTED, DECLINED, EXPIRED
- `expires_at` (TIMESTAMP, NOT NULL)
- `created_at` (TIMESTAMP, NOT NULL)
- `updated_at` (TIMESTAMP, NOT NULL)

**Constraints**:
- Expires at must be in future when created
- Status must be valid enum value
- Role cannot be OWNER

**Relationships**:
- Many-to-one with Group
- Many-to-one with User (inviter)

**Indexes**:
- `idx_invites_group_email` on `(group_id, email)`
- `idx_invites_status` on `status`
- `idx_invites_expires_at` on `expires_at` (for cleanup)

**Note**: This table is optional - invitations can also be stored as GroupMember records with status=INVITED. Separate table provides better query performance and expiration handling.

### MutationLog (Optional, for Sync)

Tracks all changes for delta sync with global sequence.

**Fields**:
- `id` (UUID, PRIMARY KEY)
- `sequence` (BIGSERIAL, UNIQUE, NOT NULL) - global change sequence
- `entity_type` (ENUM, NOT NULL) - USER, GROUP, GROUP_MEMBER, LIST, ITEM, CATEGORY
- `entity_id` (UUID, NOT NULL)
- `operation` (ENUM, NOT NULL) - CREATE, UPDATE, DELETE
- `group_id` (UUID, NULLABLE) - for filtering by tenant
- `changed_at` (TIMESTAMP, NOT NULL, DEFAULT NOW())
- `changed_by` (UUID, NULLABLE, FOREIGN KEY → users.id)

**Constraints**:
- Entity type must be valid enum value
- Operation must be valid enum value
- Sequence is auto-incrementing (never reused)

**Relationships**:
- Many-to-one with Group (nullable, for filtering)

**Indexes**:
- `idx_mutation_log_sequence` on `sequence` (UNIQUE, for cursor)
- `idx_mutation_log_group_changed` on `(group_id, changed_at)` (for group-specific delta)
- `idx_mutation_log_entity` on `(entity_type, entity_id)` (for entity lookup)

**Note**: This table enables efficient delta sync. If not used, sync relies on updated_at cursors per table (less reliable but simpler).

## State Transitions

### GroupMember Status

```
INVITED → ACCEPTED (when user accepts invitation)
INVITED → EXPIRED (when expires_at passes)
INVITED → DECLINED (when user declines)
```

### List Archive Status

```
is_archived = false → is_archived = true (when archived)
is_archived = true → is_archived = false (when unarchived)
```

### Item Purchase Status

```
is_purchased = false → is_purchased = true (when marked purchased)
is_purchased = true → is_purchased = false (when unmarked)
```

## Validation Rules

### User
- Email format validation (RFC 5322 compliant)
- Phone format validation (E.164 format)
- Password strength: minimum 8 characters, at least one letter and one number

### Group
- Name: 1-100 characters
- Description: max 500 characters

### List
- Name: 1-100 characters
- Description: max 500 characters

### Item
- Name: 1-200 characters
- Quantity: max 50 characters
- Notes: max 1000 characters

### Category
- Name: 1-50 characters
- Unique per group (case-insensitive)

## Multi-Tenant Enforcement

All queries MUST include group membership check:

```sql
-- Example: Get list only if user is group member
SELECT l.* FROM lists l
INNER JOIN group_members gm ON l.group_id = gm.group_id
WHERE l.id = $1 AND gm.user_id = $2 AND gm.status = 'ACTIVE';
```

Authorization interceptor validates membership before any group-scoped operation.

## Versioning & Conflict Resolution

All entities include `version` (BIGINT) for optimistic locking:

1. Client reads entity with version N
2. Client modifies entity locally
3. Client sends update with `expected_version = N`
4. Server checks: if current version != N, return `FailedPrecondition`
5. If match, increment version and update

This ensures Last-Write-Wins with version check (more deterministic than timestamp-only).

## Soft Deletes

Categories use soft delete pattern:
- When category deleted, set `deleted_at` timestamp
- Items retain `category_id` reference (becomes invalid/unassigned)
- Queries filter out deleted categories: `WHERE deleted_at IS NULL`

Alternatively, hard delete with cascade NULL on items.category_id.

## Audit Trail

All entities track:
- `created_at` - when created
- `updated_at` - when last modified
- `updated_by` - who last modified (nullable, for system operations)

MutationLog (if used) provides complete audit trail of all changes.

## Migration Strategy

1. Create tables in dependency order:
   - users
   - groups (depends on users)
   - group_members (depends on users, groups)
   - categories (depends on groups)
   - lists (depends on groups)
   - items (depends on lists, categories)
   - invites (optional, depends on users, groups)
   - mutation_log (optional, depends on groups, users)

2. Add indexes after table creation

3. Add foreign key constraints after all tables exist

4. Seed initial data if needed (e.g., default categories)

## Performance Considerations

- Indexes on foreign keys for join performance
- Composite indexes for common query patterns (list items sorted, group members)
- Partial indexes where applicable (active members, non-archived lists)
- Consider partitioning mutation_log by date if it grows large

## Future Enhancements (Post-MVP)

- Full-text search on item names
- Item tags (many-to-many relationship)
- List templates
- Recurring items
- Item images/attachments
- Advanced filtering and sorting options
