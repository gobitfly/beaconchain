-- +goose Up
-- +goose StatementBegin
ALTER TABLE _insert_sink_validator_dashboard_data_epoch
    ADD COLUMN IF NOT EXISTS balance_effective_attestations Int64;
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
    consolidations_outgoing_amount,
    balance_effective_attestations
FROM _insert_sink_validator_dashboard_data_epoch;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_data_epoch
    ADD COLUMN IF NOT EXISTS balance_effective_attestations Int64 DEFAULT balance_effective_end;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_epoch 
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only Int64 MATERIALIZED attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend Int64 ALIAS attestations_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor Int64 ALIAS attestations_ideal_reward,
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend Int64 ALIAS blocks_cl_reward,
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor Int64 MATERIALIZED blocks_cl_reward + blocks_cl_missed_median_reward,
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend Int64 ALIAS sync_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor Int64 ALIAS sync_localized_max_reward,
    ADD COLUMN IF NOT EXISTS efficiency_dividend Int64 MATERIALIZED efficiency_attestations_dividend + efficiency_proposals_dividend + efficiency_sync_dividend,
    ADD COLUMN IF NOT EXISTS efficiency_divisor Int64 MATERIALIZED efficiency_attestations_divisor + efficiency_proposals_divisor + efficiency_sync_divisor,
    ADD COLUMN IF NOT EXISTS balance_effective_attestations Int64 DEFAULT balance_effective_end,
    ADD COLUMN IF NOT EXISTS sync_reward Int64 MATERIALIZED sync_reward_rewards_only + sync_reward_penalties_only,
    ADD COLUMN IF NOT EXISTS roi_dividend Int64 MATERIALIZED balance_end - deposits_amount + withdrawals_amount - consolidations_incoming_amount + consolidations_outgoing_amount,
    ADD COLUMN IF NOT EXISTS roi_divisor Int64 MATERIALIZED balance_start settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64) MATERIALIZED attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64) ALIAS attestations_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64) ALIAS attestations_ideal_reward,
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64) ALIAS blocks_cl_reward,
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64) MATERIALIZED blocks_cl_reward + blocks_cl_missed_median_reward,
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64) ALIAS sync_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64) ALIAS sync_localized_max_reward,
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64) MATERIALIZED sync_reward_rewards_only + sync_reward_penalties_only,
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64) MATERIALIZED efficiency_attestations_dividend + efficiency_proposals_dividend + efficiency_sync_dividend,
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64) MATERIALIZED efficiency_attestations_divisor + efficiency_proposals_divisor + efficiency_sync_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64) MATERIALIZED attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64) ALIAS attestations_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64) ALIAS attestations_ideal_reward,
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64) ALIAS blocks_cl_reward,
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64) MATERIALIZED blocks_cl_reward + blocks_cl_missed_median_reward,
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64) ALIAS sync_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64) ALIAS sync_localized_max_reward,
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64) MATERIALIZED sync_reward_rewards_only + sync_reward_penalties_only,
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64) MATERIALIZED efficiency_attestations_dividend + efficiency_proposals_dividend + efficiency_sync_dividend,
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64) MATERIALIZED efficiency_attestations_divisor + efficiency_proposals_divisor + efficiency_sync_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64) MATERIALIZED attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64) ALIAS attestations_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64) ALIAS attestations_ideal_reward,
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64) ALIAS blocks_cl_reward,
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64) MATERIALIZED blocks_cl_reward + blocks_cl_missed_median_reward,
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64) ALIAS sync_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64) ALIAS sync_localized_max_reward,
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64) MATERIALIZED sync_reward_rewards_only + sync_reward_penalties_only,
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64) MATERIALIZED efficiency_attestations_dividend + efficiency_proposals_dividend + efficiency_sync_dividend,
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64) MATERIALIZED efficiency_attestations_divisor + efficiency_proposals_divisor + efficiency_sync_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64) MATERIALIZED attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64) ALIAS attestations_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64) ALIAS attestations_ideal_reward,
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64) ALIAS blocks_cl_reward,
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64) MATERIALIZED blocks_cl_reward + blocks_cl_missed_median_reward,
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64) ALIAS sync_reward_rewards_only,
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64) ALIAS sync_localized_max_reward,
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64) MATERIALIZED sync_reward_rewards_only + sync_reward_penalties_only,
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64) MATERIALIZED efficiency_attestations_dividend + efficiency_proposals_dividend + efficiency_sync_dividend,
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64) MATERIALIZED efficiency_attestations_divisor + efficiency_proposals_divisor + efficiency_sync_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_1h
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_1h
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_24h
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_24h
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_7d
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_7d
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_30d
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_30d
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_90d
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_90d
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_total
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_total
    ADD COLUMN IF NOT EXISTS attestations_reward_rewards_only SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_attestations_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_proposals_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_sync_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS sync_reward SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_dividend SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS efficiency_divisor SimpleAggregateFunction(sum, Int64),
    ADD COLUMN IF NOT EXISTS roi_dividend SimpleAggregateFunction(sum, Int128),
    ADD COLUMN IF NOT EXISTS roi_divisor SimpleAggregateFunction(sum, Int128) settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _final_validator_dashboard_roi_hourly
