-- +goose Up
-- +goose StatementBegin
ALTER TABLE users_val_dashboards ALTER COLUMN network TYPE BIGINT
ALTER TABLE network_notifications_history ALTER COLUMN network TYPE BIGINT
-- +goose StatementEnd

-- +goose Down
SELECT 'down SQL query - we do not revert the network id to SMALLINT as this could cause an out of range error';
-- +goose StatementBegin
-- +goose StatementEnd
