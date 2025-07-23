-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _exporter_aggregate_backfill_metadata
(
    `backfill_name` String COMMENT 'name of the backfill',
    `aggregation` String COMMENT 'aggregation type (hourly, daily, weekly, monthly)',
    `t` DateTime COMMENT 'timestamp of the aggregation',
    `max_epoch_timestamp` DateTime COMMENT 'maximum epoch timestamp to be used for the aggregation during backfill',
    `did_check` Nullable(DateTime) COMMENT 'timestamp when the check was performed',
    `backfill_id` Nullable(UUID) COMMENT 'id of the backfill',
    `successful_backfill` Nullable(DateTime) COMMENT 'if the backfill was successful',
    -- check that backfill_id is set only if did_check is not null
    CONSTRAINT check_backfill_id CHECK (backfill_id IS NULL OR did_check IS NOT NULL),
    -- check that successful_backfill is set only if backfill_id is not null
    CONSTRAINT check_successful_backfill CHECK (successful_backfill IS NULL OR backfill_id IS NOT NULL)
)
ENGINE = ReplacingMergeTree
ORDER BY (backfill_name, aggregation, t)
SETTINGS non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048;
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO _exporter_aggregate_backfill_metadata (backfill_name, aggregation, t, max_epoch_timestamp) 
select 'botched_epoch_start', '_final_validator_dashboard_data_hourly', t, (select max(epoch_timestamp) from _final_validator_dashboard_data_epoch) from _final_validator_dashboard_data_hourly PREWHERE validator_index = 0 limit 1 by t
UNION ALL
select 'botched_epoch_start', '_final_validator_dashboard_data_daily', t, (select max(epoch_timestamp) from _final_validator_dashboard_data_epoch) from _final_validator_dashboard_data_daily PREWHERE validator_index = 0 limit 1 by t
UNION ALL
select 'botched_epoch_start', '_final_validator_dashboard_data_weekly', t, (select max(epoch_timestamp) from _final_validator_dashboard_data_epoch) from _final_validator_dashboard_data_weekly PREWHERE validator_index = 0 limit 1 by t
UNION ALL
select 'botched_epoch_start', '_final_validator_dashboard_data_monthly', t, (select max(epoch_timestamp) from _final_validator_dashboard_data_epoch) from _final_validator_dashboard_data_monthly PREWHERE validator_index = 0 limit 1 by t;
-- +goose StatementEnd
-- latest entry in _exporter_aggregate_backfill_metadata gets marked as dirty from the get go - since we are doing the switch to the new revision it needs to be aggregated
-- +goose StatementBegin
INSERT INTO _exporter_aggregate_backfill_metadata (backfill_name, aggregation, t, max_epoch_timestamp, did_check, backfill_id)
SELECT 
    'botched_epoch_start',
    aggregation,
    t,
    max_epoch_timestamp,
    now(),
    generateUUIDv4()
from _exporter_aggregate_backfill_metadata
WHERE 
    backfill_name = 'botched_epoch_start'
ORDER BY t desc 
LIMIT 1 BY aggregation;
-- +goose StatementEnd
    

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS _exporter_aggregate_backfill_metadata;
-- +goose StatementEnd
