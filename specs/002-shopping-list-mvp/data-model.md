# Data Model: Mobile Offline MVP

**Status**: Placeholder  
**Feature**: Mobile Offline MVP (Android + iOS)

## Overview

This document describes the local SQLite database schema used for offline-first data persistence in the mobile clients.

## Entity Relationship

```text
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│    Group    │ 1───* │    List     │ 1───* │    Item     │
└─────────────┘       └─────────────┘       └─────────────┘
                            │                     │
                            │                     │
                            ▼                     ▼
                      ┌─────────────┐       ┌─────────────┐
                      │  Category   │◄──────│  Category   │
                      └─────────────┘       └─────────────┘

┌─────────────┐       ┌─────────────┐
│ SyncState   │       │MutationQueue│
└─────────────┘       └─────────────┘
```

## Tables

### groups

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | TEXT | PRIMARY KEY | UUID of the group |
| name | TEXT | NOT NULL | Display name |
| created_at | INTEGER | NOT NULL | Unix timestamp |
| updated_at | INTEGER | NOT NULL | Unix timestamp |
| version | INTEGER | NOT NULL | Optimistic lock version |

### lists

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | TEXT | PRIMARY KEY | UUID of the list |
| group_id | TEXT | NOT NULL, FK | Parent group |
| name | TEXT | NOT NULL | Display name |
| category_id | TEXT | NULLABLE, FK | Optional category |
| created_at | INTEGER | NOT NULL | Unix timestamp |
| updated_at | INTEGER | NOT NULL | Unix timestamp |
| version | INTEGER | NOT NULL | Optimistic lock version |

### items

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | TEXT | PRIMARY KEY | UUID of the item |
| list_id | TEXT | NOT NULL, FK | Parent list |
| name | TEXT | NOT NULL | Item name |
| is_purchased | INTEGER | NOT NULL | 0 or 1 |
| priority | INTEGER | NOT NULL | Higher = more important |
| sort_order | INTEGER | NOT NULL | Explicit ordering |
| category_id | TEXT | NULLABLE, FK | Optional category |
| created_at | INTEGER | NOT NULL | Unix timestamp |
| updated_at | INTEGER | NOT NULL | Unix timestamp |
| version | INTEGER | NOT NULL | Optimistic lock version |
| pending_mutation_id | TEXT | NULLABLE | Links to queued mutation |

### categories

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | TEXT | PRIMARY KEY | UUID of the category |
| group_id | TEXT | NOT NULL, FK | Parent group |
| name | TEXT | NOT NULL | Category name |
| color | TEXT | NULLABLE | Hex color code |
| created_at | INTEGER | NOT NULL | Unix timestamp |
| updated_at | INTEGER | NOT NULL | Unix timestamp |
| version | INTEGER | NOT NULL | Optimistic lock version |

### mutation_queue

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | TEXT | PRIMARY KEY | Mutation UUID (idempotency key) |
| batch_id | TEXT | NULLABLE | Groups related mutations |
| entity_type | TEXT | NOT NULL | "list", "item", "category" |
| entity_id | TEXT | NOT NULL | UUID of affected entity |
| mutation_type | TEXT | NOT NULL | "CREATE", "UPDATE", "DELETE" |
| payload | BLOB | NOT NULL | JSON-encoded mutation data |
| expected_version | INTEGER | NULLABLE | For optimistic locking |
| status | TEXT | NOT NULL | "pending", "sending", "failed" |
| retry_count | INTEGER | NOT NULL | Number of retry attempts |
| created_at | INTEGER | NOT NULL | Unix timestamp |
| claimed_at | INTEGER | NULLABLE | When picked up for sending |

### sync_state

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| group_id | TEXT | PRIMARY KEY | Group this state applies to |
| cursor | INTEGER | NOT NULL | Last sync cursor (0 = initial) |
| last_sync_at | INTEGER | NULLABLE | Unix timestamp of last success |
| last_error | TEXT | NULLABLE | Last error message if any |

## Indices

```sql
CREATE INDEX idx_lists_group_id ON lists(group_id);
CREATE INDEX idx_items_list_id ON items(list_id);
CREATE INDEX idx_items_sort ON items(list_id, is_purchased, priority DESC, sort_order);
CREATE INDEX idx_mutation_queue_status ON mutation_queue(status, created_at);
CREATE INDEX idx_mutation_queue_batch ON mutation_queue(batch_id);
```

## Notes

- All timestamps are stored as Unix milliseconds (INTEGER)
- Boolean values are stored as INTEGER (0/1)
- UUIDs are stored as TEXT in canonical format
- TODO: Add migration history
- TODO: Add data validation constraints
