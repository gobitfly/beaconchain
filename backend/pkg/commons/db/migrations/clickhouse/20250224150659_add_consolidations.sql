-- +goose Up
-- +goose StatementBegin
ALTER TABLE _insert_sink_validator_dashboard_data_epoch 
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count Int64 DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount Int64 DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count Int64 DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount Int64 DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_data_epoch 
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count Int64 DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount Int64 DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count Int64 DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount Int64 DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_epoch 
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count Int64 DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount Int64 DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count Int64 DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount Int64 DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily 
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly 
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly 
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_1h
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_1h
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_24h
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_24h
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_7d
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_7d
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_30d
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_30d
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_90d
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_90d
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_total
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_total
    ADD COLUMN IF NOT EXISTS consolidations_incoming_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_incoming_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_count SimpleAggregateFunction(sum, Int64) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consolidations_outgoing_amount SimpleAggregateFunction(sum, Int64) DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _mv_unsafe_validator_dashboard_data_epoch MODIFY QUERY
SELECT
    now() AS _inserted_at,
    validator_index,
    epoch,
    epoch_timestamp,
    balance_effective_start,
    balance_effective_end,
    balance_start,
    balance_end,
    deposits_count,
    deposits_amount,
    withdrawals_count,
    withdrawals_amount,
    attestations_scheduled,
    attestations_observed,
    attestations_head_matched,
    attestations_target_matched,
    attestations_source_matched,
    attestations_head_executed,
    attestations_target_executed,
    attestations_source_executed,
    greatest(attestations_head_reward, 0) AS attestations_head_reward_rewards_only,
    least(attestations_head_reward, 0) AS attestations_head_reward_penalties_only,
    greatest(attestations_target_reward, 0) AS attestations_target_reward_rewards_only,
    least(attestations_target_reward, 0) AS attestations_target_reward_penalties_only,
    greatest(attestations_source_reward, 0) AS attestations_source_reward_rewards_only,
    least(attestations_source_reward, 0) AS attestations_source_reward_penalties_only,
    greatest(attestations_inactivity_reward, 0) AS attestations_inactivity_reward_rewards_only,
    least(attestations_inactivity_reward, 0) AS attestations_inactivity_reward_penalties_only,
    greatest(attestations_inclusion_reward, 0) AS attestations_inclusion_reward_rewards_only,
    least(attestations_inclusion_reward, 0) AS attestations_inclusion_reward_penalties_only,
    attestations_ideal_head_reward,
    attestations_ideal_target_reward,
    attestations_ideal_source_reward,
    attestations_ideal_inactivity_reward,
    attestations_ideal_inclusion_reward,
    attestations_localized_max_reward,
    attestations_hyperlocalized_max_reward,
    inclusion_delay_sum,
    optimal_inclusion_delay_sum,
    length(blocks_status.proposed) AS blocks_scheduled,
    arrayCount(x -> x, blocks_status.proposed) AS blocks_proposed,
    (arraySum(block_rewards.attestations_reward) + arraySum(block_rewards.sync_aggregate_reward)) + arraySum(block_rewards.slasher_reward) AS blocks_cl_reward,
    arraySum(block_rewards.attestations_reward) AS blocks_cl_attestations_reward,
    arraySum(block_rewards.sync_aggregate_reward) AS blocks_cl_sync_aggregate_reward,
    arraySum(block_rewards.slasher_reward) AS blocks_cl_slasher_reward,
    blocks_cl_missed_median_reward,
    blocks_slashing_count,
    blocks_expected,
    sync_scheduled,
    arrayCount(x -> x, sync_status.executed) AS sync_executed,
    arraySum(x -> greatest(0, x), sync_rewards.reward) AS sync_reward_rewards_only,
    arraySum(x -> least(0, x), sync_rewards.reward) AS sync_reward_penalties_only,
    sync_localized_max_reward,
    sync_committees_expected,
    slashed,
    consolidations_incoming_count,
    consolidations_incoming_amount,
    consolidations_outgoing_count,
    consolidations_outgoing_amount
