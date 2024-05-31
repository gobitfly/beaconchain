-- +goose Up
-- +goose StatementBegin
CREATE TABLE validator_dashboard_data_daily
(
    `validator_index` UInt64,
    `day` Date,
    `epoch_start`Int64,
    `epoch_end`Int64,
    `attestations_source_reward` Nullable(Int64),
    `attestations_target_reward` Nullable(Int64),
    `attestations_head_reward` Nullable(Int64),
    `attestations_inactivity_reward` Nullable(Int64),
    `attestations_inclusion_reward` Nullable(Int64),
    `attestations_reward` Nullable(Int64),
    `attestations_ideal_source_reward` Nullable(Int64),
    `attestations_ideal_target_reward` Nullable(Int64),
    `attestations_ideal_head_reward` Nullable(Int64),
    `attestations_ideal_inactivity_reward` Nullable(Int64),
    `attestations_ideal_inclusion_reward` Nullable(Int64),
    `attestations_ideal_reward` Nullable(Int64),
    `blocks_scheduled` Nullable(Int64),
    `blocks_proposed` Nullable(Int64),
    `blocks_cl_reward` Nullable(Int64),
    `blocks_el_reward` Nullable(Int64),
    `sync_scheduled` Nullable(Int64),
    `sync_executed` Nullable(Int64),
    `sync_rewards` Nullable(Int64),
    `slashed` Nullable(Bool),
    `balance_start` Nullable(Int64),
    `balance_end` Nullable(Int64),
    `deposits_count` Nullable(Int64),
    `deposits_amount` Nullable(Int64),
    `withdrawals_count` Nullable(Int64),
    `withdrawals_amount` Nullable(Int64),
    `inclusion_delay_sum` Nullable(Int64),
    `blocks_expected` Nullable(Float64),
    `sync_committees_expected` Nullable(Float64),
    `attestations_scheduled` Nullable(Int64),
    `attestations_executed` Nullable(Int64),
    `attestation_head_executed` Nullable(Int64),
    `attestation_source_executed` Nullable(Int64),
    `attestation_target_executed` Nullable(Int64),
    `optimal_inclusion_delay_sum` Nullable(Int64),
    `slashed_by` Nullable(Int64), --- should only ever be set once so any should be fine
    `slashed_violation` Nullable(Int64), --- should only ever be set once so any should be fine
    `slasher_reward` Nullable(Int64),
    `last_executed_duty_epoch` Nullable(Int64),
    `blocks_cl_attestations_reward` Nullable(Int64),
    `blocks_cl_sync_aggregate_reward` Nullable(Int64),
    -- add projection to optimize validator_index queries   
    Projection validator_index_day_projection
    (
        select * order by validator_index, day
    )
)
ENGINE = MergeTree()
PRIMARY KEY (day, validator_index)
ORDER BY (day, validator_index)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE validator_dashboard_data_weekly
(
    `validator_index` UInt64,
    `week` Date,
    `epoch_start` SimpleAggregateFunction(min, Int64),
    `epoch_end` SimpleAggregateFunction(max, Int64),
    `attestations_source_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_target_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_head_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_inactivity_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_inclusion_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_source_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_target_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_head_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_inactivity_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_inclusion_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_scheduled` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_proposed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_cl_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_el_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `sync_scheduled` SimpleAggregateFunction(sum, Nullable(Int64)),
    `sync_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `sync_rewards` SimpleAggregateFunction(sum, Nullable(Int64)),
    `slashed` SimpleAggregateFunction(max, Nullable(Bool)),
    `balance_start` AggregateFunction(argMin, Nullable(Int64), Int64),
    `balance_end` AggregateFunction(argMax, Nullable(Int64), Int64),
    `deposits_count` SimpleAggregateFunction(sum, Nullable(Int64)),
    `deposits_amount` SimpleAggregateFunction(sum, Nullable(Int64)),
    `withdrawals_count` SimpleAggregateFunction(sum, Nullable(Int64)),
    `withdrawals_amount` SimpleAggregateFunction(sum, Nullable(Int64)),
    `inclusion_delay_sum` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_expected` SimpleAggregateFunction(sum, Nullable(Float64)),
    `sync_committees_expected` SimpleAggregateFunction(sum, Nullable(Float64)),
    `attestations_scheduled` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestation_head_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestation_source_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestation_target_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `optimal_inclusion_delay_sum` SimpleAggregateFunction(sum, Nullable(Int64)),
    `slasher_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `slashed_by` SimpleAggregateFunction(any, Nullable(Int64)), --- should only ever be set once so any should be fine
    `slashed_violation` SimpleAggregateFunction(any, Nullable(Int64)), --- should only ever be set once so any should be fine
    `last_executed_duty_epoch` SimpleAggregateFunction(max, Nullable(Int64)),
    `blocks_cl_attestations_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_cl_sync_aggregate_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    -- add projection to optimize validator_index queries   
    Projection validator_index_week_projection
    (
        select * order by validator_index, week
    )
)
ENGINE = AggregatingMergeTree()
PRIMARY KEY (week, validator_index)
ORDER BY (week, validator_index)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE validator_dashboard_data_monthly
(
    `validator_index` UInt64,
    `month` Date,
    `epoch_start` SimpleAggregateFunction(min, Int64),
    `epoch_end` SimpleAggregateFunction(max, Int64),
    `attestations_source_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_target_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_head_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_inactivity_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_inclusion_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_source_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_target_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_head_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_inactivity_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_inclusion_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_scheduled` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_proposed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_cl_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_el_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `sync_scheduled` SimpleAggregateFunction(sum, Nullable(Int64)),
    `sync_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `sync_rewards` SimpleAggregateFunction(sum, Nullable(Int64)),
    `slashed` SimpleAggregateFunction(max, Nullable(Bool)),
    `balance_start` AggregateFunction(argMin, Nullable(Int64), Int64),
    `balance_end` AggregateFunction(argMax, Nullable(Int64), Int64),
    `deposits_count` SimpleAggregateFunction(sum, Nullable(Int64)),
    `deposits_amount` SimpleAggregateFunction(sum, Nullable(Int64)),
    `withdrawals_count` SimpleAggregateFunction(sum, Nullable(Int64)),
    `withdrawals_amount` SimpleAggregateFunction(sum, Nullable(Int64)),
    `inclusion_delay_sum` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_expected` SimpleAggregateFunction(sum, Nullable(Float64)),
    `sync_committees_expected` SimpleAggregateFunction(sum, Nullable(Float64)),
    `attestations_scheduled` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestation_head_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestation_source_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestation_target_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `optimal_inclusion_delay_sum` SimpleAggregateFunction(sum, Nullable(Int64)),
    `slasher_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `slashed_by` SimpleAggregateFunction(any, Nullable(Int64)), --- should only ever be set once so any should be fine
    `slashed_violation` SimpleAggregateFunction(any, Nullable(Int64)), --- should only ever be set once so any should be fine
    `last_executed_duty_epoch` SimpleAggregateFunction(max, Nullable(Int64)),
    `blocks_cl_attestations_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_cl_sync_aggregate_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    -- add projection to optimize validator_index queries   
    Projection validator_index_month_projection
    (
        select * order by validator_index, month
    )
)
ENGINE = AggregatingMergeTree()
PRIMARY KEY (month, validator_index)
ORDER BY (month, validator_index)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE validator_dashboard_data_quarterly
(
    `validator_index` UInt64,
    `quarter` Date,
    `epoch_start` SimpleAggregateFunction(min, Int64),
    `epoch_end` SimpleAggregateFunction(max, Int64),
    `attestations_source_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_target_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_head_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_inactivity_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_inclusion_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_source_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_target_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_head_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_inactivity_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_inclusion_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_ideal_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_scheduled` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_proposed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_cl_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_el_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `sync_scheduled` SimpleAggregateFunction(sum, Nullable(Int64)),
    `sync_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `sync_rewards` SimpleAggregateFunction(sum, Nullable(Int64)),
    `slashed` SimpleAggregateFunction(max, Nullable(Bool)),
    `balance_start` AggregateFunction(argMin, Nullable(Int64), Int64),
    `balance_end` AggregateFunction(argMax, Nullable(Int64), Int64),
    `deposits_count` SimpleAggregateFunction(sum, Nullable(Int64)),
    `deposits_amount` SimpleAggregateFunction(sum, Nullable(Int64)),
    `withdrawals_count` SimpleAggregateFunction(sum, Nullable(Int64)),
    `withdrawals_amount` SimpleAggregateFunction(sum, Nullable(Int64)),
    `inclusion_delay_sum` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_expected` SimpleAggregateFunction(sum, Nullable(Float64)),
    `sync_committees_expected` SimpleAggregateFunction(sum, Nullable(Float64)),
    `attestations_scheduled` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestations_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestation_head_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestation_source_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `attestation_target_executed` SimpleAggregateFunction(sum, Nullable(Int64)),
    `optimal_inclusion_delay_sum` SimpleAggregateFunction(sum, Nullable(Int64)),
    `slasher_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `slashed_by` SimpleAggregateFunction(any, Nullable(Int64)), --- should only ever be set once so any should be fine
    `slashed_violation` SimpleAggregateFunction(any, Nullable(Int64)), --- should only ever be set once so any should be fine
    `last_executed_duty_epoch` SimpleAggregateFunction(max, Nullable(Int64)),
    `blocks_cl_attestations_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    `blocks_cl_sync_aggregate_reward` SimpleAggregateFunction(sum, Nullable(Int64)),
    -- add projection to optimize validator_index queries   
    Projection validator_index_quarter_projection
    (
        select * order by validator_index, quarter
    )
)
ENGINE = AggregatingMergeTree()
PRIMARY KEY (quarter, validator_index)
ORDER BY (quarter, validator_index)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd
-- +goose StatementBegin

