-- +goose Up
-- materialized views
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_final_validator_dashboard_data_hourly TO _final_validator_dashboard_data_hourly
AS SELECT
    validator_index AS validator_index,
    toStartOfHour(epoch_timestamp) AS t,
    groupArraySortedIfState(2048)(-foo.epoch, validator_index = 0) AS epoch_map,
    min(foo.epoch) AS epoch_start,
    max(foo.epoch) AS epoch_end,
    argMinState(foo.balance_start, foo.epoch) AS balance_start,
    argMaxState(foo.balance_end, foo.epoch) AS balance_end,
    least(min(foo.balance_start), min(foo.balance_end)) AS balance_min,
    greatest(max(foo.balance_start), max(foo.balance_end)) AS balance_max,
    sum(deposits_count) AS deposits_count,
    sum(deposits_amount) AS deposits_amount,
    sum(withdrawals_count) AS withdrawals_count,
    sum(withdrawals_amount) AS withdrawals_amount,
    sum(attestations_scheduled) AS attestations_scheduled,
    sum(attestations_observed) AS attestations_observed,
    sum(attestations_head_matched) AS attestations_head_matched,
    sum(attestations_source_matched) AS attestations_source_matched,
    sum(attestations_target_matched) AS attestations_target_matched,
    sum(attestations_head_executed) AS attestations_head_executed,
    sum(attestations_source_executed) AS attestations_source_executed,
    sum(attestations_target_executed) AS attestations_target_executed,
    sum(attestations_reward) AS attestations_reward,
    sum(attestations_reward_rewards_only) AS attestations_reward_rewards_only,
    sum(attestations_reward_penalties_only) AS attestations_reward_penalties_only,
    sum(attestations_head_reward) AS attestations_head_reward,
    sum(attestations_source_reward) AS attestations_source_reward,
    sum(attestations_target_reward) AS attestations_target_reward,
    sum(attestations_localized_max_reward) AS attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) AS attestations_hyperlocalized_max_reward,
    sum(attestations_inclusion_reward) AS attestations_inclusion_reward,
    sum(attestations_inactivity_reward) AS attestations_inactivity_reward,
    sum(attestations_ideal_reward) AS attestations_ideal_reward,
    sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
    sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
    sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
    sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
    sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
    sum(inclusion_delay_sum) AS inclusion_delay_sum,
    sum(optimal_inclusion_delay_sum) AS optimal_inclusion_delay_sum,
    sum(blocks_scheduled) AS blocks_scheduled,
    sum(blocks_proposed) AS blocks_proposed,
    sum(blocks_cl_reward) AS blocks_cl_reward,
    sum(blocks_cl_attestations_reward) AS blocks_cl_attestations_reward,
    sum(blocks_cl_sync_aggregate_reward) AS blocks_cl_sync_aggregate_reward,
    sum(blocks_cl_slasher_reward) AS blocks_cl_slasher_reward,
    sum(blocks_cl_missed_median_reward) AS blocks_cl_missed_median_reward,
    sum(blocks_slashing_count) AS blocks_slashing_count,
    sum(blocks_expected) AS blocks_expected,
    sum(sync_scheduled) AS sync_scheduled,
    sum(sync_executed) AS sync_executed,
    sum(sync_reward) AS sync_reward,
    sum(sync_reward_rewards_only) AS sync_reward_rewards_only,
    sum(sync_reward_penalties_only) AS sync_reward_penalties_only,
    sum(sync_localized_max_reward) AS sync_localized_max_reward,
    sum(sync_committees_expected) AS sync_committees_expected,
    max(slashed) AS slashed,
    maxIfOrNull(foo.epoch, (foo.blocks_proposed != 0) OR (foo.sync_executed != 0) OR (foo.attestations_observed != 0)) AS last_executed_duty_epoch,
    maxIfOrNull(foo.epoch, foo.sync_scheduled != 0) AS last_scheduled_sync_epoch,
    maxIfOrNull(foo.epoch, foo.blocks_proposed != 0) AS last_scheduled_block_epoch