FROM _insert_sink_validator_dashboard_data_epoch
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _mv_final_validator_dashboard_data_hourly MODIFY QUERY 
SELECT 
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
    sum(attestations_head_reward_rewards_only) AS attestations_head_reward_rewards_only,
    sum(attestations_head_reward_penalties_only) AS attestations_head_reward_penalties_only,
    sum(attestations_source_reward_rewards_only) AS attestations_source_reward_rewards_only,
    sum(attestations_source_reward_penalties_only) AS attestations_source_reward_penalties_only,
    sum(attestations_target_reward_rewards_only) AS attestations_target_reward_rewards_only,
    sum(attestations_target_reward_penalties_only) AS attestations_target_reward_penalties_only,
    sum(attestations_inclusion_reward_rewards_only) AS attestations_inclusion_reward_rewards_only,
    sum(attestations_inclusion_reward_penalties_only) AS attestations_inclusion_reward_penalties_only,
    sum(attestations_inactivity_reward_rewards_only) AS attestations_inactivity_reward_rewards_only,
    sum(attestations_inactivity_reward_penalties_only) AS attestations_inactivity_reward_penalties_only,
    sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
    sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
    sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
    sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
    sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
    sum(attestations_localized_max_reward) AS attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) AS attestations_hyperlocalized_max_reward,
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
    sum(sync_reward_rewards_only) AS sync_reward_rewards_only,
    sum(sync_reward_penalties_only) AS sync_reward_penalties_only,
    sum(sync_localized_max_reward) AS sync_localized_max_reward,
    sum(sync_committees_expected) AS sync_committees_expected,
    max(slashed) AS slashed,
    maxIfOrNull(foo.epoch, (foo.blocks_proposed != 0) OR (foo.sync_executed != 0) OR (foo.attestations_observed != 0)) AS last_executed_duty_epoch,
    maxIfOrNull(foo.epoch, foo.sync_scheduled != 0) AS last_scheduled_sync_epoch,
    maxIfOrNull(foo.epoch, foo.blocks_proposed != 0) AS last_scheduled_block_epoch,
    sum(consolidations_incoming_count) AS consolidations_incoming_count,
    sum(consolidations_incoming_amount) AS consolidations_incoming_amount,
    sum(consolidations_outgoing_count) AS consolidations_outgoing_count,
    sum(consolidations_outgoing_amount) AS consolidations_outgoing_amount
FROM _final_validator_dashboard_data_epoch AS foo
GROUP BY
    t,
    validator_index
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _mv_final_validator_dashboard_data_daily MODIFY QUERY
SELECT
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
    sum(attestations_head_reward_rewards_only) AS attestations_head_reward_rewards_only,
    sum(attestations_head_reward_penalties_only) AS attestations_head_reward_penalties_only,
    sum(attestations_source_reward_rewards_only) AS attestations_source_reward_rewards_only,
    sum(attestations_source_reward_penalties_only) AS attestations_source_reward_penalties_only,
    sum(attestations_target_reward_rewards_only) AS attestations_target_reward_rewards_only,
    sum(attestations_target_reward_penalties_only) AS attestations_target_reward_penalties_only,
    sum(attestations_inclusion_reward_rewards_only) AS attestations_inclusion_reward_rewards_only,
    sum(attestations_inclusion_reward_penalties_only) AS attestations_inclusion_reward_penalties_only,
    sum(attestations_inactivity_reward_rewards_only) AS attestations_inactivity_reward_rewards_only,
    sum(attestations_inactivity_reward_penalties_only) AS attestations_inactivity_reward_penalties_only,
    sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
    sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
    sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
    sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
    sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
    sum(attestations_localized_max_reward) AS attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) AS attestations_hyperlocalized_max_reward,
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
    sum(sync_reward_rewards_only) AS sync_reward_rewards_only,
    sum(sync_reward_penalties_only) AS sync_reward_penalties_only,
    sum(sync_localized_max_reward) AS sync_localized_max_reward,
    sum(sync_committees_expected) AS sync_committees_expected,
    max(slashed) AS slashed,
    max(last_executed_duty_epoch) AS last_executed_duty_epoch,
    max(last_scheduled_sync_epoch) AS last_scheduled_sync_epoch,
    max(last_scheduled_block_epoch) AS last_scheduled_block_epoch,
    sum(consolidations_incoming_count) AS consolidations_incoming_count,
    sum(consolidations_incoming_amount) AS consolidations_incoming_amount,
    sum(consolidations_outgoing_count) AS consolidations_outgoing_count,
    sum(consolidations_outgoing_amount) AS consolidations_outgoing_amount
