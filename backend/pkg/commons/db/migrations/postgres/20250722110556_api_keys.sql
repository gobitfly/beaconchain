-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE IF NOT EXISTS api_keys_v2(
    api_key_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id),
    api_key BYTEA NOT NULL UNIQUE,
    short_key TEXT NOT NULL CHECK (char_length(short_key) <= 10),
    name TEXT NOT NULL CHECK (char_length(name) <= 30),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP,
    disabled_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_api_keys_v2_name_active
ON api_keys_v2(user_id, name)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_api_keys_v2_api_key ON api_keys_v2(api_key);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_api_keys_v2_api_key;
DROP INDEX IF EXISTS uniq_api_keys_v2_name_active;
DROP TABLE IF EXISTS api_keys_v2;

-- +goose StatementEnd