FROM _final_validator_dashboard_data_epoch AS foo
GROUP BY
    t,
    validator_index
-- +goose StatementEnd
-- +goose StatementBegin
-- hour => day. need to merge already merged states
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_final_validator_dashboard_data_daily TO _final_validator_dashboard_data_daily
AS SELECT
    validator_index AS validator_index,
    toStartOfDay(foo.t) AS t,
    groupArraySortedIfMergeState(2048)(epoch_map) AS epoch_map,
    min(epoch_start) AS epoch_start,
    max(epoch_end) AS epoch_end,
    argMinStateMerge(balance_start) AS balance_start,
    argMaxStateMerge(balance_end) AS balance_end,
    min(balance_min) AS balance_min,
    max(balance_max) AS balance_max,
    sum(deposits_count) AS deposits_count,
    sum(deposits_amount) AS deposits_amount,
    sum(withdrawals_count) AS withdrawals_count,
    sum(withdrawals_amount) AS withdrawals_amount,
    sum(attestations_scheduled) AS attestations_scheduled,
    sum(attestations_observed) AS attestations_observed,
    sum(attestations_head_matched) AS attestations_head_matched,
    sum(attestations_source_matched) AS attestations_source_matched,
    sum(attestations_target_matched) AS attestations_target_matched,
    sum(attestations_head_executed) AS attestations_head_executed,
    sum(attestations_source_executed) AS attestations_source_executed,
    sum(attestations_target_executed) AS attestations_target_executed,
    sum(attestations_reward) AS attestations_reward,
    sum(attestations_reward_rewards_only) AS attestations_reward_rewards_only,
    sum(attestations_reward_penalties_only) AS attestations_reward_penalties_only,
    sum(attestations_head_reward) AS attestations_head_reward,
    sum(attestations_source_reward) AS attestations_source_reward,
    sum(attestations_target_reward) AS attestations_target_reward,
    sum(attestations_localized_max_reward) AS attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) AS attestations_hyperlocalized_max_reward,
    sum(attestations_inclusion_reward) AS attestations_inclusion_reward,
    sum(attestations_inactivity_reward) AS attestations_inactivity_reward,
    sum(attestations_ideal_reward) AS attestations_ideal_reward,
    sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
    sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
    sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
    sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
    sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
    sum(inclusion_delay_sum) AS inclusion_delay_sum,
    sum(optimal_inclusion_delay_sum) AS optimal_inclusion_delay_sum,
    sum(blocks_scheduled) AS blocks_scheduled,
    sum(blocks_proposed) AS blocks_proposed,
    sum(blocks_cl_reward) AS blocks_cl_reward,
    sum(blocks_cl_attestations_reward) AS blocks_cl_attestations_reward,
    sum(blocks_cl_sync_aggregate_reward) AS blocks_cl_sync_aggregate_reward,
    sum(blocks_cl_slasher_reward) AS blocks_cl_slasher_reward,
    sum(blocks_cl_missed_median_reward) AS blocks_cl_missed_median_reward,
    sum(blocks_slashing_count) AS blocks_slashing_count,
    sum(blocks_expected) AS blocks_expected,
    sum(sync_scheduled) AS sync_scheduled,
    sum(sync_executed) AS sync_executed,
    sum(sync_reward) AS sync_reward,
    sum(sync_reward_rewards_only) AS sync_reward_rewards_only,
    sum(sync_reward_penalties_only) AS sync_reward_penalties_only,
    sum(sync_localized_max_reward) AS sync_localized_max_reward,
    sum(sync_committees_expected) AS sync_committees_expected,
    max(slashed) AS slashed,
    max(last_executed_duty_epoch) AS last_executed_duty_epoch,
    max(last_scheduled_sync_epoch) AS last_scheduled_sync_epoch,
    max(last_scheduled_block_epoch) AS last_scheduled_block_epoch
FROM _final_validator_dashboard_data_hourly AS foo
GROUP BY
    t,
    validator_index