FROM _final_validator_dashboard_data_hourly AS foo
GROUP BY
    t,
    validator_index
-- +goose StatementEnd
-- +goose StatementBegin
-- CREATE OR REPLACE MATERIALIZED VIEW _mv_final_validator_dashboard_data_weekly TO _final_validator_dashboard_data_weekly
ALTER TABLE _mv_final_validator_dashboard_data_weekly MODIFY QUERY
SELECT
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
    sum(attestations_head_reward_rewards_only) AS attestations_head_reward_rewards_only,
    sum(attestations_head_reward_penalties_only) AS attestations_head_reward_penalties_only,
    sum(attestations_source_reward_rewards_only) AS attestations_source_reward_rewards_only,
    sum(attestations_source_reward_penalties_only) AS attestations_source_reward_penalties_only,
    sum(attestations_target_reward_rewards_only) AS attestations_target_reward_rewards_only,
    sum(attestations_target_reward_penalties_only) AS attestations_target_reward_penalties_only,
    sum(attestations_inclusion_reward_rewards_only) AS attestations_inclusion_reward_rewards_only,
    sum(attestations_inclusion_reward_penalties_only) AS attestations_inclusion_reward_penalties_only,
    sum(attestations_inactivity_reward_rewards_only) AS attestations_inactivity_reward_rewards_only,
    sum(attestations_inactivity_reward_penalties_only) AS attestations_inactivity_reward_penalties_only,
    sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
    sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
    sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
    sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
    sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
    sum(attestations_localized_max_reward) AS attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) AS attestations_hyperlocalized_max_reward,
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
    sum(sync_reward_rewards_only) AS sync_reward_rewards_only,
    sum(sync_reward_penalties_only) AS sync_reward_penalties_only,
    sum(sync_localized_max_reward) AS sync_localized_max_reward,
    sum(sync_committees_expected) AS sync_committees_expected,
    max(slashed) AS slashed,
    max(last_executed_duty_epoch) AS last_executed_duty_epoch,
    max(last_scheduled_sync_epoch) AS last_scheduled_sync_epoch,
    max(last_scheduled_block_epoch) AS last_scheduled_block_epoch,
    sum(consolidations_incoming_count) AS consolidations_incoming_count,
    sum(consolidations_incoming_amount) AS consolidations_incoming_amount,
    sum(consolidations_outgoing_count) AS consolidations_outgoing_count,
    sum(consolidations_outgoing_amount) AS consolidations_outgoing_amount
FROM _final_validator_dashboard_data_daily AS foo
GROUP BY
    t,
    validator_index
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _mv_final_validator_dashboard_data_monthly MODIFY QUERY
SELECT
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
    sum(attestations_head_reward_rewards_only) AS attestations_head_reward_rewards_only,
    sum(attestations_head_reward_penalties_only) AS attestations_head_reward_penalties_only,
    sum(attestations_source_reward_rewards_only) AS attestations_source_reward_rewards_only,
    sum(attestations_source_reward_penalties_only) AS attestations_source_reward_penalties_only,
    sum(attestations_target_reward_rewards_only) AS attestations_target_reward_rewards_only,
    sum(attestations_target_reward_penalties_only) AS attestations_target_reward_penalties_only,
    sum(attestations_inclusion_reward_rewards_only) AS attestations_inclusion_reward_rewards_only,
    sum(attestations_inclusion_reward_penalties_only) AS attestations_inclusion_reward_penalties_only,
    sum(attestations_inactivity_reward_rewards_only) AS attestations_inactivity_reward_rewards_only,
    sum(attestations_inactivity_reward_penalties_only) AS attestations_inactivity_reward_penalties_only,
    sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
    sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
    sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
    sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
    sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
    sum(attestations_localized_max_reward) AS attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) AS attestations_hyperlocalized_max_rewafrd,
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
    sum(sync_reward_rewards_only) AS sync_reward_rewards_only,
    sum(sync_reward_penalties_only) AS sync_reward_penalties_only,
    sum(sync_localized_max_reward) AS sync_localized_max_reward,
    sum(sync_committees_expected) AS sync_committees_expected,
    max(slashed) AS slashed,
    max(last_executed_duty_epoch) AS last_executed_duty_epoch,
    max(last_scheduled_sync_epoch) AS last_scheduled_sync_epoch,
    max(last_scheduled_block_epoch) AS last_scheduled_block_epoch,
    sum(consolidations_incoming_count) AS consolidations_incoming_count,
    sum(consolidations_incoming_amount) AS consolidations_incoming_amount,
    sum(consolidations_outgoing_count) AS consolidations_outgoing_count,
    sum(consolidations_outgoing_amount) AS consolidations_outgoing_amount
