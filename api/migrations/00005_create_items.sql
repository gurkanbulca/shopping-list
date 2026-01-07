-- +goose Up
-- +goose StatementBegin
CREATE TYPE item_priority AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'URGENT');

CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    name VARCHAR(200) NOT NULL,
    priority item_priority NOT NULL DEFAULT 'MEDIUM',
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_purchased BOOLEAN NOT NULL DEFAULT false,
    quantity VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    CONSTRAINT items_name_not_empty CHECK (TRIM(name) != ''),
    CONSTRAINT items_sort_order_non_negative CHECK (sort_order >= 0)
);

CREATE INDEX idx_items_list_id ON items(list_id);
CREATE INDEX idx_items_list_purchased_priority_sort ON items(list_id, is_purchased, priority DESC, sort_order ASC);
CREATE INDEX idx_items_category_id ON items(category_id);
CREATE INDEX idx_items_updated_at ON items(updated_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
DROP TYPE IF EXISTS item_priority;
-- +goose StatementEnd
