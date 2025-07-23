-- +goose Up
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    MATERIALIZE COLUMN _staging_balance_start,
    MATERIALIZE COLUMN  _staging_balance_min,
    MATERIALIZE COLUMN _staging_epoch_start
settings alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    MATERIALIZE COLUMN _staging_balance_start,
    MATERIALIZE COLUMN  _staging_balance_min,
    MATERIALIZE COLUMN _staging_epoch_start
settings alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    MATERIALIZE COLUMN _staging_balance_start,
    MATERIALIZE COLUMN  _staging_balance_min,
    MATERIALIZE COLUMN _staging_epoch_start
settings alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    MATERIALIZE COLUMN _staging_balance_start,
    MATERIALIZE COLUMN  _staging_balance_min,
    MATERIALIZE COLUMN _staging_epoch_start
settings alter_sync=1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
