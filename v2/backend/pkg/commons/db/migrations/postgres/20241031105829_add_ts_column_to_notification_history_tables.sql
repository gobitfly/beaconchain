-- +goose Up
-- +goose StatementBegin
SELECT 'add ts column to notification history tables';
ALTER TABLE users_val_dashboards_notifications_history ADD COLUMN IF NOT EXISTS ts TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_user_id_ts_dashboard_id_group_id_event_type ON users_val_dashboards_notifications_history (user_id, ts, dashboard_id, group_id, event_type);
ALTER TABLE machine_notifications_history ADD COLUMN IF NOT EXISTS ts TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_user_id_ts_machine_id_machine_name_event_type ON machine_notifications_history (user_id, ts, machine_id, machine_name, event_type);
ALTER TABLE client_notifications_history ADD COLUMN IF NOT EXISTS ts TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_user_id_ts_client ON client_notifications_history  (user_id, ts, client);
ALTER TABLE network_notifications_history ADD COLUMN IF NOT EXISTS ts TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_user_id_epoch_network_event_type ON network_notifications_history (user_id, epoch, network, event_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'remove ts column from notification history tables';
ALTER TABLE users_val_dashboards_notifications_history DROP COLUMN IF EXISTS ts;
DROP INDEX IF EXISTS idx_user_id_ts_dashboard_id_group_id_event_type;
ALTER TABLE machine_notifications_history DROP COLUMN IF EXISTS ts;
DROP INDEX IF EXISTS idx_user_id_ts_machine_id_machine_name_event_type;
ALTER TABLE client_notifications_history DROP COLUMN IF EXISTS ts;
DROP INDEX IF EXISTS idx_user_id_ts_client;
ALTER TABLE network_notifications_history DROP COLUMN IF EXISTS ts;
DROP INDEX IF EXISTS idx_user_id_epoch_network_event_type;
-- +goose StatementEnd