(
    `validator_index` UInt64 COMMENT 'validator index',
    `t` DateTime COMMENT 'timestamp of the aggregated data',
    roi_dividend SimpleAggregateFunction(sum, Int128) COMMENT 'sum of roi_dividend',
    roi_divisor SimpleAggregateFunction(sum, Int128) COMMENT 'sum of roi_divisor'
)
ENGINE = AggregatingMergeTree
PARTITION BY toStartOfMonth(t)
ORDER BY (toStartOfDay(t), validator_index, t)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048 settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_final_validator_dashboard_roi_hourly TO _final_validator_dashboard_roi_hourly
AS SELECT
    validator_index as validator_index,
    toStartOfHour(epoch_timestamp) as t,
    sum(foo.roi_dividend::Int128) as roi_dividend,
    sum(foo.roi_divisor::Int128) as roi_divisor
FROM _final_validator_dashboard_data_epoch foo
GROUP BY
    t,
    validator_index settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _final_validator_dashboard_roi_daily
(
    `validator_index` UInt64 COMMENT 'validator index',
    `t` DateTime COMMENT 'timestamp of the aggregated data',
    roi_dividend SimpleAggregateFunction(sum, Int128) COMMENT 'sum of roi_dividend',
    roi_divisor SimpleAggregateFunction(sum, Int128) COMMENT 'sum of roi_divisor'
)
ENGINE = AggregatingMergeTree
PARTITION BY toStartOfYear(t)
ORDER BY (toStartOfMonth(t), validator_index, t)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048 settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_final_validator_dashboard_roi_daily TO _final_validator_dashboard_roi_daily
AS SELECT
    validator_index as validator_index,
    toStartOfDay(foo.t) as t,
    sum(foo.roi_dividend::Int128) as roi_dividend,
    sum(foo.roi_divisor::Int128) as roi_divisor
FROM _final_validator_dashboard_roi_hourly as foo
GROUP BY
    t,
    validator_index settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _final_validator_dashboard_roi_weekly
(
    `validator_index` UInt64 COMMENT 'validator index',
    `t` DateTime COMMENT 'timestamp of the aggregated data',
    roi_dividend SimpleAggregateFunction(sum, Int128) COMMENT 'sum of roi_dividend',
    roi_divisor SimpleAggregateFunction(sum, Int128) COMMENT 'sum of roi_divisor'
)
ENGINE = AggregatingMergeTree
PARTITION BY toStartOfInterval(t, INTERVAL 3 YEARS)
ORDER BY (toStartOfInterval(t, INTERVAL 6 MONTHS), validator_index, t)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048 settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_final_validator_dashboard_roi_weekly TO _final_validator_dashboard_roi_weekly
AS SELECT
    validator_index as validator_index,
    toMonday(foo.t) AS t,
    sum(foo.roi_dividend::Int128) as roi_dividend,
    sum(foo.roi_divisor::Int128) as roi_divisor
