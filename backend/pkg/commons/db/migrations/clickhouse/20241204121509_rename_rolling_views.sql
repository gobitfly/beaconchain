-- +goose Up

-- +goose StatementBegin
-- +goose StatementEnd
-- +goose StatementBegin
RENAME TABLE validator_dashboard_rolling_1h TO validator_dashboard_data_rolling_1h
-- +goose StatementEnd
-- +goose StatementBegin
RENAME TABLE validator_dashboard_rolling_24h TO validator_dashboard_data_rolling_24h
-- +goose StatementEnd
-- +goose StatementBegin
RENAME TABLE validator_dashboard_rolling_7d TO validator_dashboard_data_rolling_7d
-- +goose StatementEnd
-- +goose StatementBegin
RENAME TABLE validator_dashboard_rolling_30d TO validator_dashboard_data_rolling_30d
-- +goose StatementEnd
-- +goose StatementBegin
RENAME TABLE validator_dashboard_rolling_90d TO validator_dashboard_data_rolling_90d
-- +goose StatementEnd
-- +goose StatementBegin
RENAME TABLE validator_dashboard_rolling_total TO validator_dashboard_data_rolling_total
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
RENAME TABLE validator_dashboard_data_rolling_1h TO validator_dashboard_rolling_1h
-- +goose StatementEnd
-- +goose StatementBegin
RENAME TABLE validator_dashboard_data_rolling_24h TO validator_dashboard_rolling_24h
-- +goose StatementEnd
-- +goose StatementBegin
RENAME TABLE validator_dashboard_data_rolling_7d TO validator_dashboard_rolling_7d
-- +goose StatementEnd
-- +goose StatementBegin
RENAME TABLE validator_dashboard_data_rolling_30d TO validator_dashboard_rolling_30d
-- +goose StatementEnd
-- +goose StatementBegin
RENAME TABLE validator_dashboard_data_rolling_90d TO validator_dashboard_rolling_90d
-- +goose StatementEnd
-- +goose StatementBegin
RENAME TABLE validator_dashboard_data_rolling_total TO validator_dashboard_rolling_total
-- +goose StatementEnd

