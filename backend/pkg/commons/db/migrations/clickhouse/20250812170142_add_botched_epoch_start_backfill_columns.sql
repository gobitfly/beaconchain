-- +goose Up

-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    COMMENT COLUMN epoch_start 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    COMMENT COLUMN balance_start 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    COMMENT COLUMN balance_min 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    ADD COLUMN IF NOT EXISTS _is_backfill SimpleAggregateFunction(max, bool) DEFAULT 0 COMMENT 'if the row is part of a backfill insert',
    ADD COLUMN IF NOT EXISTS _botched_epoch_start_backfill_revision SimpleAggregateFunction(max, Int64) DEFAULT 0 COMMENT 'revision of the "botched epoch start" backfill. incrementing this will cause _staging_{epoch_start|balance_start|balance_min} to be replaced with whatever the _legacy columns are set to in the same row.',
    ADD COLUMN IF NOT EXISTS _staging_epoch_start AggregateFunction(minSimpleStateArgMax, Int64, Int64) Materialized initializeAggregation('minSimpleStateArgMaxState', epoch_start::Int64, _botched_epoch_start_backfill_revision::Int64),
    ADD COLUMN IF NOT EXISTS _staging_balance_start AggregateFunction(argMinMergeStateArgMax, AggregateFunction(argMin, Int64, Int64), SimpleAggregateFunction(max, Int64)) Materialized initializeAggregation('argMinMergeStateArgMaxState', balance_start, _botched_epoch_start_backfill_revision::Int64),
    ADD COLUMN IF NOT EXISTS _staging_balance_min AggregateFunction(minSimpleStateArgMax, Int64, Int64) Materialized initializeAggregation('minSimpleStateArgMaxState', balance_min::Int64, _botched_epoch_start_backfill_revision::Int64),
    -- the existing columns will be swapped out with the following aliases once the backfill is complete
    ADD COLUMN IF NOT EXISTS _public_epoch_start SimpleAggregateFunction(min, Int64) ALIAS finalizeAggregation(_staging_epoch_start),
    ADD COLUMN IF NOT EXISTS _public_balance_start AggregateFunction(argMin, Int64, Int64) ALIAS finalizeAggregation(_staging_balance_start),
    ADD COLUMN IF NOT EXISTS _public_balance_min SimpleAggregateFunction(min, Int64) ALIAS finalizeAggregation(_staging_balance_min)
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    COMMENT COLUMN epoch_start 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    COMMENT COLUMN balance_start 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    COMMENT COLUMN balance_min 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    ADD COLUMN IF NOT EXISTS _is_backfill SimpleAggregateFunction(max, bool) DEFAULT 0 COMMENT 'if the row is part of a backfill insert',
    ADD COLUMN IF NOT EXISTS _botched_epoch_start_backfill_revision SimpleAggregateFunction(max, Int64) DEFAULT 0 COMMENT 'revision of the "botched epoch start" backfill. incrementing this will cause _staging_{epoch_start|balance_start|balance_min} to be replaced with whatever the _legacy columns are set to in the same row.',
    ADD COLUMN IF NOT EXISTS _staging_epoch_start AggregateFunction(minSimpleStateArgMax, Int64, Int64) Materialized initializeAggregation('minSimpleStateArgMaxState', epoch_start::Int64, _botched_epoch_start_backfill_revision::Int64),
    ADD COLUMN IF NOT EXISTS _staging_balance_start AggregateFunction(argMinMergeStateArgMax, AggregateFunction(argMin, Int64, Int64), SimpleAggregateFunction(max, Int64)) Materialized initializeAggregation('argMinMergeStateArgMaxState', balance_start, _botched_epoch_start_backfill_revision::Int64),
    ADD COLUMN IF NOT EXISTS _staging_balance_min AggregateFunction(minSimpleStateArgMax, Int64, Int64) Materialized initializeAggregation('minSimpleStateArgMaxState', balance_min::Int64, _botched_epoch_start_backfill_revision::Int64),
    -- the existing columns will be swapped out with the following aliases once the backfill is complete
    ADD COLUMN IF NOT EXISTS _public_epoch_start SimpleAggregateFunction(min, Int64) ALIAS finalizeAggregation(_staging_epoch_start),
    ADD COLUMN IF NOT EXISTS _public_balance_start AggregateFunction(argMin, Int64, Int64) ALIAS finalizeAggregation(_staging_balance_start),
    ADD COLUMN IF NOT EXISTS _public_balance_min SimpleAggregateFunction(min, Int64) ALIAS finalizeAggregation(_staging_balance_min)
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    COMMENT COLUMN epoch_start 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    COMMENT COLUMN balance_start 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    COMMENT COLUMN balance_min 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    ADD COLUMN IF NOT EXISTS _is_backfill SimpleAggregateFunction(max, bool) DEFAULT 0 COMMENT 'if the row is part of a backfill insert',
    ADD COLUMN IF NOT EXISTS _botched_epoch_start_backfill_revision SimpleAggregateFunction(max, Int64) DEFAULT 0 COMMENT 'revision of the "botched epoch start" backfill. incrementing this will cause _staging_{epoch_start|balance_start|balance_min} to be replaced with whatever the _legacy columns are set to in the same row.',
    ADD COLUMN IF NOT EXISTS _staging_epoch_start AggregateFunction(minSimpleStateArgMax, Int64, Int64) Materialized initializeAggregation('minSimpleStateArgMaxState', epoch_start::Int64, _botched_epoch_start_backfill_revision::Int64),
    ADD COLUMN IF NOT EXISTS _staging_balance_start AggregateFunction(argMinMergeStateArgMax, AggregateFunction(argMin, Int64, Int64), SimpleAggregateFunction(max, Int64)) Materialized initializeAggregation('argMinMergeStateArgMaxState', balance_start, _botched_epoch_start_backfill_revision::Int64),
    ADD COLUMN IF NOT EXISTS _staging_balance_min AggregateFunction(minSimpleStateArgMax, Int64, Int64) Materialized initializeAggregation('minSimpleStateArgMaxState', balance_min::Int64, _botched_epoch_start_backfill_revision::Int64),
    -- the existing columns will be swapped out with the following aliases once the backfill is complete
    ADD COLUMN IF NOT EXISTS _public_epoch_start SimpleAggregateFunction(min, Int64) ALIAS finalizeAggregation(_staging_epoch_start),
    ADD COLUMN IF NOT EXISTS _public_balance_start AggregateFunction(argMin, Int64, Int64) ALIAS finalizeAggregation(_staging_balance_start),
    ADD COLUMN IF NOT EXISTS _public_balance_min SimpleAggregateFunction(min, Int64) ALIAS finalizeAggregation(_staging_balance_min)
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    COMMENT COLUMN epoch_start 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    COMMENT COLUMN balance_start 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    COMMENT COLUMN balance_min 'legacy column, used by ingestion. consumers of data should use the aliased variant instead',
    ADD COLUMN IF NOT EXISTS _is_backfill SimpleAggregateFunction(max, bool) DEFAULT 0 COMMENT 'if the row is part of a backfill insert',
    ADD COLUMN IF NOT EXISTS _botched_epoch_start_backfill_revision SimpleAggregateFunction(max, Int64) DEFAULT 0 COMMENT 'revision of the "botched epoch start" backfill. incrementing this will cause _staging_{epoch_start|balance_start|balance_min} to be replaced with whatever the _legacy columns are set to in the same row.',
    ADD COLUMN IF NOT EXISTS _staging_epoch_start AggregateFunction(minSimpleStateArgMax, Int64, Int64) Materialized initializeAggregation('minSimpleStateArgMaxState', epoch_start::Int64, _botched_epoch_start_backfill_revision::Int64),
    ADD COLUMN IF NOT EXISTS _staging_balance_start AggregateFunction(argMinMergeStateArgMax, AggregateFunction(argMin, Int64, Int64), SimpleAggregateFunction(max, Int64)) Materialized initializeAggregation('argMinMergeStateArgMaxState', balance_start, _botched_epoch_start_backfill_revision::Int64),
    ADD COLUMN IF NOT EXISTS _staging_balance_min AggregateFunction(minSimpleStateArgMax, Int64, Int64) Materialized initializeAggregation('minSimpleStateArgMaxState', balance_min::Int64, _botched_epoch_start_backfill_revision::Int64),
    -- the existing columns will be swapped out with the following aliases once the backfill is complete
    ADD COLUMN IF NOT EXISTS _public_epoch_start SimpleAggregateFunction(min, Int64) ALIAS finalizeAggregation(_staging_epoch_start),
    ADD COLUMN IF NOT EXISTS _public_balance_start AggregateFunction(argMin, Int64, Int64) ALIAS finalizeAggregation(_staging_balance_start),
    ADD COLUMN IF NOT EXISTS _public_balance_min SimpleAggregateFunction(min, Int64) ALIAS finalizeAggregation(_staging_balance_min)
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    DROP COLUMN IF EXISTS _is_backfill,
    DROP COLUMN IF EXISTS _botched_epoch_start_backfill_revision,
    DROP COLUMN IF EXISTS _staging_epoch_start,
    DROP COLUMN IF EXISTS _staging_balance_start,
    DROP COLUMN IF EXISTS _staging_balance_min,
    DROP COLUMN IF EXISTS _public_epoch_start,
    DROP COLUMN IF EXISTS _public_balance_start,
    DROP COLUMN IF EXISTS _public_balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    DROP COLUMN IF EXISTS _is_backfill,
    DROP COLUMN IF EXISTS _botched_epoch_start_backfill_revision,
    DROP COLUMN IF EXISTS _staging_epoch_start,
    DROP COLUMN IF EXISTS _staging_balance_start,
    DROP COLUMN IF EXISTS _staging_balance_min,
    DROP COLUMN IF EXISTS _public_epoch_start,
    DROP COLUMN IF EXISTS _public_balance_start,
    DROP COLUMN IF EXISTS _public_balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    DROP COLUMN IF EXISTS _is_backfill,
    DROP COLUMN IF EXISTS _botched_epoch_start_backfill_revision,
    DROP COLUMN IF EXISTS _staging_epoch_start,
    DROP COLUMN IF EXISTS _staging_balance_start,
    DROP COLUMN IF EXISTS _staging_balance_min,
    DROP COLUMN IF EXISTS _public_epoch_start,
    DROP COLUMN IF EXISTS _public_balance_start,
    DROP COLUMN IF EXISTS _public_balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    DROP COLUMN IF EXISTS _is_backfill,
    DROP COLUMN IF EXISTS _botched_epoch_start_backfill_revision,
    DROP COLUMN IF EXISTS _staging_epoch_start,
    DROP COLUMN IF EXISTS _staging_balance_start,
    DROP COLUMN IF EXISTS _staging_balance_min,
    DROP COLUMN IF EXISTS _public_epoch_start,
    DROP COLUMN IF EXISTS _public_balance_start,
    DROP COLUMN IF EXISTS _public_balance_min
settings mutations_sync=2, alter_sync=1;
-- +goose StatementEnd
