-- +goose Up
-- +goose StatementBegin
RENAME validator_dashboard_data_epoch_max_ts TO view_validator_dashboard_data_epoch_max_ts
-- +goose StatementEnd
-- +goose StatementBegin
RENAME validator_dashboard_data_hourly_max_ts TO view_validator_dashboard_data_hourly_max_ts
-- +goose StatementEnd
-- +goose StatementBegin
RENAME validator_dashboard_data_daily_max_ts TO view_validator_dashboard_data_daily_max_ts
-- +goose StatementEnd
-- +goose StatementBegin
RENAME validator_dashboard_data_weekly_max_ts TO view_validator_dashboard_data_weekly_max_ts
-- +goose StatementEnd
-- +goose StatementBegin
RENAME validator_dashboard_data_monthly_max_ts TO view_validator_dashboard_data_monthly_max_ts
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
RENAME view_validator_dashboard_data_epoch_max_ts TO validator_dashboard_data_epoch_max_ts
-- +goose StatementEnd
-- +goose StatementBegin
RENAME view_validator_dashboard_data_hourly_max_ts TO validator_dashboard_data_hourly_max_ts
-- +goose StatementEnd
-- +goose StatementBegin
RENAME view_validator_dashboard_data_daily_max_ts TO validator_dashboard_data_daily_max_ts
-- +goose StatementEnd
-- +goose StatementBegin
RENAME view_validator_dashboard_data_weekly_max_ts TO validator_dashboard_data_weekly_max_ts
-- +goose StatementEnd
-- +goose StatementBegin
RENAME view_validator_dashboard_data_monthly_max_ts TO validator_dashboard_data_monthly_max_ts
-- +goose StatementEnd