FROM _final_validator_dashboard_roi_daily as foo
GROUP BY
    t,
    validator_index settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _final_validator_dashboard_roi_monthly
(
    `validator_index` UInt64 COMMENT 'validator index',
    `t` DateTime COMMENT 'timestamp of the aggregated data',
    roi_dividend SimpleAggregateFunction(sum, Int128) COMMENT 'sum of roi_dividend',
    roi_divisor SimpleAggregateFunction(sum, Int128) COMMENT 'sum of roi_divisor'
)
ENGINE = AggregatingMergeTree
ORDER BY (toStartOfYear(t), validator_index, t)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048 settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_final_validator_dashboard_roi_monthly TO _final_validator_dashboard_roi_monthly
AS SELECT
    validator_index as validator_index,
    toStartOfMonth(foo.t) as t,
    sum(foo.roi_dividend::Int128) as roi_dividend,
    sum(foo.roi_divisor::Int128) as roi_divisor
FROM _final_validator_dashboard_roi_daily as foo
GROUP BY
    t,
    validator_index settings mutations_sync=2;
-- +goose StatementEnd
-- INSERT INTO _final_validator_dashboard_roi_hourly
-- SELECT
--     validator_index,
--     toStartOfHour(epoch_timestamp) as t,
--     sum(roi_dividend) as roi_dividend,
--     sum(roi_divisor) as roi_divisor
-- FROM _final_validator_dashboard_data_epoch
-- GROUP BY
--     t,
--     validator_index settings mutations_sync=2;

-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_hourly
AS SELECT
    *,
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only AS attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only AS attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only AS attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only AS attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only AS attestations_inactivity_reward,
    attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    efficiency_attestations_dividend,
    efficiency_attestations_divisor,
    efficiency_proposals_dividend,
    efficiency_proposals_divisor,
    efficiency_sync_dividend,
    efficiency_sync_divisor,
    sync_reward,
    efficiency_dividend,
    efficiency_divisor
FROM _final_validator_dashboard_data_hourly FINAL settings mutations_sync=2
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
    attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    efficiency_attestations_dividend,
    efficiency_attestations_divisor,
    efficiency_proposals_dividend,
    efficiency_proposals_divisor,
    efficiency_sync_dividend,
    efficiency_sync_divisor,
    sync_reward,
    efficiency_dividend,
    efficiency_divisor
FROM _final_validator_dashboard_data_daily FINAL settings mutations_sync=2
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
    attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    efficiency_attestations_dividend,
    efficiency_attestations_divisor,
    efficiency_proposals_dividend,
    efficiency_proposals_divisor,
    efficiency_sync_dividend,
    efficiency_sync_divisor,
    sync_reward,
    efficiency_dividend,
    efficiency_divisor
FROM _final_validator_dashboard_data_weekly FINAL settings mutations_sync=2
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
    attestations_reward_rewards_only,
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
    efficiency_attestations_dividend,
    efficiency_attestations_divisor,
    efficiency_proposals_dividend,
    efficiency_proposals_divisor,
    efficiency_sync_dividend,
    efficiency_sync_divisor,
    sync_reward,
    efficiency_dividend,
    efficiency_divisor
FROM _final_validator_dashboard_data_monthly FINAL settings mutations_sync=2
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
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward,
FROM _final_validator_dashboard_rolling_1h FINAL settings mutations_sync=2
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
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward
FROM _final_validator_dashboard_rolling_24h FINAL settings mutations_sync=2
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
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward
FROM _final_validator_dashboard_rolling_7d FINAL settings mutations_sync=2
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
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward
FROM _final_validator_dashboard_rolling_30d FINAL settings mutations_sync=2
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
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward
FROM _final_validator_dashboard_rolling_90d FINAL settings mutations_sync=2
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
    (((attestations_head_reward_penalties_only + attestations_source_reward_penalties_only) + attestations_target_reward_penalties_only) + attestations_inclusion_reward_penalties_only) + attestations_inactivity_reward_penalties_only AS attestations_reward_penalties_only,
    attestations_reward,
    attestations_ideal_reward
