-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _final_validator_dashboard_effective_balance_lookup
(
    `validator_index` UInt64 COMMENT 'validator index',
    `epoch_timestamp` DateTime COMMENT 'timestamp of the aggregated data',
    `epoch` UInt64 COMMENT 'epoch of the aggregated data',
    `balance_effective_end` Int64
)
ENGINE = ReplacingMergeTree(epoch_timestamp)
ORDER BY (validator_index)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048, mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_final_validator_dashboard_effective_balance_lookup TO _final_validator_dashboard_effective_balance_lookup
AS SELECT
    validator_index,
    epoch_timestamp,
    epoch, 
    balance_effective_end
FROM _final_validator_dashboard_data_epoch
WHERE attestations_scheduled > 0
-- +goose StatementEnd
-- +goose StatementBegin
CREATE VIEW IF NOT EXISTS validator_dashboard_effective_balance_lookup AS
SELECT
    validator_index,
    epoch_timestamp,
    epoch, 
    balance_effective_end
FROM _final_validator_dashboard_effective_balance_lookup FINAL
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP VIEW validator_dashboard_effective_balance_lookup
-- +goose StatementEnd
-- +goose StatementBegin
DROP MATERIALIZED VIEW _mv_final_validator_dashboard_effective_balance_lookup
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE _final_validator_dashboard_effective_balance_lookup
-- +goose StatementEnd
