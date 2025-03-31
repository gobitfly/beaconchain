-- +goose Up
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_1h as
select 
    validator_index,
    max(t) as t,
    groupArraySortedIfMergeState(2048)(epoch_map),
    min(epoch_start) as epoch_start,
    max(epoch_end) as epoch_end,
    argMinMergeState(balance_start) as balance_start,
    argMaxMergeState(balance_end) as balance_end,
    min(balance_min) as balance_min,
    max(balance_max) as balance_max,
    sum(deposits_count) as deposits_count,
    sum(deposits_amount) as deposits_amount,
    sum(withdrawals_count) as withdrawals_count,
    sum(withdrawals_amount) as withdrawals_amount,
    sum(attestations_scheduled) as attestations_scheduled,
    sum(attestations_observed) as attestations_observed,
    sum(attestations_head_matched) as attestations_head_matched,
    sum(attestations_target_matched) as attestations_target_matched,
    sum(attestations_source_matched) as attestations_source_matched,
    sum(attestations_head_executed) as attestations_head_executed,
    sum(attestations_target_executed) as attestations_target_executed,
    sum(attestations_source_executed) as attestations_source_executed,
    sum(attestations_head_reward_rewards_only) as attestations_head_reward_rewards_only,
    sum(attestations_head_reward_penalties_only) as attestations_head_reward_penalties_only,
    sum(attestations_target_reward_rewards_only) as attestations_target_reward_rewards_only,
    sum(attestations_target_reward_penalties_only) as attestations_target_reward_penalties_only,
    sum(attestations_source_reward_rewards_only) as attestations_source_reward_rewards_only,
    sum(attestations_source_reward_penalties_only) as attestations_source_reward_penalties_only,
    sum(attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward_rewards_only,
    sum(attestations_inactivity_reward_penalties_only) as attestations_inactivity_reward_penalties_only,
    sum(attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward_rewards_only,
    sum(attestations_inclusion_reward_penalties_only) as attestations_inclusion_reward_penalties_only,
    sum(attestations_ideal_head_reward) as attestations_ideal_head_reward,
    sum(attestations_ideal_target_reward) as attestations_ideal_target_reward,
    sum(attestations_ideal_source_reward) as attestations_ideal_source_reward,
    sum(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
    sum(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
    sum(attestations_localized_max_reward) as attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) as attestations_hyperlocalized_max_reward,
    sum(inclusion_delay_sum) as inclusion_delay_sum,
    sum(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum,
    sum(blocks_scheduled) as blocks_scheduled,
    sum(blocks_proposed) as blocks_proposed,
    sum(blocks_cl_reward) as blocks_cl_reward,
    sum(blocks_cl_attestations_reward) as blocks_cl_attestations_reward,
    sum(blocks_cl_sync_aggregate_reward) as blocks_cl_sync_aggregate_reward,
    sum(blocks_cl_slasher_reward) as blocks_cl_slasher_reward,
    sum(blocks_cl_missed_median_reward) as blocks_cl_missed_median_reward,
    sum(blocks_slashing_count) as blocks_slashing_count,
    sum(blocks_expected) as blocks_expected,
    sum(sync_scheduled) as sync_scheduled,
    sum(sync_executed) as sync_executed,
    sum(sync_reward_rewards_only) as sync_reward_rewards_only,
    sum(sync_reward_penalties_only) as sync_reward_penalties_only,
    sum(sync_localized_max_reward) as sync_localized_max_reward,
    sum(sync_committees_expected) as sync_committees_expected,
    max(slashed) as slashed,
    sum(last_executed_duty_epoch) last_executed_duty_epoch,
    sum(last_scheduled_sync_epoch) last_scheduled_sync_epoch,
    sum(last_scheduled_block_epoch) last_scheduled_block_epoch,
    sum(attestations_reward) as attestations_reward,
    sum(attestations_ideal_reward) as attestations_ideal_reward,
    sum(foo.attestations_head_reward_penalties_only + foo.attestations_head_reward_rewards_only) as attestations_head_reward,
    sum(foo.attestations_source_reward_penalties_only + foo.attestations_source_reward_rewards_only) as attestations_source_reward,
    sum(foo.attestations_target_reward_penalties_only + foo.attestations_target_reward_rewards_only) as attestations_target_reward,
    sum(foo.attestations_inclusion_reward_penalties_only + foo.attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward,
    sum(foo.attestations_inactivity_reward_penalties_only + foo.attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward,
    sum(
        foo.attestations_head_reward_rewards_only +
        foo.attestations_source_reward_rewards_only +
        foo.attestations_target_reward_rewards_only +
        foo.attestations_inclusion_reward_rewards_only +
        foo.attestations_inactivity_reward_rewards_only
    ) as attestations_reward_rewards_only,
    sum(
        foo.attestations_head_reward_penalties_only +
        foo.attestations_source_reward_penalties_only +
        foo.attestations_target_reward_penalties_only +
        foo.attestations_inclusion_reward_penalties_only +
        foo.attestations_inactivity_reward_penalties_only
    ) as attestations_reward_penalties_only,
    sum(foo.sync_reward_rewards_only + foo.sync_reward_penalties_only) as sync_reward
    from _final_validator_dashboard_rolling_1h as foo GROUP BY validator_index
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_24h as 
select 
    validator_index,
    max(t) as t,
    groupArraySortedIfMergeState(2048)(epoch_map),
    min(epoch_start) as epoch_start,
    max(epoch_end) as epoch_end,
    argMinMergeState(balance_start) as balance_start,
    argMaxMergeState(balance_end) as balance_end,
    min(balance_min) as balance_min,
    max(balance_max) as balance_max,
    sum(deposits_count) as deposits_count,
    sum(deposits_amount) as deposits_amount,
    sum(withdrawals_count) as withdrawals_count,
    sum(withdrawals_amount) as withdrawals_amount,
    sum(attestations_scheduled) as attestations_scheduled,
    sum(attestations_observed) as attestations_observed,
    sum(attestations_head_matched) as attestations_head_matched,
    sum(attestations_target_matched) as attestations_target_matched,
    sum(attestations_source_matched) as attestations_source_matched,
    sum(attestations_head_executed) as attestations_head_executed,
    sum(attestations_target_executed) as attestations_target_executed,
    sum(attestations_source_executed) as attestations_source_executed,
    sum(attestations_head_reward_rewards_only) as attestations_head_reward_rewards_only,
    sum(attestations_head_reward_penalties_only) as attestations_head_reward_penalties_only,
    sum(attestations_target_reward_rewards_only) as attestations_target_reward_rewards_only,
    sum(attestations_target_reward_penalties_only) as attestations_target_reward_penalties_only,
    sum(attestations_source_reward_rewards_only) as attestations_source_reward_rewards_only,
    sum(attestations_source_reward_penalties_only) as attestations_source_reward_penalties_only,
    sum(attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward_rewards_only,
    sum(attestations_inactivity_reward_penalties_only) as attestations_inactivity_reward_penalties_only,
    sum(attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward_rewards_only,
    sum(attestations_inclusion_reward_penalties_only) as attestations_inclusion_reward_penalties_only,
    sum(attestations_ideal_head_reward) as attestations_ideal_head_reward,
    sum(attestations_ideal_target_reward) as attestations_ideal_target_reward,
    sum(attestations_ideal_source_reward) as attestations_ideal_source_reward,
    sum(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
    sum(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
    sum(attestations_localized_max_reward) as attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) as attestations_hyperlocalized_max_reward,
    sum(inclusion_delay_sum) as inclusion_delay_sum,
    sum(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum,
    sum(blocks_scheduled) as blocks_scheduled,
    sum(blocks_proposed) as blocks_proposed,
    sum(blocks_cl_reward) as blocks_cl_reward,
    sum(blocks_cl_attestations_reward) as blocks_cl_attestations_reward,
    sum(blocks_cl_sync_aggregate_reward) as blocks_cl_sync_aggregate_reward,
    sum(blocks_cl_slasher_reward) as blocks_cl_slasher_reward,
    sum(blocks_cl_missed_median_reward) as blocks_cl_missed_median_reward,
    sum(blocks_slashing_count) as blocks_slashing_count,
    sum(blocks_expected) as blocks_expected,
    sum(sync_scheduled) as sync_scheduled,
    sum(sync_executed) as sync_executed,
    sum(sync_reward_rewards_only) as sync_reward_rewards_only,
    sum(sync_reward_penalties_only) as sync_reward_penalties_only,
    sum(sync_localized_max_reward) as sync_localized_max_reward,
    sum(sync_committees_expected) as sync_committees_expected,
    max(slashed) as slashed,
    sum(last_executed_duty_epoch) last_executed_duty_epoch,
    sum(last_scheduled_sync_epoch) last_scheduled_sync_epoch,
    sum(last_scheduled_block_epoch) last_scheduled_block_epoch,
    sum(attestations_reward) as attestations_reward,
    sum(attestations_ideal_reward) as attestations_ideal_reward,
    sum(foo.attestations_head_reward_penalties_only + foo.attestations_head_reward_rewards_only) as attestations_head_reward,
    sum(foo.attestations_source_reward_penalties_only + foo.attestations_source_reward_rewards_only) as attestations_source_reward,
    sum(foo.attestations_target_reward_penalties_only + foo.attestations_target_reward_rewards_only) as attestations_target_reward,
    sum(foo.attestations_inclusion_reward_penalties_only + foo.attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward,
    sum(foo.attestations_inactivity_reward_penalties_only + foo.attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward,
    sum(
        foo.attestations_head_reward_rewards_only +
        foo.attestations_source_reward_rewards_only +
        foo.attestations_target_reward_rewards_only +
        foo.attestations_inclusion_reward_rewards_only +
        foo.attestations_inactivity_reward_rewards_only
    ) as attestations_reward_rewards_only,
    sum(
        foo.attestations_head_reward_penalties_only +
        foo.attestations_source_reward_penalties_only +
        foo.attestations_target_reward_penalties_only +
        foo.attestations_inclusion_reward_penalties_only +
        foo.attestations_inactivity_reward_penalties_only
    ) as attestations_reward_penalties_only,
    sum(foo.sync_reward_rewards_only + foo.sync_reward_penalties_only) as sync_reward
    from _final_validator_dashboard_rolling_24h as foo GROUP BY validator_index
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_7d as 
select 
    validator_index,
    max(t) as t,
    groupArraySortedIfMergeState(2048)(epoch_map),
    min(epoch_start) as epoch_start,
    max(epoch_end) as epoch_end,
    argMinMergeState(balance_start) as balance_start,
    argMaxMergeState(balance_end) as balance_end,
    min(balance_min) as balance_min,
    max(balance_max) as balance_max,
    sum(deposits_count) as deposits_count,
    sum(deposits_amount) as deposits_amount,
    sum(withdrawals_count) as withdrawals_count,
    sum(withdrawals_amount) as withdrawals_amount,
    sum(attestations_scheduled) as attestations_scheduled,
    sum(attestations_observed) as attestations_observed,
    sum(attestations_head_matched) as attestations_head_matched,
    sum(attestations_target_matched) as attestations_target_matched,
    sum(attestations_source_matched) as attestations_source_matched,
    sum(attestations_head_executed) as attestations_head_executed,
    sum(attestations_target_executed) as attestations_target_executed,
    sum(attestations_source_executed) as attestations_source_executed,
    sum(attestations_head_reward_rewards_only) as attestations_head_reward_rewards_only,
    sum(attestations_head_reward_penalties_only) as attestations_head_reward_penalties_only,
    sum(attestations_target_reward_rewards_only) as attestations_target_reward_rewards_only,
    sum(attestations_target_reward_penalties_only) as attestations_target_reward_penalties_only,
    sum(attestations_source_reward_rewards_only) as attestations_source_reward_rewards_only,
    sum(attestations_source_reward_penalties_only) as attestations_source_reward_penalties_only,
    sum(attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward_rewards_only,
    sum(attestations_inactivity_reward_penalties_only) as attestations_inactivity_reward_penalties_only,
    sum(attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward_rewards_only,
    sum(attestations_inclusion_reward_penalties_only) as attestations_inclusion_reward_penalties_only,
    sum(attestations_ideal_head_reward) as attestations_ideal_head_reward,
    sum(attestations_ideal_target_reward) as attestations_ideal_target_reward,
    sum(attestations_ideal_source_reward) as attestations_ideal_source_reward,
    sum(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
    sum(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
    sum(attestations_localized_max_reward) as attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) as attestations_hyperlocalized_max_reward,
    sum(inclusion_delay_sum) as inclusion_delay_sum,
    sum(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum,
    sum(blocks_scheduled) as blocks_scheduled,
    sum(blocks_proposed) as blocks_proposed,
    sum(blocks_cl_reward) as blocks_cl_reward,
    sum(blocks_cl_attestations_reward) as blocks_cl_attestations_reward,
    sum(blocks_cl_sync_aggregate_reward) as blocks_cl_sync_aggregate_reward,
    sum(blocks_cl_slasher_reward) as blocks_cl_slasher_reward,
    sum(blocks_cl_missed_median_reward) as blocks_cl_missed_median_reward,
    sum(blocks_slashing_count) as blocks_slashing_count,
    sum(blocks_expected) as blocks_expected,
    sum(sync_scheduled) as sync_scheduled,
    sum(sync_executed) as sync_executed,
    sum(sync_reward_rewards_only) as sync_reward_rewards_only,
    sum(sync_reward_penalties_only) as sync_reward_penalties_only,
    sum(sync_localized_max_reward) as sync_localized_max_reward,
    sum(sync_committees_expected) as sync_committees_expected,
    max(slashed) as slashed,
    sum(last_executed_duty_epoch) last_executed_duty_epoch,
    sum(last_scheduled_sync_epoch) last_scheduled_sync_epoch,
    sum(last_scheduled_block_epoch) last_scheduled_block_epoch,
    sum(attestations_reward) as attestations_reward,
    sum(attestations_ideal_reward) as attestations_ideal_reward,
    sum(foo.attestations_head_reward_penalties_only + foo.attestations_head_reward_rewards_only) as attestations_head_reward,
    sum(foo.attestations_source_reward_penalties_only + foo.attestations_source_reward_rewards_only) as attestations_source_reward,
    sum(foo.attestations_target_reward_penalties_only + foo.attestations_target_reward_rewards_only) as attestations_target_reward,
    sum(foo.attestations_inclusion_reward_penalties_only + foo.attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward,
    sum(foo.attestations_inactivity_reward_penalties_only + foo.attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward,
    sum(
        foo.attestations_head_reward_rewards_only +
        foo.attestations_source_reward_rewards_only +
        foo.attestations_target_reward_rewards_only +
        foo.attestations_inclusion_reward_rewards_only +
        foo.attestations_inactivity_reward_rewards_only
    ) as attestations_reward_rewards_only,
    sum(
        foo.attestations_head_reward_penalties_only +
        foo.attestations_source_reward_penalties_only +
        foo.attestations_target_reward_penalties_only +
        foo.attestations_inclusion_reward_penalties_only +
        foo.attestations_inactivity_reward_penalties_only
    ) as attestations_reward_penalties_only,
    sum(foo.sync_reward_rewards_only + foo.sync_reward_penalties_only) as sync_reward
    from _final_validator_dashboard_rolling_7d as foo GROUP BY validator_index
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_30d as 
select
    validator_index,
    max(t) as t,
    groupArraySortedIfMergeState(2048)(epoch_map),
    min(epoch_start) as epoch_start,
    max(epoch_end) as epoch_end,
    argMinMergeState(balance_start) as balance_start,
    argMaxMergeState(balance_end) as balance_end,
    min(balance_min) as balance_min,
    max(balance_max) as balance_max,
    sum(deposits_count) as deposits_count,
    sum(deposits_amount) as deposits_amount,
    sum(withdrawals_count) as withdrawals_count,
    sum(withdrawals_amount) as withdrawals_amount,
    sum(attestations_scheduled) as attestations_scheduled,
    sum(attestations_observed) as attestations_observed,
    sum(attestations_head_matched) as attestations_head_matched,
    sum(attestations_target_matched) as attestations_target_matched,
    sum(attestations_source_matched) as attestations_source_matched,
    sum(attestations_head_executed) as attestations_head_executed,
    sum(attestations_target_executed) as attestations_target_executed,
    sum(attestations_source_executed) as attestations_source_executed,
    sum(attestations_head_reward_rewards_only) as attestations_head_reward_rewards_only,
    sum(attestations_head_reward_penalties_only) as attestations_head_reward_penalties_only,
    sum(attestations_target_reward_rewards_only) as attestations_target_reward_rewards_only,
    sum(attestations_target_reward_penalties_only) as attestations_target_reward_penalties_only,
    sum(attestations_source_reward_rewards_only) as attestations_source_reward_rewards_only,
    sum(attestations_source_reward_penalties_only) as attestations_source_reward_penalties_only,
    sum(attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward_rewards_only,
    sum(attestations_inactivity_reward_penalties_only) as attestations_inactivity_reward_penalties_only,
    sum(attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward_rewards_only,
    sum(attestations_inclusion_reward_penalties_only) as attestations_inclusion_reward_penalties_only,
    sum(attestations_ideal_head_reward) as attestations_ideal_head_reward,
    sum(attestations_ideal_target_reward) as attestations_ideal_target_reward,
    sum(attestations_ideal_source_reward) as attestations_ideal_source_reward,
    sum(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
    sum(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
    sum(attestations_localized_max_reward) as attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) as attestations_hyperlocalized_max_reward,
    sum(inclusion_delay_sum) as inclusion_delay_sum,
    sum(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum,
    sum(blocks_scheduled) as blocks_scheduled,
    sum(blocks_proposed) as blocks_proposed,
    sum(blocks_cl_reward) as blocks_cl_reward,
    sum(blocks_cl_attestations_reward) as blocks_cl_attestations_reward,
    sum(blocks_cl_sync_aggregate_reward) as blocks_cl_sync_aggregate_reward,
    sum(blocks_cl_slasher_reward) as blocks_cl_slasher_reward,
    sum(blocks_cl_missed_median_reward) as blocks_cl_missed_median_reward,
    sum(blocks_slashing_count) as blocks_slashing_count,
    sum(blocks_expected) as blocks_expected,
    sum(sync_scheduled) as sync_scheduled,
    sum(sync_executed) as sync_executed,
    sum(sync_reward_rewards_only) as sync_reward_rewards_only,
    sum(sync_reward_penalties_only) as sync_reward_penalties_only,
    sum(sync_localized_max_reward) as sync_localized_max_reward,
    sum(sync_committees_expected) as sync_committees_expected,
    max(slashed) as slashed,
    sum(last_executed_duty_epoch) last_executed_duty_epoch,
    sum(last_scheduled_sync_epoch) last_scheduled_sync_epoch,
    sum(last_scheduled_block_epoch) last_scheduled_block_epoch,
    sum(attestations_reward) as attestations_reward,
    sum(attestations_ideal_reward) as attestations_ideal_reward,
    sum(foo.attestations_head_reward_penalties_only + foo.attestations_head_reward_rewards_only) as attestations_head_reward,
    sum(foo.attestations_source_reward_penalties_only + foo.attestations_source_reward_rewards_only) as attestations_source_reward,
    sum(foo.attestations_target_reward_penalties_only + foo.attestations_target_reward_rewards_only) as attestations_target_reward,
    sum(foo.attestations_inclusion_reward_penalties_only + foo.attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward,
    sum(foo.attestations_inactivity_reward_penalties_only + foo.attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward,
    sum(
        foo.attestations_head_reward_rewards_only +
        foo.attestations_source_reward_rewards_only +
        foo.attestations_target_reward_rewards_only +
        foo.attestations_inclusion_reward_rewards_only +
        foo.attestations_inactivity_reward_rewards_only
    ) as attestations_reward_rewards_only,
    sum(
        foo.attestations_head_reward_penalties_only +
        foo.attestations_source_reward_penalties_only +
        foo.attestations_target_reward_penalties_only +
        foo.attestations_inclusion_reward_penalties_only +
        foo.attestations_inactivity_reward_penalties_only
    ) as attestations_reward_penalties_only,
    sum(foo.sync_reward_rewards_only + foo.sync_reward_penalties_only) as sync_reward
    from _final_validator_dashboard_rolling_30d as foo GROUP BY validator_index
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_90d as 
select
    validator_index,
    max(t) as t,
    groupArraySortedIfMergeState(2048)(epoch_map),
    min(epoch_start) as epoch_start,
    max(epoch_end) as epoch_end,
    argMinMergeState(balance_start) as balance_start,
    argMaxMergeState(balance_end) as balance_end,
    min(balance_min) as balance_min,
    max(balance_max) as balance_max,
    sum(deposits_count) as deposits_count,
    sum(deposits_amount) as deposits_amount,
    sum(withdrawals_count) as withdrawals_count,
    sum(withdrawals_amount) as withdrawals_amount,
    sum(attestations_scheduled) as attestations_scheduled,
    sum(attestations_observed) as attestations_observed,
    sum(attestations_head_matched) as attestations_head_matched,
    sum(attestations_target_matched) as attestations_target_matched,
    sum(attestations_source_matched) as attestations_source_matched,
    sum(attestations_head_executed) as attestations_head_executed,
    sum(attestations_target_executed) as attestations_target_executed,
    sum(attestations_source_executed) as attestations_source_executed,
    sum(attestations_head_reward_rewards_only) as attestations_head_reward_rewards_only,
    sum(attestations_head_reward_penalties_only) as attestations_head_reward_penalties_only,
    sum(attestations_target_reward_rewards_only) as attestations_target_reward_rewards_only,
    sum(attestations_target_reward_penalties_only) as attestations_target_reward_penalties_only,
    sum(attestations_source_reward_rewards_only) as attestations_source_reward_rewards_only,
    sum(attestations_source_reward_penalties_only) as attestations_source_reward_penalties_only,
    sum(attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward_rewards_only,
    sum(attestations_inactivity_reward_penalties_only) as attestations_inactivity_reward_penalties_only,
    sum(attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward_rewards_only,
    sum(attestations_inclusion_reward_penalties_only) as attestations_inclusion_reward_penalties_only,
    sum(attestations_ideal_head_reward) as attestations_ideal_head_reward,
    sum(attestations_ideal_target_reward) as attestations_ideal_target_reward,
    sum(attestations_ideal_source_reward) as attestations_ideal_source_reward,
    sum(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
    sum(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
    sum(attestations_localized_max_reward) as attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) as attestations_hyperlocalized_max_reward,
    sum(inclusion_delay_sum) as inclusion_delay_sum,
    sum(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum,
    sum(blocks_scheduled) as blocks_scheduled,
    sum(blocks_proposed) as blocks_proposed,
    sum(blocks_cl_reward) as blocks_cl_reward,
    sum(blocks_cl_attestations_reward) as blocks_cl_attestations_reward,
    sum(blocks_cl_sync_aggregate_reward) as blocks_cl_sync_aggregate_reward,
    sum(blocks_cl_slasher_reward) as blocks_cl_slasher_reward,
    sum(blocks_cl_missed_median_reward) as blocks_cl_missed_median_reward,
    sum(blocks_slashing_count) as blocks_slashing_count,
    sum(blocks_expected) as blocks_expected,
    sum(sync_scheduled) as sync_scheduled,
    sum(sync_executed) as sync_executed,
    sum(sync_reward_rewards_only) as sync_reward_rewards_only,
    sum(sync_reward_penalties_only) as sync_reward_penalties_only,
    sum(sync_localized_max_reward) as sync_localized_max_reward,
    sum(sync_committees_expected) as sync_committees_expected,
    max(slashed) as slashed,
    sum(last_executed_duty_epoch) last_executed_duty_epoch,
    sum(last_scheduled_sync_epoch) last_scheduled_sync_epoch,
    sum(last_scheduled_block_epoch) last_scheduled_block_epoch,
    sum(attestations_reward) as attestations_reward,
    sum(attestations_ideal_reward) as attestations_ideal_reward,
    sum(foo.attestations_head_reward_penalties_only + foo.attestations_head_reward_rewards_only) as attestations_head_reward,
    sum(foo.attestations_source_reward_penalties_only + foo.attestations_source_reward_rewards_only) as attestations_source_reward,
    sum(foo.attestations_target_reward_penalties_only + foo.attestations_target_reward_rewards_only) as attestations_target_reward,
    sum(foo.attestations_inclusion_reward_penalties_only + foo.attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward,
    sum(foo.attestations_inactivity_reward_penalties_only + foo.attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward,
    sum(
        foo.attestations_head_reward_rewards_only +
        foo.attestations_source_reward_rewards_only +
        foo.attestations_target_reward_rewards_only +
        foo.attestations_inclusion_reward_rewards_only +
        foo.attestations_inactivity_reward_rewards_only
    ) as attestations_reward_rewards_only,
    sum(
        foo.attestations_head_reward_penalties_only +
        foo.attestations_source_reward_penalties_only +
        foo.attestations_target_reward_penalties_only +
        foo.attestations_inclusion_reward_penalties_only +
        foo.attestations_inactivity_reward_penalties_only
    ) as attestations_reward_penalties_only,
    sum(foo.sync_reward_rewards_only + foo.sync_reward_penalties_only) as sync_reward
    from _final_validator_dashboard_rolling_90d as foo GROUP BY validator_index
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_total as 
select
    validator_index,
    max(t) as t,
    groupArraySortedIfMergeState(2048)(epoch_map),
    min(epoch_start) as epoch_start,
    max(epoch_end) as epoch_end,
    argMinMergeState(balance_start) as balance_start,
    argMaxMergeState(balance_end) as balance_end,
    min(balance_min) as balance_min,
    max(balance_max) as balance_max,
    sum(deposits_count) as deposits_count,
    sum(deposits_amount) as deposits_amount,
    sum(withdrawals_count) as withdrawals_count,
    sum(withdrawals_amount) as withdrawals_amount,
    sum(attestations_scheduled) as attestations_scheduled,
    sum(attestations_observed) as attestations_observed,
    sum(attestations_head_matched) as attestations_head_matched,
    sum(attestations_target_matched) as attestations_target_matched,
    sum(attestations_source_matched) as attestations_source_matched,
    sum(attestations_head_executed) as attestations_head_executed,
    sum(attestations_target_executed) as attestations_target_executed,
    sum(attestations_source_executed) as attestations_source_executed,
    sum(attestations_head_reward_rewards_only) as attestations_head_reward_rewards_only,
    sum(attestations_head_reward_penalties_only) as attestations_head_reward_penalties_only,
    sum(attestations_target_reward_rewards_only) as attestations_target_reward_rewards_only,
    sum(attestations_target_reward_penalties_only) as attestations_target_reward_penalties_only,
    sum(attestations_source_reward_rewards_only) as attestations_source_reward_rewards_only,
    sum(attestations_source_reward_penalties_only) as attestations_source_reward_penalties_only,
    sum(attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward_rewards_only,
    sum(attestations_inactivity_reward_penalties_only) as attestations_inactivity_reward_penalties_only,
    sum(attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward_rewards_only,
    sum(attestations_inclusion_reward_penalties_only) as attestations_inclusion_reward_penalties_only,
    sum(attestations_ideal_head_reward) as attestations_ideal_head_reward,
    sum(attestations_ideal_target_reward) as attestations_ideal_target_reward,
    sum(attestations_ideal_source_reward) as attestations_ideal_source_reward,
    sum(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
    sum(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
    sum(attestations_localized_max_reward) as attestations_localized_max_reward,
    sum(attestations_hyperlocalized_max_reward) as attestations_hyperlocalized_max_reward,
    sum(inclusion_delay_sum) as inclusion_delay_sum,
    sum(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum,
    sum(blocks_scheduled) as blocks_scheduled,
    sum(blocks_proposed) as blocks_proposed,
    sum(blocks_cl_reward) as blocks_cl_reward,
    sum(blocks_cl_attestations_reward) as blocks_cl_attestations_reward,
    sum(blocks_cl_sync_aggregate_reward) as blocks_cl_sync_aggregate_reward,
    sum(blocks_cl_slasher_reward) as blocks_cl_slasher_reward,
    sum(blocks_cl_missed_median_reward) as blocks_cl_missed_median_reward,
    sum(blocks_slashing_count) as blocks_slashing_count,
    sum(blocks_expected) as blocks_expected,
    sum(sync_scheduled) as sync_scheduled,
    sum(sync_executed) as sync_executed,
    sum(sync_reward_rewards_only) as sync_reward_rewards_only,
    sum(sync_reward_penalties_only) as sync_reward_penalties_only,
    sum(sync_localized_max_reward) as sync_localized_max_reward,
    sum(sync_committees_expected) as sync_committees_expected,
    max(slashed) as slashed,
    sum(last_executed_duty_epoch) last_executed_duty_epoch,
    sum(last_scheduled_sync_epoch) last_scheduled_sync_epoch,
    sum(last_scheduled_block_epoch) last_scheduled_block_epoch,
    sum(attestations_reward) as attestations_reward,
    sum(attestations_ideal_reward) as attestations_ideal_reward,
    sum(foo.attestations_head_reward_penalties_only + foo.attestations_head_reward_rewards_only) as attestations_head_reward,
    sum(foo.attestations_source_reward_penalties_only + foo.attestations_source_reward_rewards_only) as attestations_source_reward,
    sum(foo.attestations_target_reward_penalties_only + foo.attestations_target_reward_rewards_only) as attestations_target_reward,
    sum(foo.attestations_inclusion_reward_penalties_only + foo.attestations_inclusion_reward_rewards_only) as attestations_inclusion_reward,
    sum(foo.attestations_inactivity_reward_penalties_only + foo.attestations_inactivity_reward_rewards_only) as attestations_inactivity_reward,
    sum(
        foo.attestations_head_reward_rewards_only +
        foo.attestations_source_reward_rewards_only +
        foo.attestations_target_reward_rewards_only +
        foo.attestations_inclusion_reward_rewards_only +
        foo.attestations_inactivity_reward_rewards_only
    ) as attestations_reward_rewards_only,
    sum(
        foo.attestations_head_reward_penalties_only +
        foo.attestations_source_reward_penalties_only +
        foo.attestations_target_reward_penalties_only +
        foo.attestations_inclusion_reward_penalties_only +
        foo.attestations_inactivity_reward_penalties_only
    ) as attestations_reward_penalties_only,
    sum(foo.sync_reward_rewards_only + foo.sync_reward_penalties_only) as sync_reward
    from _final_validator_dashboard_rolling_total as foo GROUP BY validator_index
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