-- +goose StatementEnd
-- +goose StatementBegin
-- day => week. same as the daily view, but with a different grouping
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_final_validator_dashboard_data_weekly TO _final_validator_dashboard_data_weekly
AS SELECT
    validator_index AS validator_index,
    toMonday(foo.t) AS t,
    groupArraySortedIfMergeState(2048)(epoch_map) AS epoch_map,
    min(epoch_start) AS epoch_start,
    max(epoch_end) AS epoch_end,
    argMinStateMerge(balance_start) AS balance_start,
    argMaxStateMerge(balance_end) AS balance_end,
    min(balance_min) AS balance_min,
    max(balance_max) AS balance_max,
    sum(deposits_count) AS deposits_count,
    sum(deposits_amount) AS deposits_amount,
    sum(withdrawals_count) AS withdrawals_count,
    sum(withdrawals_amount) AS withdrawals_amount,
    sum(attestations_scheduled) AS attestations_scheduled,
    sum(attestations_observed) AS attestations_observed,
    sum(attestations_head_matched) AS attestations_head_matched,
    sum(attestations_source_matched) AS attestations_source_matched,
    sum(attestations_target_matched) AS attestations_target_matched,
    sum(attestations_head_executed) AS attestations_head_executed,
    sum(attestations_source_executed) AS attestations_source_executed,
    sum(attestations_target_executed) AS attestations_target_executed,
    sum(attestations_reward) AS attestations_reward,
    sum(attestations_reward_rewards_only) AS attestations_reward_rewards_only,
    sum(attestations_reward_penalties_only) AS attestations_reward_penalties_only,
    sum(attestations_head_reward) AS attestations_head_reward,
    sum(attestations_source_reward) AS attestations_source_reward,
    sum(attestations_target_reward) AS attestations_target_reward,
    sum(attestations_localized_max_reward) AS attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) AS attestations_hyperlocalized_max_reward,
    sum(attestations_inclusion_reward) AS attestations_inclusion_reward,
    sum(attestations_inactivity_reward) AS attestations_inactivity_reward,
    sum(attestations_ideal_reward) AS attestations_ideal_reward,
    sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
    sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
    sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
    sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
    sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
    sum(inclusion_delay_sum) AS inclusion_delay_sum,
    sum(optimal_inclusion_delay_sum) AS optimal_inclusion_delay_sum,
    sum(blocks_scheduled) AS blocks_scheduled,
    sum(blocks_proposed) AS blocks_proposed,
    sum(blocks_cl_reward) AS blocks_cl_reward,
    sum(blocks_cl_attestations_reward) AS blocks_cl_attestations_reward,
    sum(blocks_cl_sync_aggregate_reward) AS blocks_cl_sync_aggregate_reward,
    sum(blocks_cl_slasher_reward) AS blocks_cl_slasher_reward,
    sum(blocks_cl_missed_median_reward) AS blocks_cl_missed_median_reward,
    sum(blocks_slashing_count) AS blocks_slashing_count,
    sum(blocks_expected) AS blocks_expected,
    sum(sync_scheduled) AS sync_scheduled,
    sum(sync_executed) AS sync_executed,
    sum(sync_reward) AS sync_reward,
    sum(sync_reward_rewards_only) AS sync_reward_rewards_only,
    sum(sync_reward_penalties_only) AS sync_reward_penalties_only,
    sum(sync_localized_max_reward) AS sync_localized_max_reward,
    sum(sync_committees_expected) AS sync_committees_expected,
    max(slashed) AS slashed,
    max(last_executed_duty_epoch) AS last_executed_duty_epoch,
    max(last_scheduled_sync_epoch) AS last_scheduled_sync_epoch,
    max(last_scheduled_block_epoch) AS last_scheduled_block_epoch
FROM _final_validator_dashboard_data_daily AS foo
GROUP BY
    t,
    validator_index
