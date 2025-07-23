-- +goose Up

-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    RENAME COLUMN epoch_start TO _legacy_epoch_start,
    RENAME COLUMN balance_start TO _legacy_balance_start,
    RENAME COLUMN balance_min TO _legacy_balance_min,
    RENAME COLUMN _public_epoch_start TO epoch_start,
    RENAME COLUMN _public_balance_start TO balance_start,
    RENAME COLUMN _public_balance_min TO balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    RENAME COLUMN epoch_start TO _legacy_epoch_start,
    RENAME COLUMN balance_start TO _legacy_balance_start,
    RENAME COLUMN balance_min TO _legacy_balance_min,
    RENAME COLUMN _public_epoch_start TO epoch_start,
    RENAME COLUMN _public_balance_start TO balance_start,
    RENAME COLUMN _public_balance_min TO balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    RENAME COLUMN epoch_start TO _legacy_epoch_start,
    RENAME COLUMN balance_start TO _legacy_balance_start,
    RENAME COLUMN balance_min TO _legacy_balance_min,
    RENAME COLUMN _public_epoch_start TO epoch_start,
    RENAME COLUMN _public_balance_start TO balance_start,
    RENAME COLUMN _public_balance_min TO balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    RENAME COLUMN epoch_start TO _legacy_epoch_start,
    RENAME COLUMN balance_start TO _legacy_balance_start,
    RENAME COLUMN balance_min TO _legacy_balance_min,
    RENAME COLUMN _public_epoch_start TO epoch_start,
    RENAME COLUMN _public_balance_start TO balance_start,
    RENAME COLUMN _public_balance_min TO balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    RENAME COLUMN _legacy_epoch_start TO epoch_start,
    RENAME COLUMN _legacy_balance_start TO balance_start,
    RENAME COLUMN _legacy_balance_min TO balance_min,
    RENAME COLUMN epoch_start TO _public_epoch_start,
    RENAME COLUMN balance_start TO _public_balance_start,
    RENAME COLUMN balance_min TO _public_balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    RENAME COLUMN _legacy_epoch_start TO epoch_start,
    RENAME COLUMN _legacy_balance_start TO balance_start,
    RENAME COLUMN _legacy_balance_min TO balance_min,
    RENAME COLUMN epoch_start TO _public_epoch_start,
    RENAME COLUMN balance_start TO _public_balance_start,
    RENAME COLUMN balance_min TO _public_balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    RENAME COLUMN _legacy_epoch_start TO epoch_start,
    RENAME COLUMN _legacy_balance_start TO balance_start,
    RENAME COLUMN _legacy_balance_min TO balance_min,
    RENAME COLUMN epoch_start TO _public_epoch_start,
    RENAME COLUMN balance_start TO _public_balance_start,
    RENAME COLUMN balance_min TO _public_balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    RENAME COLUMN _legacy_epoch_start TO epoch_start,
    RENAME COLUMN _legacy_balance_start TO balance_start,
    RENAME COLUMN _legacy_balance_min TO balance_min,
    RENAME COLUMN epoch_start TO _public_epoch_start,
    RENAME COLUMN balance_start TO _public_balance_start,
    RENAME COLUMN balance_min TO _public_balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
