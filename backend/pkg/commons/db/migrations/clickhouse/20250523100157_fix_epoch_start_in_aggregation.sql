-- +goose Up
-- +goose ENVSUB ON
-- +goose StatementBegin
set
    param_fork_epoch_ts = ${ELECTRA_FORK_EPOCH_TIMESTAMP?missing},
    param_genesis_epoch_ts = ${GENESIS_EPOCH_TIMESTAMP?missing},
    param_epoch_duration = ${EPOCH_DURATION_IN_SECONDS?missing};
-- +goose StatementEnd
-- +goose StatementBegin
-- generate table to be used to updated balance_start
CREATE TABLE IF NOT EXISTS _tmp_fix_epoch_balance_start
(
    epoch Int64,
    t DateTime,
    validator_index UInt64,
    balance_start Int64,
) ENGINE = ReplacingMergeTree()
ORDER BY (validator_index, t)
SETTINGS index_granularity=128, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE DICTIONARY IF NOT EXISTS _dict_fix_epoch_balance_start (validator_index Int64, t DateTime, epoch Int64, balance_start Int64)
PRIMARY KEY t, validator_index SOURCE(CLICKHOUSE(TABLE _tmp_fix_epoch_balance_start DB currentDatabase()))
LAYOUT(complex_key_direct());
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO _tmp_fix_epoch_balance_start
SELECT
    epoch,
    toStartOfHour({fork_epoch_ts:DateTime}) as t,
    validator_index,
    balance_start
FROM _final_validator_dashboard_data_epoch
WHERE epoch = GREATEST(
        0,
        CEIL(
            (
                toStartOfHour({fork_epoch_ts:DateTime})::DateTime - {genesis_epoch_ts:DateTime}
            ) / {epoch_duration:Int64}
        )
    )
UNION ALL 
SELECT 
    epoch,
    toStartOfDay({fork_epoch_ts:DateTime}) as t,
    validator_index,
    balance_start
FROM _final_validator_dashboard_data_epoch
WHERE epoch = GREATEST(
        0,
        CEIL(
            (
                toStartOfDay({fork_epoch_ts:DateTime})::DateTime - {genesis_epoch_ts:DateTime}
            ) / {epoch_duration:Int64}
        )
    )
UNION ALL 
SELECT 
    epoch,
    toMonday({fork_epoch_ts:DateTime}) as t,
    validator_index,
    balance_start
FROM _final_validator_dashboard_data_epoch
WHERE epoch = GREATEST(
        0,
        CEIL(
            (
                toMonday({fork_epoch_ts:DateTime})::DateTime - {genesis_epoch_ts:DateTime}
            ) / {epoch_duration:Int64}
        )
    )
UNION ALL 
SELECT 
    epoch,
    toStartOfMonth({fork_epoch_ts:DateTime}) as t,
    validator_index,
    balance_start
FROM _final_validator_dashboard_data_epoch
WHERE epoch = GREATEST(
        0,
        CEIL(
            (
                toStartOfMonth({fork_epoch_ts:DateTime})::DateTime - {genesis_epoch_ts:DateTime}
            ) / {epoch_duration:Int64}
        )
    )
-- +goose StatementEnd
-- +goose StatementBegin
SYSTEM SYNC REPLICA _tmp_fix_epoch_balance_start LIGHTWEIGHT;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    UPDATE
        epoch_start = dictGet(_dict_fix_epoch_balance_start, 'epoch', (t, validator_index)),
        balance_start = initializeAggregation('argMinState', dictGet(_dict_fix_epoch_balance_start, 'balance_start', (t, validator_index)), dictGet(_dict_fix_epoch_balance_start, 'epoch', (t, validator_index)))
    WHERE t = toStartOfHour({fork_epoch_ts:DateTime})
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    UPDATE
        epoch_start = dictGet(_dict_fix_epoch_balance_start, 'epoch', (t, validator_index)),
        balance_start = initializeAggregation('argMinState', dictGet(_dict_fix_epoch_balance_start, 'balance_start', (t, validator_index)), dictGet(_dict_fix_epoch_balance_start, 'epoch', (t, validator_index)))
    WHERE t = toStartOfDay({fork_epoch_ts:DateTime})
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    UPDATE
        epoch_start = dictGet(_dict_fix_epoch_balance_start, 'epoch', (t, validator_index)),
        balance_start = initializeAggregation('argMinState', dictGet(_dict_fix_epoch_balance_start, 'balance_start', (t, validator_index)), dictGet(_dict_fix_epoch_balance_start, 'epoch', (t, validator_index)))
    WHERE t = toMonday({fork_epoch_ts:DateTime})
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    UPDATE
        epoch_start = dictGet(_dict_fix_epoch_balance_start, 'epoch', (t, validator_index)),
        balance_start = initializeAggregation('argMinState', dictGet(_dict_fix_epoch_balance_start, 'balance_start', (t, validator_index)), dictGet(_dict_fix_epoch_balance_start, 'epoch', (t, validator_index)))
    WHERE t = toStartOfMonth({fork_epoch_ts:DateTime})
-- +goose StatementEnd
-- +goose ENVSUB OFF

-- +goose Down
-- +goose StatementBegin
select * from "downgrade not supported"."manual re-export recommended"
-- +goose StatementEnd