FROM _final_validator_dashboard_rolling_total FINAL settings mutations_sync=2
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_1h
AS SELECT
    * EXCEPT (attestations_reward_rewards_only, sync_reward),
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
FROM _final_validator_dashboard_rolling_1h FINAL settings mutations_sync=2 
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_24h
AS SELECT
    * EXCEPT (attestations_reward_rewards_only, sync_reward),
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
FROM _final_validator_dashboard_rolling_24h FINAL settings mutations_sync=2
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_7d
AS SELECT
    * EXCEPT (attestations_reward_rewards_only, sync_reward),
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
FROM _final_validator_dashboard_rolling_7d FINAL settings mutations_sync=2
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_30d
AS SELECT
    * EXCEPT (attestations_reward_rewards_only, sync_reward),
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
FROM _final_validator_dashboard_rolling_30d FINAL settings mutations_sync=2
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_90d
AS SELECT
    * EXCEPT (attestations_reward_rewards_only, sync_reward),
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
FROM _final_validator_dashboard_rolling_90d FINAL settings mutations_sync=2
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_rolling_total
AS SELECT
    * EXCEPT (attestations_reward_rewards_only, sync_reward),
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
FROM _final_validator_dashboard_rolling_total FINAL settings mutations_sync=2
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_hourly
AS SELECT
    * EXCEPT (attestations_reward_rewards_only, sync_reward),
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
FROM _final_validator_dashboard_data_hourly FINAL settings mutations_sync=2
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_daily
AS SELECT
    * EXCEPT (attestations_reward_rewards_only, sync_reward),
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
FROM _final_validator_dashboard_data_daily FINAL settings mutations_sync=2
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_weekly
AS SELECT
    * EXCEPT (attestations_reward_rewards_only, sync_reward),
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
FROM _final_validator_dashboard_data_weekly FINAL settings mutations_sync=2
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_data_monthly
AS SELECT
    * EXCEPT (attestations_reward_rewards_only, sync_reward),
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
FROM _final_validator_dashboard_data_monthly FINAL settings mutations_sync=2
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW IF EXISTS _mv_final_validator_dashboard_roi_monthly settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW IF EXISTS _mv_final_validator_dashboard_roi_weekly settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW IF EXISTS _mv_final_validator_dashboard_roi_daily settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW IF EXISTS _mv_final_validator_dashboard_roi_hourly settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _final_validator_dashboard_roi_monthly settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _final_validator_dashboard_roi_weekly settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _final_validator_dashboard_roi_daily settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _final_validator_dashboard_roi_hourly settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_1h
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_1h
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_24h
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_24h
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_7d
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_7d
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_30d
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_30d
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_90d
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_90d
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_total
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_total
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS sync_reward,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_epoch
    DROP COLUMN IF EXISTS roi_dividend,
    DROP COLUMN IF EXISTS roi_divisor,
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS sync_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS sync_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS sync_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS sync_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS sync_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_data_epoch
    DROP COLUMN IF EXISTS efficiency_dividend,
    DROP COLUMN IF EXISTS efficiency_divisor,
    DROP COLUMN IF EXISTS efficiency_attestations_dividend,
    DROP COLUMN IF EXISTS efficiency_attestations_divisor,
    DROP COLUMN IF EXISTS efficiency_proposals_dividend,
    DROP COLUMN IF EXISTS efficiency_proposals_divisor,
    DROP COLUMN IF EXISTS efficiency_sync_dividend,
    DROP COLUMN IF EXISTS efficiency_sync_divisor,
    DROP COLUMN IF EXISTS attestations_reward_rewards_only,
    DROP COLUMN IF EXISTS sync_reward
settings mutations_sync=2;
-- +goose StatementEnd