-- +goose Up
-- +goose StatementBegin
CREATE TYPE member_role AS ENUM ('OWNER', 'ADMIN', 'MEMBER');
CREATE TYPE member_status AS ENUM ('ACTIVE', 'INVITED');

CREATE TABLE group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role member_role NOT NULL,
    status member_status NOT NULL DEFAULT 'INVITED',
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP,
    accepted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 1,
    CONSTRAINT group_members_unique UNIQUE (group_id, user_id)
);

CREATE INDEX idx_group_members_group_user ON group_members(group_id, user_id);
CREATE INDEX idx_group_members_user_id ON group_members(user_id);
CREATE INDEX idx_group_members_status ON group_members(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS group_members;
DROP TYPE IF EXISTS member_status;
DROP TYPE IF EXISTS member_role;
-- +goose StatementEnd
