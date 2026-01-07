-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    CONSTRAINT categories_name_not_empty CHECK (TRIM(name) != ''),
    CONSTRAINT categories_group_name_unique UNIQUE (group_id, LOWER(name))
);

CREATE INDEX idx_categories_group_id ON categories(group_id);
CREATE INDEX idx_categories_group_name ON categories(group_id, name);
CREATE INDEX idx_categories_updated_at ON categories(updated_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
