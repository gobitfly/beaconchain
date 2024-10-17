-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _exporter_metadata (
    `epoch` Int64 COMMENT 'epoch number' CODEC(DoubleDelta, ZSTD(8)),
    `insert_batch_id` Int64 COMMENT 'id of the batch the epoch is part of during the insert process' CODEC(T64, ZSTD(8)),
    `successful_insert` Bool COMMENT 'if the batch was successfully inserted. This is set after the exporter has received confirmation from the clickhouse server' CODEC(T64, ZSTD(8)),
    `transfer_batch_id` Int64 COMMENT 'id of the batch the epoch is part of during the transfer to the final table' CODEC(T64, ZSTD(8)),
    `successful_transfer` Bool COMMENT 'if the batch was successfully transferred to the final table. This is set after the exporter has received confirmation from the clickhouse server' CODEC(T64, ZSTD(8)),
)
ENGINE = ReplacingMergeTree
ORDER BY (epoch)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _exporter_tasks (
    `hostname` LowCardinality(String),
    `uuid` UUID DEFAULT generateUUIDv4(),
    `priority` Int64,
    `start_ts` DateTime,
    `end_ts` DateTime,
    `status` Enum('pending', 'running', 'completed'),
)
ENGINE = ReplacingMergeTree
Order By (hostname, priority, start_ts) -- order by start_ts DESC to get the latest task firs
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048;
-- +goose StatementEnd
-- +goose StatementBegin
SET flatten_nested = 1;
-- +goose StatementEnd
-- +goose StatementBegin
--- create sink table that will only be used as a sink for the insert process - materialized views will be to redirect the data to the correct tables
-- uses Null Engine (https://clickhouse.tech/docs/en/engines/table-engines/special/null/)
CREATE TABLE IF NOT EXISTS _insert_sink_validator_dashboard_data_epoch
(
    `validator_index` UInt64 COMMENT 'validator index',
    `epoch` Int64 COMMENT 'epoch number',
    `epoch_timestamp` DateTime COMMENT 'timestamp of the first slot of the epoch',
    `balance_effective_start` Int64 DEFAULT -1 COMMENT 'effective balance at the first slot of the current epoch',
    `balance_effective_end` Int64 DEFAULT -1 COMMENT 'effective balance at the last slot of the current epoch',
    `balance_start` Int64 COMMENT 'balance at the last slot of the previous epoch',
    `balance_end` Int64 COMMENT 'balance at the last slot of the current epoch',
    `deposits_count` Int64 COMMENT 'number of deposits',
    `deposits_amount` Int64 COMMENT 'total amount of deposits',
    `withdrawals_count` Int64 COMMENT 'number of withdrawals',
    `withdrawals_amount` Int64 COMMENT 'total amount of withdrawals',
    `attestations_scheduled` Int64 COMMENT 'number of attestations scheduled',
    `attestations_observed` Int64 COMMENT 'number of attestations executed',
    `attestations_head_matched` Int64 COMMENT 'number of attestations matching head',
    `attestations_target_matched` Int64 COMMENT 'number of attestations matching target',
    `attestations_source_matched` Int64 COMMENT 'number of attestations matching source',
    `attestations_head_executed` Int64 COMMENT 'number of attestations executed on head',
    `attestations_target_executed` Int64 COMMENT 'number of attestations executed on target',
    `attestations_source_executed` Int64 COMMENT 'number of attestations executed on source',
    `attestations_head_reward` Int64 COMMENT 'total reward for attestations on head',
    `attestations_target_reward` Int64 COMMENT 'total reward for attestations on target',
    `attestations_source_reward` Int64 COMMENT 'total reward for attestations on source',
    `attestations_inactivity_reward` Int64 COMMENT 'total reward for attestations inactivity',
    `attestations_inclusion_reward` Int64 COMMENT 'total reward for attestations inclusion',
    `attestations_ideal_head_reward` Int64 COMMENT 'ideal reward for attestations on head' ,
    `attestations_ideal_target_reward` Int64 COMMENT 'ideal reward for attestations on target',
    `attestations_ideal_source_reward` Int64 COMMENT 'ideal reward for attestations on source',
    `attestations_ideal_inactivity_reward` Int64 COMMENT 'ideal reward for attestations inactivity',
    `attestations_ideal_inclusion_reward` Int64 COMMENT 'ideal reward for attestations inclusion',
    `attestations_localized_max_reward` Int64 COMMENT 'slot localized max reward for attestations',
    `attestations_hyperlocalized_max_reward` Int64 COMMENT 'committee localized max reward for attestations',
    `inclusion_delay_sum` Int64 COMMENT 'sum of inclusion delays',
    `optimal_inclusion_delay_sum` Int64 COMMENT 'sum of optimal inclusion delays',
    `blocks_status` Nested (
        `slot` Int64,
        `proposed` Bool
    ) COMMENT 'block status',
    `block_rewards` Nested (
        `slot` Int64,
        `attestations_reward` Int64,
        `sync_aggregate_reward` Int64,
        `slasher_reward` Int64
    ) COMMENT 'block rewards',
    `blocks_cl_missed_median_reward` Int64 COMMENT 'average reward for missed blocks',
    `blocks_slashing_count` Int64 COMMENT 'slashings in block count',
    `blocks_expected` Float64 COMMENT 'expected blocks',
    `sync_scheduled` Int64 COMMENT  'number of syncs scheduled', -- left because its not supposed to count skipped slots
    `sync_status` Nested (
        `slot` Int64,
        `executed` Bool
    ) COMMENT 'sync status',
    `sync_rewards` Nested (
        `slot` Int64,
        `reward` Int64
    ) COMMENT 'sync rewards',
    `sync_localized_max_reward` Int64 COMMENT 'slot localized max reward for syncs',
    `sync_committees_expected` Float64 COMMENT 'expected sync committees',
    `slashed` Bool COMMENT 'if the validator was slashed in the epoch ',
    `attestation_assignments` Nested (
        `slot` Int64,
        `committee` Int64,
        `index` Int64
    ) COMMENT 'attestation assignments',
    `sync_committee_assignments` Nested (
        `period` Int64,
        `index` Int64
    ) COMMENT 'sync committee assignments'
)
ENGINE = Null;
-- +goose StatementEnd
-- +goose StatementBegin
SET flatten_nested = 1; -- reset to default
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _unsafe_validator_dashboard_data_epoch
(   
    `_inserted_at` DateTime COMMENT 'insertion timestamp',
    `validator_index` UInt64 COMMENT 'validator index',
    `epoch` Int64 COMMENT 'epoch number',
    `epoch_timestamp` DateTime COMMENT 'timestamp of the first slot of the epoch',
    `balance_effective_start` Int64 COMMENT 'effective balance at the first slot of the current epoch',
    `balance_effective_end` Int64 COMMENT 'effective balance at the last slot of the current epoch',
    `balance_start` Int64 COMMENT 'balance at the last slot of the previous epoch',
    `balance_end` Int64 COMMENT 'balance at the last slot of the current epoch',
    `deposits_count` Int64 COMMENT 'number of deposits',
    `deposits_amount` Int64 COMMENT 'total amount of deposits',
    `withdrawals_count` Int64 COMMENT 'number of withdrawals',
    `withdrawals_amount` Int64 COMMENT 'total amount of withdrawals',
    `attestations_scheduled` Int64 COMMENT 'number of attestations scheduled',
    `attestations_observed` Int64 COMMENT 'number of attestations executed',
    `attestations_head_matched` Int64 COMMENT 'number of attestations matching head',
    `attestations_target_matched` Int64 COMMENT 'number of attestations matching target',
    `attestations_source_matched` Int64 COMMENT 'number of attestations matching source',
    `attestations_head_executed` Int64 COMMENT 'number of attestations executed on head',
    `attestations_target_executed` Int64 COMMENT 'number of attestations executed on target',
    `attestations_source_executed` Int64 COMMENT 'number of attestations executed on source',
    `attestations_reward` Int64 COMMENT 'total reward for attestations',
    `attestations_reward_rewards_only` Int64 COMMENT 'total reward for attestations, rewards only',
    `attestations_reward_penalties_only` Int64 COMMENT 'total reward for attestations, penalties only',
    `attestations_head_reward` Int64 COMMENT 'total reward for attestations on head',
    `attestations_target_reward` Int64 COMMENT 'total reward for attestations on target',
    `attestations_source_reward` Int64 COMMENT 'total reward for attestations on source',
    `attestations_inactivity_reward` Int64 COMMENT 'total reward for attestations inactivity',
    `attestations_inclusion_reward` Int64 COMMENT 'total reward for attestations inclusion',
    `attestations_ideal_reward` Int64 COMMENT 'ideal reward for attestations',
    `attestations_ideal_head_reward` Int64 COMMENT 'ideal reward for attestations on head',
    `attestations_ideal_target_reward` Int64 COMMENT 'ideal reward for attestations on target',
    `attestations_ideal_source_reward` Int64 COMMENT 'ideal reward for attestations on source',
    `attestations_ideal_inactivity_reward` Int64 COMMENT 'ideal reward for attestations inactivity',
    `attestations_ideal_inclusion_reward` Int64 COMMENT 'ideal reward for attestations inclusion',
    `attestations_localized_max_reward` Int64 COMMENT 'slot localized max reward for attestations',
    `attestations_hyperlocalized_max_reward` Int64 COMMENT 'committee localized max reward for attestations',
    `inclusion_delay_sum` Int64 COMMENT 'sum of inclusion delays',
    `optimal_inclusion_delay_sum` Int64 COMMENT 'sum of optimal inclusion delays',
    `blocks_scheduled` Int64 COMMENT 'number of blocks scheduled',
    `blocks_proposed` Int64 COMMENT 'number of blocks proposed',
    `blocks_cl_reward` Int64 COMMENT 'total consensus layer block reward',
    `blocks_cl_attestations_reward` Int64 COMMENT 'attestation consensus layer block reward',
    `blocks_cl_sync_aggregate_reward` Int64 COMMENT 'sync aggregate consensus layer block reward',
    `blocks_cl_slasher_reward` Int64 COMMENT 'slasher consensus layer block reward',
    `blocks_cl_missed_median_reward` Int64 COMMENT 'average reward for missed blocks',
    `blocks_slashing_count` Int64 COMMENT 'slashings in block count',
    `blocks_expected` Float64 COMMENT 'expected blocks',
    `sync_scheduled` Int64 COMMENT 'number of syncs scheduled',
    `sync_executed` Int64 COMMENT 'number of syncs executed',
    `sync_reward` Int64 COMMENT 'total sync rewards',
    `sync_reward_rewards_only` Int64 COMMENT 'total sync rewards, rewards only',
    `sync_reward_penalties_only` Int64 COMMENT 'total sync rewards, penalties only',
    `sync_localized_max_reward` Int64 COMMENT 'slot localized max reward for syncs',
    `sync_committees_expected` Float64 COMMENT 'expected sync committees',
    `slashed` Bool COMMENT 'if the validator was slashed in the epoch'
)
ENGINE = ReplacingMergeTree(_inserted_at)
ORDER BY (validator_index, epoch_timestamp, epoch)
PARTITION BY toStartOfHour(epoch_timestamp) --- this should be fine as the table will be cleaned up by the TTL
TTL _inserted_at + INTERVAL 1 DAY
SETTINGS index_granularity = 8192;
-- +goose StatementEnd
-- +goose StatementBegin
-- create safe non replacing version
CREATE TABLE IF NOT EXISTS _final_validator_dashboard_data_epoch as _unsafe_validator_dashboard_data_epoch
ENGINE = MergeTree
ORDER BY (validator_index, epoch_timestamp, epoch)
PARTITION BY toMonday(epoch_timestamp)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048;
-- +goose StatementEnd
-- +goose StatementBegin
-- make _inserted_at column EPHEMERAL
ALTER TABLE _final_validator_dashboard_data_epoch DROP COLUMN IF EXISTS _inserted_at;
-- +goose StatementEnd
-- +goose StatementBegin
-- materialized view to forward data to the unsafe table
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_unsafe_validator_dashboard_data_epoch TO _unsafe_validator_dashboard_data_epoch
AS SELECT
    now() as _inserted_at,
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
    attestations_head_reward + attestations_target_reward + attestations_source_reward + attestations_inactivity_reward + attestations_inclusion_reward as attestations_reward,
    greatest(attestations_head_reward, 0) + greatest(attestations_target_reward, 0) + greatest(attestations_source_reward, 0) + greatest(attestations_inactivity_reward, 0) + greatest(attestations_inclusion_reward, 0) as attestations_reward_rewards_only,
    least(attestations_head_reward, 0) + least(attestations_target_reward, 0) + least(attestations_source_reward, 0) + least(attestations_inactivity_reward, 0) + least(attestations_inclusion_reward, 0) as attestations_reward_penalties_only,
    attestations_head_reward,
    attestations_target_reward,
    attestations_source_reward,
    attestations_inactivity_reward,
    attestations_inclusion_reward,
    attestations_ideal_head_reward + attestations_ideal_target_reward + attestations_ideal_source_reward + attestations_ideal_inactivity_reward + attestations_ideal_inclusion_reward as attestations_ideal_reward,
    attestations_ideal_head_reward,
    attestations_ideal_target_reward,
    attestations_ideal_source_reward,
    attestations_ideal_inactivity_reward,
    attestations_ideal_inclusion_reward,
    attestations_localized_max_reward,
    attestations_hyperlocalized_max_reward,
    inclusion_delay_sum, 
    optimal_inclusion_delay_sum, 
    length(blocks_status.proposed) as blocks_scheduled,
    arrayCount(x -> x, blocks_status.proposed) as blocks_proposed,
    arraySum(block_rewards.attestations_reward) + arraySum(block_rewards.sync_aggregate_reward) + arraySum(block_rewards.slasher_reward) as blocks_cl_reward,
    arraySum(block_rewards.attestations_reward) as blocks_cl_attestations_reward, 
    arraySum(block_rewards.sync_aggregate_reward) as blocks_cl_sync_aggregate_reward,
    arraySum(block_rewards.slasher_reward) as blocks_cl_slasher_reward,
    blocks_cl_missed_median_reward,
    blocks_slashing_count, 
    blocks_expected, 
    sync_scheduled, 
    arrayCount(x -> x, sync_status.executed) as sync_executed,
    arraySum(sync_rewards.reward) as sync_reward,
    arraySum(x -> x > 0, sync_rewards.reward) as sync_reward_rewards_only,
    arraySum(x -> x < 0, sync_rewards.reward) as sync_reward_penalties_only,
    sync_localized_max_reward,
    sync_committees_expected,
    slashed
FROM _insert_sink_validator_dashboard_data_epoch;
-- +goose StatementEnd
-- +goose StatementBegin
-- attesations assignements
CREATE TABLE IF NOT EXISTS validator_attestation_assignments_slot (
    `validator_index` UInt64 COMMENT 'validator index' CODEC(DoubleDelta, ZSTD(8)),
    `epoch` Int64 COMMENT 'epoch number' CODEC(Delta, ZSTD(8)),
    `epoch_timestamp` DateTime COMMENT 'timestamp of the first slot of the epoch' CODEC(Delta, ZSTD(8)),
    `slot` Int64 CODEC(T64, ZSTD(8)),
    `committee` Int64 CODEC(T64, ZSTD(8)),
    `committee_index` Int64 CODEC(T64, ZSTD(8)),
)
ENGINE = ReplacingMergeTree
ORDER BY (validator_index, epoch_timestamp, epoch)
PARTITION BY toMonday(epoch_timestamp)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd
-- +goose StatementBegin
-- materialized view to forward data to the unsafe table
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_validator_attestation_assignments_slot TO validator_attestation_assignments_slot 
AS SELECT
    validator_index, 
    epoch, 
    epoch_timestamp, 
    x.slot as slot,
    x.committee as committee,
    x.index as committee_index
FROM _insert_sink_validator_dashboard_data_epoch ARRAY JOIN attestation_assignments as x;
-- +goose StatementEnd
-- +goose StatementBegin
-- proposal assignments
CREATE TABLE IF NOT EXISTS validator_proposal_assignments_slot (
    `validator_index` UInt64 COMMENT 'validator index' CODEC(T64, ZSTD(8)),
    `epoch` Int64 COMMENT 'epoch number' CODEC(Delta, ZSTD(8)),
    `epoch_timestamp` DateTime COMMENT 'timestamp of the first slot of the epoch' CODEC(Delta, ZSTD(8)),
    `slot` Int64 CODEC(Delta, ZSTD(8))
)
ENGINE = ReplacingMergeTree
PARTITION BY toStartOfQuarter(epoch_timestamp)
ORDER BY (validator_index, epoch_timestamp, epoch, slot)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd
-- +goose StatementBegin
-- materialized view to forward data to the unsafe table
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_validator_proposal_assignments_slot TO validator_proposal_assignments_slot
AS SELECT
    validator_index, 
    epoch, 
    epoch_timestamp, 
    slot
FROM _insert_sink_validator_dashboard_data_epoch ARRAY JOIN blocks_status.slot as slot;
-- +goose StatementEnd
-- +goose StatementBegin
-- sync committee assignments
CREATE TABLE IF NOT EXISTS validator_sync_committee_assignments_epoch (
    `validator_index` UInt64 COMMENT 'validator index' CODEC(T64, ZSTD(8)),
    `epoch` Int64 COMMENT 'epoch number' CODEC(Delta, ZSTD(8)),
    `epoch_timestamp` DateTime COMMENT 'timestamp of the first slot of the epoch' CODEC(Delta, ZSTD(8)),
    `period` Int64 CODEC(Delta, ZSTD(8)),
    `period_index` Int64 CODEC(Delta, ZSTD(8))
)
ENGINE = ReplacingMergeTree
PARTITION BY toStartOfQuarter(epoch_timestamp)
ORDER BY (validator_index, epoch_timestamp, epoch, period)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd
-- +goose StatementBegin
-- materialized view to forward data to the unsafe table
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_validator_sync_committee_assignments_epoch TO validator_sync_committee_assignments_epoch
AS SELECT
    validator_index,
    epoch,
    epoch_timestamp,
    x.period as period,
    x.index as period_index
FROM _insert_sink_validator_dashboard_data_epoch ARRAY JOIN sync_committee_assignments as x;
-- +goose StatementEnd
-- +goose StatementBegin
-- proposal reward table
CREATE TABLE IF NOT EXISTS validator_proposal_rewards_slot (
    `validator_index` UInt64 COMMENT 'validator index' CODEC(T64, ZSTD(8)),
    `epoch` Int64 COMMENT 'epoch number' CODEC(Delta, ZSTD(8)),
    `epoch_timestamp` DateTime COMMENT 'timestamp of the first slot of the epoch' CODEC(Delta, ZSTD(8)),
    `slot` Int64 CODEC(Delta, ZSTD(8)),
    `attestations_reward` Int64 COMMENT 'reward for including attestations in the proposal' CODEC(T64, ZSTD(8)),
    `sync_aggregate_reward` Int64 COMMENT 'reward for including sync aggregate in the proposal' CODEC(T64, ZSTD(8)),
    `slasher_reward` Int64 COMMENT 'reward for including slasher in the proposal' CODEC(T64, ZSTD(8)),
)
ENGINE = ReplacingMergeTree
ORDER BY (validator_index, epoch_timestamp, epoch, slot)
PARTITION BY toMonday(epoch_timestamp)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd
-- +goose StatementBegin
-- materialized view to forward data to the unsafe table
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_validator_proposal_rewards_slot TO validator_proposal_rewards_slot
AS SELECT
    validator_index, 
    epoch, 
    epoch_timestamp, 
    x.slot as slot,
    x.attestations_reward as attestations_reward,
    x.sync_aggregate_reward as sync_aggregate_reward,
    x.slasher_reward as slasher_reward
FROM _insert_sink_validator_dashboard_data_epoch ARRAY JOIN block_rewards as x;
-- +goose StatementEnd
-- +goose StatementBegin
-- sync_executed_map array join to generate table of validator_index, epoch_timestamp, slot
CREATE TABLE IF NOT EXISTS validator_sync_committee_votes_epoch (
    `validator_index` UInt64 COMMENT 'validator index' CODEC(DoubleDelta, ZSTD(8)),
    `epoch` Int64 COMMENT 'epoch number' CODEC(Delta, ZSTD(8)),
    `epoch_timestamp` DateTime COMMENT 'timestamp of the first slot of the epoch' CODEC(Delta, ZSTD(8)),
    `slot` Int64 COMMENT 'slot number' CODEC(T64, ZSTD(8)),
    `executed` Bool COMMENT 'if the sync was executed'
)
ENGINE = ReplacingMergeTree
ORDER BY (validator_index, epoch_timestamp, epoch, slot)
PARTITION BY toMonday(epoch_timestamp)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd
-- +goose StatementBegin
-- materialized view to forward data to the unsafe table
CREATE MATERIALIZED VIEW IF NOT EXISTS _mv_validator_sync_committee_votes_epoch TO validator_sync_committee_votes_epoch
AS SELECT
    validator_index, 
    epoch, 
    epoch_timestamp, 
    x.slot as slot,
    x.executed as executed
FROM _insert_sink_validator_dashboard_data_epoch ARRAY JOIN sync_status as x;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _final_validator_dashboard_data_hourly (
    `validator_index` UInt64 COMMENT 'validator index' CODEC(DoubleDelta, ZSTD(8)),
    `t` DateTime COMMENT 'timestamp of the aggregated data' CODEC(Delta, ZSTD(8)),
    `epoch_map` AggregateFunction(groupArraySortedIf(2048), Int64, Bool) COMMENT 'consistency check data - only available for validator 0',
    `epoch_start` SimpleAggregateFunction(min, Int64) COMMENT 'first epoch included in the aggregation',
    `epoch_end` SimpleAggregateFunction(max, Int64) COMMENT 'last epoch included in the aggregation',
    `balance_start` AggregateFunction(argMin, Int64, Int64) COMMENT 'balance at the first slot of the first epoch included in the aggregation',
    `balance_end` AggregateFunction(argMax, Int64, Int64) COMMENT 'balance at the last slot of the last epoch included in the aggregation',
    `balance_min` SimpleAggregateFunction(min, Int64) COMMENT 'minimum balance in the aggregation',
    `balance_max` SimpleAggregateFunction(max, Int64) COMMENT 'maximum balance in the aggregation',
    `deposits_count` SimpleAggregateFunction(sum, Int64) COMMENT 'number of deposits',
    `deposits_amount` SimpleAggregateFunction(sum, Int64) COMMENT 'total amount of deposits',
    `withdrawals_count` SimpleAggregateFunction(sum, Int64) COMMENT 'number of withdrawals',
    `withdrawals_amount` SimpleAggregateFunction(sum, Int64) COMMENT 'total amount of withdrawals',
    `attestations_scheduled` SimpleAggregateFunction(sum, Int64) COMMENT 'number of attestations scheduled',
    `attestations_observed` SimpleAggregateFunction(sum, Int64) COMMENT 'number of attestations executed',
    `attestations_head_matched` SimpleAggregateFunction(sum, Int64) COMMENT 'number of attestations matching head',
    `attestations_target_matched` SimpleAggregateFunction(sum, Int64) COMMENT 'number of attestations matching target',
    `attestations_source_matched` SimpleAggregateFunction(sum, Int64) COMMENT 'number of attestations matching source',
    `attestations_head_executed` SimpleAggregateFunction(sum, Int64) COMMENT 'number of attestations executed on head',
    `attestations_target_executed` SimpleAggregateFunction(sum, Int64) COMMENT 'number of attestations executed on target',
    `attestations_source_executed` SimpleAggregateFunction(sum, Int64) COMMENT 'number of attestations executed on source',
    `attestations_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'total reward for attestations' CODEC(T64, ZSTD(8)),
    `attestations_reward_rewards_only` SimpleAggregateFunction(sum, Int64) COMMENT 'total reward for attestations, rewards only' CODEC(T64, ZSTD(8)),
    `attestations_reward_penalties_only` SimpleAggregateFunction(sum, Int64) COMMENT 'total reward for attestations, penalties only' CODEC(T64, ZSTD(8)),
    `attestations_head_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'total reward for attestations on head' CODEC(T64, ZSTD(8)),
    `attestations_target_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'total reward for attestations on target' CODEC(T64, ZSTD(8)),
    `attestations_source_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'total reward for attestations on source' CODEC(T64, ZSTD(8)),
    `attestations_inactivity_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'total reward for attestations inactivity' CODEC(T64, ZSTD(8)),
    `attestations_inclusion_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'total reward for attestations inclusion' CODEC(T64, ZSTD(8)),
    `attestations_ideal_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'ideal reward for attestations' CODEC(Delta, ZSTD(8)),
    `attestations_ideal_head_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'ideal reward for attestations on head' CODEC(Delta, ZSTD(8)),
    `attestations_ideal_target_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'ideal reward for attestations on target' CODEC(Delta, ZSTD(8)),
    `attestations_ideal_source_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'ideal reward for attestations on source' CODEC(Delta, ZSTD(8)),
    `attestations_ideal_inactivity_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'ideal reward for attestations inactivity' CODEC(Delta, ZSTD(8)),
    `attestations_ideal_inclusion_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'ideal reward for attestations inclusion' CODEC(Delta, ZSTD(8)),
    `attestations_localized_max_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'slot localized max reward for attestations' CODEC(Delta, ZSTD(8)),
    `attestations_hyperlocalized_max_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'committee localized max reward for attestations' CODEC(Delta, ZSTD(8)),
    `inclusion_delay_sum` SimpleAggregateFunction(sum, Int64) COMMENT 'sum of inclusion delays' CODEC(T64, ZSTD(8)),
    `optimal_inclusion_delay_sum` SimpleAggregateFunction(sum, Int64) COMMENT 'sum of optimal inclusion delays' CODEC(T64, ZSTD(8)),
    `blocks_scheduled` SimpleAggregateFunction(sum, Int64) COMMENT 'number of blocks scheduled',
    `blocks_proposed` SimpleAggregateFunction(sum, Int64) COMMENT 'number of blocks proposed',
    `blocks_cl_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'total consensus layer block reward',
    `blocks_cl_attestations_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'attestation consensus layer block reward',
    `blocks_cl_sync_aggregate_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'sync aggregate consensus layer block reward',
    `blocks_cl_slasher_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'slasher consensus layer block reward',
    `blocks_cl_missed_median_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'average reward for missed blocks' CODEC(T64, ZSTD(8)),
    `blocks_slashing_count` SimpleAggregateFunction(sum, Int64) COMMENT 'slashings in block count',
    `blocks_expected` SimpleAggregateFunction(sum, Float64) COMMENT 'expected blocks' CODEC(FPC, ZSTD(8)),
    `sync_scheduled` SimpleAggregateFunction(sum, Int64) COMMENT 'number of syncs scheduled',
    `sync_executed` SimpleAggregateFunction(sum, Int64) COMMENT 'number of syncs executed',
    `sync_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'total sync rewards',
    `sync_reward_rewards_only` SimpleAggregateFunction(sum, Int64) COMMENT 'total sync rewards, rewards only' CODEC(T64, ZSTD(8)),
    `sync_reward_penalties_only` SimpleAggregateFunction(sum, Int64) COMMENT 'total sync rewards, penalties only' CODEC(T64, ZSTD(8)),
    `sync_localized_max_reward` SimpleAggregateFunction(sum, Int64) COMMENT 'slot localized max reward for syncs' CODEC(Delta, ZSTD(8)),
    `sync_committees_expected` SimpleAggregateFunction(sum, Float64) COMMENT 'expected sync committees' CODEC(FPC, ZSTD(8)),
    `slashed` SimpleAggregateFunction(max, Bool) COMMENT 'if the validator was slashed in the epoch',
    `last_executed_duty_epoch` SimpleAggregateFunction(max, Nullable(Int64)),
    `last_scheduled_sync_epoch` SimpleAggregateFunction(max, Nullable(Int64)),
    `last_scheduled_block_epoch` SimpleAggregateFunction(max, Nullable(Int64))
)
ENGINE = AggregatingMergeTree
PARTITION BY toYYYYMM(t)
ORDER BY (validator_index, t)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048;
-- +goose StatementEnd
-- +goose StatementBegin
-- daily
CREATE TABLE IF NOT EXISTS _final_validator_dashboard_data_daily as _final_validator_dashboard_data_hourly
ENGINE = AggregatingMergeTree
PARTITION BY toStartOfQuarter(t)
ORDER BY (validator_index, t)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048;
-- +goose StatementEnd
-- +goose StatementBegin
-- weekly
CREATE TABLE IF NOT EXISTS _final_validator_dashboard_data_weekly as _final_validator_dashboard_data_hourly
ENGINE = AggregatingMergeTree
PARTITION BY toStartOfYear(t)
ORDER BY (validator_index, t)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048;
-- +goose StatementEnd
-- +goose StatementBegin
-- monthly
CREATE TABLE IF NOT EXISTS _final_validator_dashboard_data_monthly as _final_validator_dashboard_data_hourly
ENGINE = AggregatingMergeTree
ORDER BY (validator_index, t)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS _final_validator_dashboard_data_monthly;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _final_validator_dashboard_data_weekly;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _final_validator_dashboard_data_daily;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _final_validator_dashboard_data_hourly;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS validator_proposal_assignments_slot;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS validator_attestation_assignments_slot;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _mv_validator_proposal_assignments_slot;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _mv_validator_attestation_assignments_slot;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _mv_unsafe_validator_dashboard_data_epoch;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _unsafe_validator_dashboard_data_epoch;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _insert_sink_validator_dashboard_data_epoch;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _final_validator_dashboard_data_epoch;
-- +goose StatementEnd
