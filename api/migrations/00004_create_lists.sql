-- +goose Up
-- +goose StatementBegin
CREATE TABLE lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    is_archived BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    CONSTRAINT lists_name_not_empty CHECK (TRIM(name) != '')
);

CREATE INDEX idx_lists_group_id ON lists(group_id);
CREATE INDEX idx_lists_group_archived ON lists(group_id, is_archived);
CREATE INDEX idx_lists_updated_at ON lists(updated_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lists;
-- +goose StatementEnd
