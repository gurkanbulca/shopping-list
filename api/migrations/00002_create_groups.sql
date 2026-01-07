-- +goose Up
-- +goose StatementBegin
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    CONSTRAINT groups_name_not_empty CHECK (TRIM(name) != '')
);

CREATE INDEX idx_groups_owner_id ON groups(owner_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS groups;
-- +goose StatementEnd
