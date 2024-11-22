-- +goose Up


-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS validator_dashboard_data_rolling_1h
(
    `version` DateTime,
    `validator_index` UInt64,
    `epoch_start` Int64,
    `epoch_end` Int64,
    `balance_start` Int64,
    `balance_end` Int64,
    `balance_min` Int64,
    `balance_max` Int64,
    `deposits_count` Int64,
    `deposits_amount` Int64,
    `withdrawals_count` Int64,
    `withdrawals_amount` Int64,
    `attestations_scheduled` Int64,
    `attestations_executed` Int64,
    `attestation_head_executed` Int64,
    `attestation_source_executed` Int64,
    `attestation_target_executed` Int64,
    `attestations_reward` Int64,
    `attestations_reward_rewards_only` Int64,
    `attestations_reward_penalties_only` Int64,
    `attestations_source_reward` Int64,
    `attestations_target_reward` Int64,
    `attestations_head_reward` Int64,
    `attestations_inactivity_reward` Int64,
    `attestations_inclusion_reward` Int64,
    `attestations_ideal_reward` Int64,
    `attestations_ideal_source_reward` Int64,
    `attestations_ideal_target_reward` Int64,
    `attestations_ideal_head_reward` Int64,
    `attestations_ideal_inactivity_reward` Int64,
    `attestations_ideal_inclusion_reward` Int64,
    `inclusion_delay_sum` Int64,
    `optimal_inclusion_delay_sum` Int64,
    `blocks_scheduled` Int64,
    `blocks_proposed` Int64,
    `blocks_cl_reward` Int64,
    `blocks_cl_attestations_reward` Int64,
    `blocks_cl_sync_aggregate_reward` Int64,
    `blocks_cl_slasher_reward` Int64,
    `blocks_slashing_count` Int64,
    `blocks_expected` Float64,
    `sync_scheduled` Int64,
    `sync_executed` Int64,
    `sync_rewards` Int64,
    `sync_committees_expected` Float64,
    `slashed` Bool,
    `last_executed_duty_epoch` Nullable(Int64),
    `last_scheduled_sync_epoch` Nullable(Int64),
    `last_scheduled_block_epoch` Nullable(Int64)
)
ENGINE = ReplacingMergeTree()
ORDER BY validator_index
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS validator_dashboard_data_rolling_24h
(
    `version` DateTime,
    `validator_index` UInt64,
    `epoch_start` Int64,
    `epoch_end` Int64,
    `balance_start` Int64,
    `balance_end` Int64,
    `balance_min` Int64,
    `balance_max` Int64,
    `deposits_count` Int64,
    `deposits_amount` Int64,
    `withdrawals_count` Int64,
    `withdrawals_amount` Int64,
    `attestations_scheduled` Int64,
    `attestations_executed` Int64,
    `attestation_head_executed` Int64,
    `attestation_source_executed` Int64,
    `attestation_target_executed` Int64,
    `attestations_reward` Int64,
    `attestations_reward_rewards_only` Int64,
    `attestations_reward_penalties_only` Int64,
    `attestations_source_reward` Int64,
    `attestations_target_reward` Int64,
    `attestations_head_reward` Int64,
    `attestations_inactivity_reward` Int64,
    `attestations_inclusion_reward` Int64,
    `attestations_ideal_reward` Int64,
    `attestations_ideal_source_reward` Int64,
    `attestations_ideal_target_reward` Int64,
    `attestations_ideal_head_reward` Int64,
    `attestations_ideal_inactivity_reward` Int64,
    `attestations_ideal_inclusion_reward` Int64,
    `inclusion_delay_sum` Int64,
    `optimal_inclusion_delay_sum` Int64,
    `blocks_scheduled` Int64,
    `blocks_proposed` Int64,
    `blocks_cl_reward` Int64,
    `blocks_cl_attestations_reward` Int64,
    `blocks_cl_sync_aggregate_reward` Int64,
    `blocks_cl_slasher_reward` Int64,
    `blocks_slashing_count` Int64,
    `blocks_expected` Float64,
    `sync_scheduled` Int64,
    `sync_executed` Int64,
    `sync_rewards` Int64,
    `sync_committees_expected` Float64,
    `slashed` Bool,
    `last_executed_duty_epoch` Nullable(Int64),
    `last_scheduled_sync_epoch` Nullable(Int64),
    `last_scheduled_block_epoch` Nullable(Int64)
)
ENGINE = ReplacingMergeTree()
ORDER BY validator_index
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS validator_dashboard_data_rolling_7d
(
    `version` DateTime,
    `validator_index` UInt64,
    `epoch_start` Int64,
    `epoch_end` Int64,
    `balance_start` Int64,
    `balance_end` Int64,
    `balance_min` Int64,
    `balance_max` Int64,
    `deposits_count` Int64,
    `deposits_amount` Int64,
    `withdrawals_count` Int64,
    `withdrawals_amount` Int64,
    `attestations_scheduled` Int64,
    `attestations_executed` Int64,
    `attestation_head_executed` Int64,
    `attestation_source_executed` Int64,
    `attestation_target_executed` Int64,
    `attestations_reward` Int64,
    `attestations_reward_rewards_only` Int64,
    `attestations_reward_penalties_only` Int64,
    `attestations_source_reward` Int64,
    `attestations_target_reward` Int64,
    `attestations_head_reward` Int64,
    `attestations_inactivity_reward` Int64,
    `attestations_inclusion_reward` Int64,
    `attestations_ideal_reward` Int64,
    `attestations_ideal_source_reward` Int64,
    `attestations_ideal_target_reward` Int64,
    `attestations_ideal_head_reward` Int64,
    `attestations_ideal_inactivity_reward` Int64,
    `attestations_ideal_inclusion_reward` Int64,
    `inclusion_delay_sum` Int64,
    `optimal_inclusion_delay_sum` Int64,
    `blocks_scheduled` Int64,
    `blocks_proposed` Int64,
    `blocks_cl_reward` Int64,
    `blocks_cl_attestations_reward` Int64,
    `blocks_cl_sync_aggregate_reward` Int64,
    `blocks_cl_slasher_reward` Int64,
    `blocks_slashing_count` Int64,
    `blocks_expected` Float64,
    `sync_scheduled` Int64,
    `sync_executed` Int64,
    `sync_rewards` Int64,
    `sync_committees_expected` Float64,
    `slashed` Bool,
    `last_executed_duty_epoch` Nullable(Int64),
    `last_scheduled_sync_epoch` Nullable(Int64),
    `last_scheduled_block_epoch` Nullable(Int64)
)
ENGINE = ReplacingMergeTree()
ORDER BY validator_index
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS validator_dashboard_data_rolling_30d
(
    `version` DateTime,
    `validator_index` UInt64,
    `epoch_start` Int64,
    `epoch_end` Int64,
    `balance_start` Int64,
    `balance_end` Int64,
    `balance_min` Int64,
    `balance_max` Int64,
    `deposits_count` Int64,
    `deposits_amount` Int64,
    `withdrawals_count` Int64,
    `withdrawals_amount` Int64,
    `attestations_scheduled` Int64,
    `attestations_executed` Int64,
    `attestation_head_executed` Int64,
    `attestation_source_executed` Int64,
    `attestation_target_executed` Int64,
    `attestations_reward` Int64,
    `attestations_reward_rewards_only` Int64,
    `attestations_reward_penalties_only` Int64,
    `attestations_source_reward` Int64,
    `attestations_target_reward` Int64,
    `attestations_head_reward` Int64,
    `attestations_inactivity_reward` Int64,
    `attestations_inclusion_reward` Int64,
    `attestations_ideal_reward` Int64,
    `attestations_ideal_source_reward` Int64,
    `attestations_ideal_target_reward` Int64,
    `attestations_ideal_head_reward` Int64,
    `attestations_ideal_inactivity_reward` Int64,
    `attestations_ideal_inclusion_reward` Int64,
    `inclusion_delay_sum` Int64,
    `optimal_inclusion_delay_sum` Int64,
    `blocks_scheduled` Int64,
    `blocks_proposed` Int64,
    `blocks_cl_reward` Int64,
    `blocks_cl_attestations_reward` Int64,
    `blocks_cl_sync_aggregate_reward` Int64,
    `blocks_cl_slasher_reward` Int64,
    `blocks_slashing_count` Int64,
    `blocks_expected` Float64,
    `sync_scheduled` Int64,
    `sync_executed` Int64,
    `sync_rewards` Int64,
    `sync_committees_expected` Float64,
    `slashed` Bool,
    `last_executed_duty_epoch` Nullable(Int64),
    `last_scheduled_sync_epoch` Nullable(Int64),
    `last_scheduled_block_epoch` Nullable(Int64)
)
ENGINE = ReplacingMergeTree()
ORDER BY validator_index
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS validator_dashboard_data_rolling_90d
(
    `version` DateTime,
    `validator_index` UInt64,
    `epoch_start` Int64,
    `epoch_end` Int64,
    `balance_start` Int64,
    `balance_end` Int64,
    `balance_min` Int64,
    `balance_max` Int64,
    `deposits_count` Int64,
    `deposits_amount` Int64,
    `withdrawals_count` Int64,
    `withdrawals_amount` Int64,
    `attestations_scheduled` Int64,
    `attestations_executed` Int64,
    `attestation_head_executed` Int64,
    `attestation_source_executed` Int64,
    `attestation_target_executed` Int64,
    `attestations_reward` Int64,
    `attestations_reward_rewards_only` Int64,
    `attestations_reward_penalties_only` Int64,
    `attestations_source_reward` Int64,
    `attestations_target_reward` Int64,
    `attestations_head_reward` Int64,
    `attestations_inactivity_reward` Int64,
    `attestations_inclusion_reward` Int64,
    `attestations_ideal_reward` Int64,
    `attestations_ideal_source_reward` Int64,
    `attestations_ideal_target_reward` Int64,
    `attestations_ideal_head_reward` Int64,
    `attestations_ideal_inactivity_reward` Int64,
    `attestations_ideal_inclusion_reward` Int64,
    `inclusion_delay_sum` Int64,
    `optimal_inclusion_delay_sum` Int64,
    `blocks_scheduled` Int64,
    `blocks_proposed` Int64,
    `blocks_cl_reward` Int64,
    `blocks_cl_attestations_reward` Int64,
    `blocks_cl_sync_aggregate_reward` Int64,
    `blocks_cl_slasher_reward` Int64,
    `blocks_slashing_count` Int64,
    `blocks_expected` Float64,
    `sync_scheduled` Int64,
    `sync_executed` Int64,
    `sync_rewards` Int64,
    `sync_committees_expected` Float64,
    `slashed` Bool,
    `last_executed_duty_epoch` Nullable(Int64),
    `last_scheduled_sync_epoch` Nullable(Int64),
    `last_scheduled_block_epoch` Nullable(Int64)
)
ENGINE = ReplacingMergeTree()
ORDER BY validator_index
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS validator_dashboard_data_rolling_total
(
    `version` DateTime,
    `validator_index` UInt64,
    `epoch_start` Int64,
    `epoch_end` Int64,
    `balance_start` Int64,
    `balance_end` Int64,
    `balance_min` Int64,
    `balance_max` Int64,
    `deposits_count` Int64,
    `deposits_amount` Int64,
    `withdrawals_count` Int64,
    `withdrawals_amount` Int64,
    `attestations_scheduled` Int64,
    `attestations_executed` Int64,
    `attestation_head_executed` Int64,
    `attestation_source_executed` Int64,
    `attestation_target_executed` Int64,
    `attestations_reward` Int64,
    `attestations_reward_rewards_only` Int64,
    `attestations_reward_penalties_only` Int64,
    `attestations_source_reward` Int64,
    `attestations_target_reward` Int64,
    `attestations_head_reward` Int64,
    `attestations_inactivity_reward` Int64,
    `attestations_inclusion_reward` Int64,
    `attestations_ideal_reward` Int64,
    `attestations_ideal_source_reward` Int64,
    `attestations_ideal_target_reward` Int64,
    `attestations_ideal_head_reward` Int64,
    `attestations_ideal_inactivity_reward` Int64,
    `attestations_ideal_inclusion_reward` Int64,
    `inclusion_delay_sum` Int64,
    `optimal_inclusion_delay_sum` Int64,
    `blocks_scheduled` Int64,
    `blocks_proposed` Int64,
    `blocks_cl_reward` Int64,
    `blocks_cl_attestations_reward` Int64,
    `blocks_cl_sync_aggregate_reward` Int64,
    `blocks_cl_slasher_reward` Int64,
    `blocks_slashing_count` Int64,
    `blocks_expected` Float64,
    `sync_scheduled` Int64,
    `sync_executed` Int64,
    `sync_rewards` Int64,
    `sync_committees_expected` Float64,
    `slashed` Bool,
    `last_executed_duty_epoch` Nullable(Int64),
    `last_scheduled_sync_epoch` Nullable(Int64),
    `last_scheduled_block_epoch` Nullable(Int64)
)
ENGINE = ReplacingMergeTree()
ORDER BY validator_index
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE validator_dashboard_data_epoch
(
    `validator_index` UInt64 CODEC(DoubleDelta, ZSTD(8)),
    `epoch` Int64 CODEC(Delta(4), ZSTD(8)),
    `epoch_timestamp` DateTime CODEC(Delta(4), ZSTD(8)),
    `balance_effective_start` Int64 DEFAULT -1 CODEC(T64, ZSTD(8)),
    `balance_effective_end` Int64 DEFAULT -1 CODEC(T64, ZSTD(8)),
    `balance_start` Int64 CODEC(T64, ZSTD(8)),
    `balance_end` Int64 CODEC(T64, ZSTD(8)),
    `deposits_count` Int64,
    `deposits_amount` Int64,
    `withdrawals_count` Int64,
    `withdrawals_amount` Int64,
    `attestations_scheduled` Int64,
    `attestations_executed` Int64,
    `attestation_head_executed` Int64,
    `attestation_source_executed` Int64,
    `attestation_target_executed` Int64,
    `attestations_reward` Int64 CODEC(T64, ZSTD(8)),
    `attestations_source_reward` Int64 CODEC(T64, ZSTD(8)),
    `attestations_target_reward` Int64 CODEC(T64, ZSTD(8)),
    `attestations_head_reward` Int64 CODEC(T64, ZSTD(8)),
    `attestations_inactivity_reward` Int64 CODEC(T64, ZSTD(8)),
    `attestations_inclusion_reward` Int64 CODEC(T64, ZSTD(8)),
    `attestations_ideal_reward` Int64 CODEC(Delta(8), ZSTD(8)),
    `attestations_ideal_source_reward` Int64 CODEC(Delta(8), ZSTD(8)),
    `attestations_ideal_target_reward` Int64 CODEC(Delta(8), ZSTD(8)),
    `attestations_ideal_head_reward` Int64 CODEC(Delta(8), ZSTD(8)),
    `attestations_ideal_inactivity_reward` Int64 CODEC(Delta(8), ZSTD(8)),
    `attestations_ideal_inclusion_reward` Int64 CODEC(Delta(8), ZSTD(8)),
    `inclusion_delay_sum` Int64 CODEC(T64, ZSTD(8)),
    `optimal_inclusion_delay_sum` Int64 CODEC(T64, ZSTD(8)),
    `blocks_scheduled` Int64,
    `blocks_proposed` Int64,
    `blocks_cl_reward` Int64,
    `blocks_cl_attestations_reward` Int64,
    `blocks_cl_sync_aggregate_reward` Int64,
    `blocks_cl_slasher_reward` Int64,
    `blocks_slashing_count` Int64,
    `blocks_expected` Float64 CODEC(FPC(12), ZSTD(8)),
    `sync_scheduled` Int64,
    `sync_executed` Int64,
    `sync_rewards` Int64,
    `sync_committees_expected` Float64 CODEC(FPC(12), ZSTD(8)),
    `slashed` Bool
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(epoch_timestamp)
ORDER BY (epoch_timestamp,
 validator_index)
SETTINGS index_granularity = 8192,
 non_replicated_deduplication_window = 100;
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS validator_dashboard_data_rolling_1h;
DROP TABLE IF EXISTS validator_dashboard_data_rolling_24h;
DROP TABLE IF EXISTS validator_dashboard_data_rolling_7d;
DROP TABLE IF EXISTS validator_dashboard_data_rolling_30d;
DROP TABLE IF EXISTS validator_dashboard_data_rolling_90d;
DROP TABLE IF EXISTS validator_dashboard_data_rolling_total;
DROP TABLE IF EXISTS validator_dashboard_data_epoch;
-- +goose StatementBegin

-- +goose StatementEnd
