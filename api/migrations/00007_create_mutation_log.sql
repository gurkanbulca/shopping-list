-- +goose Up
-- +goose StatementBegin

-- Entity type enum for mutation tracking
CREATE TYPE entity_type AS ENUM ('USER', 'GROUP', 'GROUP_MEMBER', 'LIST', 'ITEM', 'CATEGORY');

-- Change operation enum
CREATE TYPE change_operation AS ENUM ('CREATE', 'UPDATE', 'DELETE');

-- Mutation log table for delta sync
CREATE TABLE mutation_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence BIGSERIAL UNIQUE NOT NULL,
    entity_type entity_type NOT NULL,
    entity_id UUID NOT NULL,
    operation change_operation NOT NULL,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    changed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    changed_by UUID REFERENCES users(id)
);

-- Index for cursor-based delta sync
CREATE INDEX idx_mutation_log_sequence ON mutation_log(sequence);

-- Index for group-specific delta queries
CREATE INDEX idx_mutation_log_group_sequence ON mutation_log(group_id, sequence);

-- Index for entity lookup (deduplication)
CREATE INDEX idx_mutation_log_entity ON mutation_log(entity_type, entity_id);

-- Idempotency key table for tracking processed mutations
CREATE TABLE processed_mutations (
    mutation_id UUID PRIMARY KEY,
    entity_id UUID NOT NULL,
    version BIGINT NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for cleanup of old processed mutations
CREATE INDEX idx_processed_mutations_processed_at ON processed_mutations(processed_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS processed_mutations;
DROP TABLE IF EXISTS mutation_log;
DROP TYPE IF EXISTS change_operation;
DROP TYPE IF EXISTS entity_type;
-- +goose StatementEnd
