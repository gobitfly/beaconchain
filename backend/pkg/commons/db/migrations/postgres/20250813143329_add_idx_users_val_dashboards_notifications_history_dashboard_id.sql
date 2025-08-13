-- +goose NO TRANSACTION

-- +goose Up
SELECT 'create idx_users_val_dashboards_notifications_history_dashboard_id';
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_val_dashboards_notifications_history_dashboard_id ON users_val_dashboards_notifications_history (dashboard_id);
-- +goose StatementEnd

-- +goose Down
SELECT 'drop idx_users_val_dashboards_notifications_history_dashboard_id';
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_users_val_dashboards_notifications_history_dashboard_id;
-- +goose StatementEnd