-- +goose StatementEnd
-- +goose StatementBegin
-- day => month. same as the daily view, but with a different grouping
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_final_validator_dashboard_data_monthly TO _final_validator_dashboard_data_monthly
AS SELECT
    validator_index AS validator_index,
    toStartOfMonth(foo.t) AS t,
    groupArraySortedIfMergeState(2048)(epoch_map) AS epoch_map,
    min(epoch_start) AS epoch_start,
    max(epoch_end) AS epoch_end,
    argMinStateMerge(balance_start) AS balance_start,
    argMaxStateMerge(balance_end) AS balance_end,
    min(balance_min) AS balance_min,
    max(balance_max) AS balance_max,
    sum(deposits_count) AS deposits_count,
    sum(deposits_amount) AS deposits_amount,
    sum(withdrawals_count) AS withdrawals_count,
    sum(withdrawals_amount) AS withdrawals_amount,
    sum(attestations_scheduled) AS attestations_scheduled,
    sum(attestations_observed) AS attestations_observed,
    sum(attestations_head_matched) AS attestations_head_matched,
    sum(attestations_source_matched) AS attestations_source_matched,
    sum(attestations_target_matched) AS attestations_target_matched,
    sum(attestations_head_executed) AS attestations_head_executed,
    sum(attestations_source_executed) AS attestations_source_executed,
    sum(attestations_target_executed) AS attestations_target_executed,
    sum(attestations_reward) AS attestations_reward,
    sum(attestations_reward_rewards_only) AS attestations_reward_rewards_only,
    sum(attestations_reward_penalties_only) AS attestations_reward_penalties_only,
    sum(attestations_head_reward) AS attestations_head_reward,
    sum(attestations_source_reward) AS attestations_source_reward,
    sum(attestations_target_reward) AS attestations_target_reward,
    sum(attestations_localized_max_reward) AS attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) AS attestations_hyperlocalized_max_reward,
    sum(attestations_inclusion_reward) AS attestations_inclusion_reward,
    sum(attestations_inactivity_reward) AS attestations_inactivity_reward,
    sum(attestations_ideal_reward) AS attestations_ideal_reward,
    sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
    sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
    sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
    sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
    sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
    sum(inclusion_delay_sum) AS inclusion_delay_sum,
    sum(optimal_inclusion_delay_sum) AS optimal_inclusion_delay_sum,
    sum(blocks_scheduled) AS blocks_scheduled,
    sum(blocks_proposed) AS blocks_proposed,
    sum(blocks_cl_reward) AS blocks_cl_reward,
    sum(blocks_cl_attestations_reward) AS blocks_cl_attestations_reward,
    sum(blocks_cl_sync_aggregate_reward) AS blocks_cl_sync_aggregate_reward,
    sum(blocks_cl_slasher_reward) AS blocks_cl_slasher_reward,
    sum(blocks_cl_missed_median_reward) AS blocks_cl_missed_median_reward,
    sum(blocks_slashing_count) AS blocks_slashing_count,
    sum(blocks_expected) AS blocks_expected,
    sum(sync_scheduled) AS sync_scheduled,
    sum(sync_executed) AS sync_executed,
    sum(sync_reward) AS sync_reward,
    sum(sync_reward_rewards_only) AS sync_reward_rewards_only,
    sum(sync_reward_penalties_only) AS sync_reward_penalties_only,
    sum(sync_localized_max_reward) AS sync_localized_max_reward,
    sum(sync_committees_expected) AS sync_committees_expected,
    max(slashed) AS slashed,
    max(last_executed_duty_epoch) AS last_executed_duty_epoch,
    max(last_scheduled_sync_epoch) AS last_scheduled_sync_epoch,
    max(last_scheduled_block_epoch) AS last_scheduled_block_epoch
FROM _final_validator_dashboard_data_daily AS foo
GROUP BY
    t,
    validator_index
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS _final_validator_dashboard_data_hourly_mv
-- +goose StatementEnd
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS _final_validator_dashboard_data_daily_mv
-- +goose StatementEnd
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS _final_validator_dashboard_data_weekly_mv
-- +goose StatementEnd
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS _final_validator_dashboard_data_monthly_mv
-- +goose StatementEnd
