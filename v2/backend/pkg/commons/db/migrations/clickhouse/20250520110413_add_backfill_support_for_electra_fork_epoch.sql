-- +goose Up
-- +goose StatementBegin
 CREATE TABLE IF NOT EXISTS _insert_sink_backfill_electra_fork_epoch_events
(
    `epoch_timestamp` DateTime COMMENT 'timestamp of the epoch',
    `validator_index` UInt64 COMMENT 'validator index',
    `withdrawals_amount` Int64 COMMENT 'sum of withdrawals_amount',
    `withdrawals_count` Int64 COMMENT 'sum of withdrawals_count',
    -- we will need something to fix the roi as well
    -- roi_dividend is `balance_end - deposits_amount + withdrawals_amount - consolidations_incoming_amount + consolidations_outgoing_amount`
    -- so to fix it, subtract the withdrawals_amount from the roi_divisor
)
ENGINE = ReplacingMergeTree() -- we need it persistent because we will use it to update the epoch table through the dictionary
ORDER BY (validator_index, epoch_timestamp)
SETTINGS index_granularity=128, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048;
-- +goose StatementEnd
-- +goose StatementBegin
-- connect the sink to the _final_validator_dashboard_roi_hourly table
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_backfill_final_validator_dashboard_roi_hourly_electra_fork TO _final_validator_dashboard_roi_hourly
(
    `validator_index` UInt64,
    `t` DateTime,
    `roi_dividend` Int128,
    `roi_divisor` Int128
)
AS SELECT
    validator_index AS validator_index,
    toStartOfHour(foo.epoch_timestamp) AS t,
    SUM(CAST(foo.withdrawals_amount, 'Int128')) AS roi_dividend, -- i.e, we add the withdrawal amount to the roi_dividend
    0 AS roi_divisor -- i.e, we dont modify the divisor
FROM _insert_sink_backfill_electra_fork_epoch_events AS foo
GROUP BY
    t,
    validator_index
-- once it is in the hourly, it will be automatically propagated to the daily, weekly and monthly tables by the existing materialized views
-- +goose StatementEnd
-- +goose StatementBegin
-- connect the sink to the _final_validator_dashboard_data_hourly table
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_backfill_final_validator_dashboard_data_hourly_electra_fork TO _final_validator_dashboard_data_hourly
(
    `validator_index` UInt64,
    `t` DateTime,
    `withdrawals_amount` Int64,
    `withdrawals_count` Int64
) 
AS SELECT
    validator_index AS validator_index,
    toStartOfHour(foo.epoch_timestamp) AS t,
    SUM(foo.withdrawals_amount) AS withdrawals_amount, -- i.e we add the withdrawal amount to the withdrawal amount
    SUM(foo.withdrawals_count) AS withdrawals_count -- i.e we add the withdrawal count to the withdrawal count
FROM _insert_sink_backfill_electra_fork_epoch_events AS foo
GROUP BY
    t,
    validator_index
-- +goose StatementEnd
-- +goose StatementBegin
SELECT 'the migration will now attempt to create a dictionary. this will only work if the migration is run as the default (root) user.';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE DICTIONARY IF NOT EXISTS _dict_backfill_electra_fork_epoch_events (validator_index Int64, epoch_timestamp DateTime, withdrawals_amount Int64, withdrawals_count Int64)
PRIMARY KEY epoch_timestamp, validator_index SOURCE(CLICKHOUSE(TABLE _insert_sink_backfill_electra_fork_epoch_events DB currentDatabase()))
LAYOUT(complex_key_direct()); -- we dont use complex_key_cache here because that would require us to make sure that the cache is up to date when we trigger the mutation. it should only be at most 200k keys anyways
-- +goose StatementEnd
-- +goose StatementBegin
-- kickstart the backfill with the latest exported epoch
INSERT INTO _exporter_backfill_metadata 
    SELECT
        epoch,
        'electra_fork_epoch_events' as backfill_name,
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
SELECT 'successfully created the dictionary. now start the dashboard exporter again. once the backfill is done (i.e it shows up as 0 epochs left in the grafana dashboard), run the next migration version to update the epoch table.';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS _mv_backfill_final_validator_dashboard_roi_hourly_electra_fork;
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW IF EXISTS _mv_backfill_final_validator_dashboard_data_hourly_electra_fork;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _exporter_backfill_metadata 
    DELETE WHERE backfill_name = 'electra_fork_epoch_events';
-- +goose StatementEnd
-- +goose StatementBegin
DROP DICTIONARY IF EXISTS _dict_backfill_electra_fork_epoch_events;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _insert_sink_backfill_electra_fork_epoch_events;
-- +goose StatementEnd