-- +goose Up
-- +goose StatementBegin
SELECT 'add columns to table users_val_dashboards_groups';
ALTER TABLE users_val_dashboards_groups ADD COLUMN IF NOT EXISTS webhook_last_sent TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE users_val_dashboards_groups ADD COLUMN IF NOT EXISTS webhook_retries INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'remove columns from table users_val_dashboards_groups';
ALTER TABLE users_val_dashboards_groups DROP COLUMN IF EXISTS webhook_last_sent;
ALTER TABLE users_val_dashboards_groups DROP COLUMN IF EXISTS webhook_retries;
-- +goose StatementEnd