CREATE MATERIALIZED VIEW validator_dashboard_data_weekly_mv TO validator_dashboard_data_weekly
AS SELECT
    validator_index AS validator_index,
    toMonday(day) AS week,
    minSimpleState(epoch_start) as epoch_start,
    maxSimpleState(epoch_end) as epoch_end,
    sumSimpleState(attestations_source_reward) as attestations_source_reward,
    sumSimpleState(attestations_target_reward) as attestations_target_reward,
    sumSimpleState(attestations_head_reward) as attestations_head_reward,
    sumSimpleState(attestations_inactivity_reward) as attestations_inactivity_reward,
    sumSimpleState(attestations_inclusion_reward) as attestations_inclusion_reward,
    sumSimpleState(attestations_reward) as attestations_reward,
    sumSimpleState(attestations_ideal_source_reward) as attestations_ideal_source_reward,
    sumSimpleState(attestations_ideal_target_reward) as attestations_ideal_target_reward,
    sumSimpleState(attestations_ideal_head_reward) as attestations_ideal_head_reward,
    sumSimpleState(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
    sumSimpleState(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
    sumSimpleState(attestations_ideal_reward) as attestations_ideal_reward,
    sumSimpleState(blocks_scheduled) as blocks_scheduled,
    sumSimpleState(blocks_proposed) as blocks_proposed,
    sumSimpleState(blocks_cl_reward) as blocks_cl_reward,
    sumSimpleState(blocks_el_reward) as blocks_el_reward,
    sumSimpleState(sync_scheduled) as sync_scheduled,
    sumSimpleState(sync_executed) as sync_executed,
    sumSimpleState(sync_rewards) as sync_rewards,
    maxSimpleState(slashed) as slashed,
    argMinState(foo.balance_start, foo.epoch_start) as balance_start,
    argMaxState(foo.balance_end, foo.epoch_end) as balance_end,
    sumSimpleState(deposits_count) as deposits_count,
    sumSimpleState(deposits_amount) as deposits_amount,
    sumSimpleState(withdrawals_count) as withdrawals_count,
    sumSimpleState(withdrawals_amount) as withdrawals_amount,
    sumSimpleState(inclusion_delay_sum) as inclusion_delay_sum,
    sumSimpleState(blocks_expected) as blocks_expected,
    sumSimpleState(sync_committees_expected) as sync_committees_expected,
    sumSimpleState(attestations_scheduled) as attestations_scheduled,
    sumSimpleState(attestations_executed) as attestations_executed,
    sumSimpleState(attestation_head_executed) as attestation_head_executed,
    sumSimpleState(attestation_source_executed) as attestation_source_executed,
    sumSimpleState(attestation_target_executed) as attestation_target_executed,
    sumSimpleState(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum,
    sumSimpleState(slasher_reward) as slasher_reward,
    anySimpleState(slashed_by) as slashed_by,
    anySimpleState(slashed_violation) as slashed_violation,
    maxSimpleState(last_executed_duty_epoch) as last_executed_duty_epoch,
    sumSimpleState(blocks_cl_attestations_reward) as blocks_cl_attestations_reward,
    sumSimpleState(blocks_cl_sync_aggregate_reward) as blocks_cl_sync_aggregate_reward
FROM validator_dashboard_data_daily foo
GROUP BY
    week,
    validator_index
-- +goose StatementEnd
-- +goose StatementBegin

CREATE MATERIALIZED VIEW validator_dashboard_data_monthly_mv TO validator_dashboard_data_monthly
AS SELECT
    validator_index AS validator_index,
    toStartOfMonth(week) AS month,
    minSimpleState(epoch_start) as epoch_start,
    maxSimpleState(epoch_end) as epoch_end,
    sumSimpleState(attestations_source_reward) as attestations_source_reward,
    sumSimpleState(attestations_target_reward) as attestations_target_reward,
    sumSimpleState(attestations_head_reward) as attestations_head_reward,
    sumSimpleState(attestations_inactivity_reward) as attestations_inactivity_reward,
    sumSimpleState(attestations_inclusion_reward) as attestations_inclusion_reward,
    sumSimpleState(attestations_reward) as attestations_reward,
    sumSimpleState(attestations_ideal_source_reward) as attestations_ideal_source_reward,
    sumSimpleState(attestations_ideal_target_reward) as attestations_ideal_target_reward,
    sumSimpleState(attestations_ideal_head_reward) as attestations_ideal_head_reward,
    sumSimpleState(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
    sumSimpleState(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
    sumSimpleState(attestations_ideal_reward) as attestations_ideal_reward,
    sumSimpleState(blocks_scheduled) as blocks_scheduled,
    sumSimpleState(blocks_proposed) as blocks_proposed,
    sumSimpleState(blocks_cl_reward) as blocks_cl_reward,
    sumSimpleState(blocks_el_reward) as blocks_el_reward,
    sumSimpleState(sync_scheduled) as sync_scheduled,
    sumSimpleState(sync_executed) as sync_executed,
    sumSimpleState(sync_rewards) as sync_rewards,
    maxSimpleState(slashed) as slashed,
    argMinMergeState(balance_start) as balance_start,
    argMaxMergeState(balance_end) as balance_end,
    sumSimpleState(deposits_count) as deposits_count,
    sumSimpleState(deposits_amount) as deposits_amount,
    sumSimpleState(withdrawals_count) as withdrawals_count,
    sumSimpleState(withdrawals_amount) as withdrawals_amount,
    sumSimpleState(inclusion_delay_sum) as inclusion_delay_sum,
    sumSimpleState(blocks_expected) as blocks_expected,
    sumSimpleState(sync_committees_expected) as sync_committees_expected,
    sumSimpleState(attestations_scheduled) as attestations_scheduled,
    sumSimpleState(attestations_executed) as attestations_executed,
    sumSimpleState(attestation_head_executed) as attestation_head_executed,
    sumSimpleState(attestation_source_executed) as attestation_source_executed,
    sumSimpleState(attestation_target_executed) as attestation_target_executed,
    sumSimpleState(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum,
    sumSimpleState(slasher_reward) as slasher_reward,
    anySimpleState(slashed_by) as slashed_by,
    anySimpleState(slashed_violation) as slashed_violation,
    maxSimpleState(last_executed_duty_epoch) as last_executed_duty_epoch,
    sumSimpleState(blocks_cl_attestations_reward) as blocks_cl_attestations_reward,
    sumSimpleState(blocks_cl_sync_aggregate_reward) as blocks_cl_sync_aggregate_reward
FROM validator_dashboard_data_weekly
GROUP BY
    month,
    validator_index
-- +goose StatementEnd
-- +goose StatementBegin

CREATE MATERIALIZED VIEW validator_dashboard_data_quarterly_mv TO validator_dashboard_data_quarterly
AS SELECT
    validator_index AS validator_index,
    toStartOfQuarter(month) AS quarter,
    minSimpleState(epoch_start) as epoch_start,
    maxSimpleState(epoch_end) as epoch_end,
    sumSimpleState(attestations_source_reward) as attestations_source_reward,
    sumSimpleState(attestations_target_reward) as attestations_target_reward,
    sumSimpleState(attestations_head_reward) as attestations_head_reward,
    sumSimpleState(attestations_inactivity_reward) as attestations_inactivity_reward,
    sumSimpleState(attestations_inclusion_reward) as attestations_inclusion_reward,
    sumSimpleState(attestations_reward) as attestations_reward,
    sumSimpleState(attestations_ideal_source_reward) as attestations_ideal_source_reward,
    sumSimpleState(attestations_ideal_target_reward) as attestations_ideal_target_reward,
    sumSimpleState(attestations_ideal_head_reward) as attestations_ideal_head_reward,
    sumSimpleState(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
    sumSimpleState(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
    sumSimpleState(attestations_ideal_reward) as attestations_ideal_reward,
    sumSimpleState(blocks_scheduled) as blocks_scheduled,
    sumSimpleState(blocks_proposed) as blocks_proposed,
    sumSimpleState(blocks_cl_reward) as blocks_cl_reward,
    sumSimpleState(blocks_el_reward) as blocks_el_reward,
    sumSimpleState(sync_scheduled) as sync_scheduled,
    sumSimpleState(sync_executed) as sync_executed,
    sumSimpleState(sync_rewards) as sync_rewards,
    maxSimpleState(slashed) as slashed,
    argMinMergeState(balance_start) as balance_start,
    argMaxMergeState(balance_end) as balance_end,
    sumSimpleState(deposits_count) as deposits_count,
    sumSimpleState(deposits_amount) as deposits_amount,
    sumSimpleState(withdrawals_count) as withdrawals_count,
    sumSimpleState(withdrawals_amount) as withdrawals_amount,
    sumSimpleState(inclusion_delay_sum) as inclusion_delay_sum,
    sumSimpleState(blocks_expected) as blocks_expected,
    sumSimpleState(sync_committees_expected) as sync_committees_expected,
    sumSimpleState(attestations_scheduled) as attestations_scheduled,
    sumSimpleState(attestations_executed) as attestations_executed,
    sumSimpleState(attestation_head_executed) as attestation_head_executed,
    sumSimpleState(attestation_source_executed) as attestation_source_executed,
    sumSimpleState(attestation_target_executed) as attestation_target_executed,
    sumSimpleState(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum,
    sumSimpleState(slasher_reward) as slasher_reward,
    anySimpleState(slashed_by) as slashed_by,
    anySimpleState(slashed_violation) as slashed_violation,
    maxSimpleState(last_executed_duty_epoch) as last_executed_duty_epoch,
    sumSimpleState(blocks_cl_attestations_reward) as blocks_cl_attestations_reward,
    sumSimpleState(blocks_cl_sync_aggregate_reward) as blocks_cl_sync_aggregate_reward
FROM validator_dashboard_data_monthly
GROUP BY
    month,
    validator_index
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS validator_dashboard_data_daily;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS validator_dashboard_data_weekly;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS validator_dashboard_data_monthly;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS validator_dashboard_data_quarterly;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS validator_dashboard_data_weekly_mv;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS validator_dashboard_data_monthly_mv;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS validator_dashboard_data_quarterly_mv;
-- +goose StatementEnd