-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add column notifications_do_not_disturb_ts and change event_threshold data type';
ALTER TABLE users ADD COLUMN IF NOT EXISTS notifications_do_not_disturb_ts TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE users_subscriptions ALTER COLUMN event_threshold SET DATA TYPE NUMERIC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - remove column notifications_do_not_disturb_ts and revert event_threshold data type';
ALTER TABLE users DROP COLUMN IF EXISTS notifications_do_not_disturb_ts;
ALTER TABLE users_subscriptions ALTER COLUMN event_threshold SET DATA TYPE REAL;
-- +goose StatementEnd
