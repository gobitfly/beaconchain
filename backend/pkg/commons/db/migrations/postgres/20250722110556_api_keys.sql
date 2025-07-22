-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE user_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id),
    api_key BYTEA NOT NULL,
    short_key TEXT NOT NULL CHECK (char_length(short_key) <= 10),
    name TEXT NOT NULL CHECK (char_length(name) <= 30),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP,
    disabled_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX uniq_user_api_keys_name_active
ON user_api_keys(user_id, name)
WHERE deleted_at IS NULL;

CREATE INDEX idx_user_api_keys_api_key ON user_api_keys(api_key);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_user_api_keys_api_key;
DROP INDEX IF EXISTS uniq_user_api_keys_name_active;
DROP TABLE IF EXISTS user_api_keys;

-- +goose StatementEnd
