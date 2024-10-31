-- +goose Up
-- +goose StatementBegin
SELECT 'add ts column to notification history tables';
ALTER TABLE users_val_dashboards_notifications_history ADD COLUMN IF NOT EXISTS ts TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE machine_notifications_history ADD COLUMN IF NOT EXISTS ts TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE client_notifications_history ADD COLUMN IF NOT EXISTS ts TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE network_notifications_history ADD COLUMN IF NOT EXISTS ts TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'remove ts column from notification history tables';
ALTER TABLE users_val_dashboards_notifications_history DROP COLUMN IF EXISTS ts;
ALTER TABLE machine_notifications_history DROP COLUMN IF EXISTS ts;
ALTER TABLE client_notifications_history DROP COLUMN IF EXISTS ts;
ALTER TABLE network_notifications_history DROP COLUMN IF EXISTS ts;
-- +goose StatementEnd