FROM _final_validator_dashboard_data_daily AS foo -- NOT weekly because weeks are not aligned with months
GROUP BY
    t,
    validator_index
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_epoch
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    (((attestations_head_reward_rewards_only + attestations_source_reward_rewards_only) + attestations_target_reward_rewards_only) + attestations_inclusion_reward_rewards_only) + attestations_inactivity_reward_rewards_only AS attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_data_epoch
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_hourly
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    (((attestations_head_reward_rewards_only + attestations_source_reward_rewards_only) + attestations_target_reward_rewards_only) + attestations_inclusion_reward_rewards_only) + attestations_inactivity_reward_rewards_only AS attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_data_hourly FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_daily
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    (((attestations_head_reward_rewards_only + attestations_source_reward_rewards_only) + attestations_target_reward_rewards_only) + attestations_inclusion_reward_rewards_only) + attestations_inactivity_reward_rewards_only AS attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_data_daily FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_weekly
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    (((attestations_head_reward_rewards_only + attestations_source_reward_rewards_only) + attestations_target_reward_rewards_only) + attestations_inclusion_reward_rewards_only) + attestations_inactivity_reward_rewards_only AS attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_data_weekly FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_monthly
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    (((attestations_head_reward_rewards_only + attestations_source_reward_rewards_only) + attestations_target_reward_rewards_only) + attestations_inclusion_reward_rewards_only) + attestations_inactivity_reward_rewards_only AS attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_data_monthly FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_1h
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    (((attestations_head_reward_rewards_only + attestations_source_reward_rewards_only) + attestations_target_reward_rewards_only) + attestations_inclusion_reward_rewards_only) + attestations_inactivity_reward_rewards_only AS attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_rolling_1h FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_24h
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    (((attestations_head_reward_rewards_only + attestations_source_reward_rewards_only) + attestations_target_reward_rewards_only) + attestations_inclusion_reward_rewards_only) + attestations_inactivity_reward_rewards_only AS attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_rolling_24h FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_7d
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    (((attestations_head_reward_rewards_only + attestations_source_reward_rewards_only) + attestations_target_reward_rewards_only) + attestations_inclusion_reward_rewards_only) + attestations_inactivity_reward_rewards_only AS attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_rolling_7d FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_30d
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    (((attestations_head_reward_rewards_only + attestations_source_reward_rewards_only) + attestations_target_reward_rewards_only) + attestations_inclusion_reward_rewards_only) + attestations_inactivity_reward_rewards_only AS attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_rolling_30d FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_90d
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    (((attestations_head_reward_rewards_only + attestations_source_reward_rewards_only) + attestations_target_reward_rewards_only) + attestations_inclusion_reward_rewards_only) + attestations_inactivity_reward_rewards_only AS attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_rolling_90d FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_total
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    (((attestations_head_reward_rewards_only + attestations_source_reward_rewards_only) + attestations_target_reward_rewards_only) + attestations_inclusion_reward_rewards_only) + attestations_inactivity_reward_rewards_only AS attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_rolling_total FINAL
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd