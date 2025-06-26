-- +goose Up
-- +goose StatementBegin 
INSERT INTO _exporter_backfill_metadata 
    SELECT
        epoch,
        'missing_missed_rewards' as backfill_name,
        transfer_batch_id as backfill_batch_id,
        NULL AS successful_backfill
    FROM
    (
        SELECT *
        FROM _exporter_metadata
        WHERE successful_transfer IS NOT NULL
        ORDER BY epoch DESC
        LIMIT 1
    )
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    ADD COLUMN IF NOT EXISTS _fixed_blocks_cl_missed_median_reward SimpleAggregateFunction(sum, Int64) settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    ADD COLUMN IF NOT EXISTS _fixed_blocks_cl_missed_median_reward SimpleAggregateFunction(sum, Int64) settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    ADD COLUMN IF NOT EXISTS _fixed_blocks_cl_missed_median_reward SimpleAggregateFunction(sum, Int64) settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    ADD COLUMN IF NOT EXISTS _fixed_blocks_cl_missed_median_reward SimpleAggregateFunction(sum, Int64) settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- 4) add backfill insert sink
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _insert_sink_backfill_missing_missed_rewards
(
    `validator_index` UInt64 COMMENT 'validator index',
    `t` DateTime COMMENT 'timestamp',
    `epoch` Int64 COMMENT 'epoch start', -- leaving it at the default of 0 would break epoch_start
    `balance_start` Int64 COMMENT 'balance at the start of the epoch',  -- leaving it at the default of 0 would break balance_start and balance_min
    `_fixed_blocks_cl_missed_median_reward` Int64 COMMENT 'fixed blocks cl missed median reward'
)
ENGINE = Null;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_backfill_final_validator_dashboard_data_hourly_missing_missed_rewards
TO _final_validator_dashboard_data_hourly
AS SELECT
    validator_index as validator_index,
    toStartOfHour(t) as t,
    epoch as epoch_start,
    foo.balance_start as balance_min,
    initializeAggregation('argMinState', balance_start, epoch) as balance_start,
    _fixed_blocks_cl_missed_median_reward as _fixed_blocks_cl_missed_median_reward
from _insert_sink_backfill_missing_missed_rewards as foo
WHERE _fixed_blocks_cl_missed_median_reward != 0;
-- +goose StatementEnd
-- +goose StatementBegin
-- following exists to ensure that during normal importing of epochs we also update
-- the column that will be used during backfilling.
-- at the end of the backfill we should be able to cleanly swap the blocks_cl_missed_median_reward
-- column with the _fixed_blocks_cl_missed_median_reward column with the _fixed_blocks_cl_missed_median_reward column
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_backfill_forwards_sync_final_validator_dashboard_data_hourly_missing_missed_rewards
TO _final_validator_dashboard_data_hourly
AS SELECT
    validator_index as validator_index,
    toStartOfHour(epoch_timestamp) as t,
    epoch as epoch_start,
    foo.balance_start as balance_min,
    initializeAggregation('argMinState', balance_start, epoch) as balance_start,
    blocks_cl_missed_median_reward as _fixed_blocks_cl_missed_median_reward
from _final_validator_dashboard_data_epoch foo
WHERE blocks_cl_missed_median_reward != 0
-- +goose StatementEnd
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_backfill_final_validator_dashboard_data_daily_missing_missed_rewards
TO _final_validator_dashboard_data_daily
AS SELECT
    validator_index as validator_index,
    toStartOfDay(t) as t,
    epoch_start as epoch_start,
    balance_min as balance_min,
    balance_start as balance_start,
    _fixed_blocks_cl_missed_median_reward as _fixed_blocks_cl_missed_median_reward
FROM _final_validator_dashboard_data_hourly  -- this means it also gets triggered during the normal insert flow, which is correct and wanted
WHERE _fixed_blocks_cl_missed_median_reward != 0;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_backfill_final_validator_dashboard_data_weekly_missing_missed_rewards
TO _final_validator_dashboard_data_weekly
AS SELECT
    validator_index as validator_index,
    toMonday(t) as t,
    epoch_start as epoch_start,
    balance_min as balance_min,
    balance_start as balance_start,
    _fixed_blocks_cl_missed_median_reward as _fixed_blocks_cl_missed_median_reward
FROM _final_validator_dashboard_data_daily  -- this means it also gets triggered during the normal insert flow, which is correct and wanted
WHERE _fixed_blocks_cl_missed_median_reward != 0;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_backfill_final_validator_dashboard_data_monthly_missing_missed_rewards
TO _final_validator_dashboard_data_monthly
AS SELECT
    validator_index as validator_index,
    toStartOfMonth(t) as t,
    epoch_start as epoch_start,
    balance_min as balance_min,
    balance_start as balance_start,
    _fixed_blocks_cl_missed_median_reward as _fixed_blocks_cl_missed_median_reward
FROM _final_validator_dashboard_data_daily  -- this means it also gets triggered during the normal insert flow, which is correct and wanted
WHERE _fixed_blocks_cl_missed_median_reward != 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
