-- +goose Up
-- +goose StatementBegin


CREATE TABLE IF NOT EXISTS validator_dashboard_data_epoch (
    validator_index int NOT NULL,
    epoch int NOT NULL,
    attestations_source_reward BIGINT ,
    attestations_target_reward BIGINT,
    attestations_head_reward BIGINT,
    attestations_inactivity_reward BIGINT,
    attestations_inclusion_reward BIGINT,
    attestations_reward BIGINT,
    attestations_ideal_source_reward BIGINT,
    attestations_ideal_target_reward BIGINT,
    attestations_ideal_head_reward BIGINT,
    attestations_ideal_inactivity_reward BIGINT,
    attestations_ideal_inclusion_reward BIGINT,
    attestations_ideal_reward BIGINT,
    blocks_scheduled smallint,
    blocks_proposed smallint,
    blocks_cl_reward BIGINT, -- gwei
    blocks_el_reward NUMERIC, -- wei
    sync_scheduled smallint,
    sync_executed smallint,
    sync_rewards BIGINT,
    slashed BOOLEAN,
    balance_start BIGINT,
    balance_end BIGINT,
    deposits_count smallint,
    deposits_amount BIGINT,
    withdrawals_count smallint,
    withdrawals_amount BIGINT,
    inclusion_delay_sum smallint,
    sync_chance decimal(8, 7), -- slots_per_epochs / sum(active_validators)
    block_chance decimal(8, 7), -- size_of_sync / number_of_active_validators * slots_per_sync_period
    attestations_scheduled smallint,
    attestations_executed smallint,
    attestation_head_executed smallint,
    attestation_source_executed smallint,
    attestation_target_executed smallint,
    primary key (validator_index, epoch)
) PARTITION BY range (epoch);


CREATE TABLE IF NOT EXISTS validator_dashboard_data_rolling_daily (
    validator_index int NOT NULL,
    attestations_source_reward BIGINT,
    attestations_target_reward BIGINT,
    attestations_head_reward BIGINT,
    attestations_inactivity_reward BIGINT,
    attestations_inclusion_reward BIGINT,
    attestations_reward BIGINT,
    attestations_ideal_source_reward BIGINT,
    attestations_ideal_target_reward BIGINT,
    attestations_ideal_head_reward BIGINT,
    attestations_ideal_inactivity_reward BIGINT,
    attestations_ideal_inclusion_reward BIGINT,
    attestations_ideal_reward BIGINT,
    blocks_scheduled int,
    blocks_proposed int,
    blocks_cl_reward BIGINT, -- gwei
    blocks_el_reward NUMERIC, -- wei
    sync_scheduled int,
    sync_executed int,
    sync_rewards BIGINT,
    slashed BOOLEAN,
    balance_start BIGINT,
    balance_end BIGINT,
    deposits_count int,
    deposits_amount BIGINT,
    withdrawals_count int,
    withdrawals_amount BIGINT,
    inclusion_delay_sum int,
    sync_chance decimal(8, 7), -- slots_per_epochs / sum(active_validators)
    block_chance decimal(8, 7), -- size_of_sync / number_of_active_validators * slots_per_sync_period
    attestations_scheduled int,
    attestations_executed int,
    attestation_head_executed int,
    attestation_source_executed int,
    attestation_target_executed int,
    primary key (validator_index)
);


CREATE TABLE IF NOT EXISTS validator_dashboard_data_hourly (
    validator_index int NOT NULL,
    epoch_start int NOT NULL,
    epoch_end int NOT NULL,
    attestations_source_reward BIGINT,
    attestations_target_reward BIGINT,
    attestations_head_reward BIGINT,
    attestations_inactivity_reward BIGINT,
    attestations_inclusion_reward BIGINT,
    attestations_reward BIGINT,
    attestations_ideal_source_reward BIGINT,
    attestations_ideal_target_reward BIGINT,
    attestations_ideal_head_reward BIGINT,
    attestations_ideal_inactivity_reward BIGINT,
    attestations_ideal_inclusion_reward BIGINT,
    attestations_ideal_reward BIGINT,
    blocks_scheduled int,
    blocks_proposed int,
    blocks_cl_reward BIGINT, -- gwei
    blocks_el_reward NUMERIC, -- wei
    sync_scheduled int,
    sync_executed int,
    sync_rewards BIGINT,
    slashed BOOLEAN,
    balance_start BIGINT,
    balance_end BIGINT,
    deposits_count int,
    deposits_amount BIGINT,
    withdrawals_count int,
    withdrawals_amount BIGINT,
    inclusion_delay_sum int,
    sync_chance decimal(8, 7), -- slots_per_epochs / sum(active_validators)
    block_chance decimal(8, 7), -- size_of_sync / number_of_active_validators * slots_per_sync_period
    attestations_scheduled smallint,
    attestations_executed smallint,
    attestation_head_executed smallint,
    attestation_source_executed smallint,
    attestation_target_executed smallint,
    primary key (epoch_start, epoch_end, validator_index)
) PARTITION BY range(epoch_start);


CREATE TABLE IF NOT EXISTS validator_dashboard_data_daily (
    validator_index int NOT NULL,
    day date NOT NULL,
    attestations_source_reward BIGINT,
    attestations_target_reward BIGINT,
    attestations_head_reward BIGINT,
    attestations_inactivity_reward BIGINT,
    attestations_inclusion_reward BIGINT,
    attestations_reward BIGINT,
    attestations_ideal_source_reward BIGINT,
    attestations_ideal_target_reward BIGINT,
    attestations_ideal_head_reward BIGINT,
    attestations_ideal_inactivity_reward BIGINT,
    attestations_ideal_inclusion_reward BIGINT,
    attestations_ideal_reward BIGINT,
    blocks_scheduled int,
    blocks_proposed int,
    blocks_cl_reward BIGINT, -- gwei
    blocks_el_reward NUMERIC, -- wei
    sync_scheduled int,
    sync_executed int,
    sync_rewards BIGINT,
    slashed BOOLEAN,
    balance_start BIGINT,
    balance_end BIGINT,
    deposits_count int,
    deposits_amount BIGINT,
    withdrawals_count int,
    withdrawals_amount BIGINT,
    inclusion_delay_sum int,
    sync_chance decimal(8, 7), -- slots_per_epochs / sum(active_validators)
    block_chance decimal(8, 7), -- size_of_sync / number_of_active_validators * slots_per_sync_period
    primary key (day, validator_index)
);

CREATE TABLE IF NOT EXISTS validator_dashboard_data_rolling_weekly (
    validator_index int NOT NULL,
    attestations_source_reward BIGINT,
    attestations_target_reward BIGINT,
    attestations_head_reward BIGINT,
    attestations_inactivity_reward BIGINT,
    attestations_inclusion_reward BIGINT,
    attestations_reward BIGINT,
    attestations_ideal_source_reward BIGINT,
    attestations_ideal_target_reward BIGINT,
    attestations_ideal_head_reward BIGINT,
    attestations_ideal_inactivity_reward BIGINT,
    attestations_ideal_inclusion_reward BIGINT,
    attestations_ideal_reward BIGINT,
    blocks_scheduled int,
    blocks_proposed int,
    blocks_cl_reward BIGINT, -- gwei
    blocks_el_reward NUMERIC, -- wei
    sync_scheduled int,
    sync_executed int,
    sync_rewards BIGINT,
    slashed BOOLEAN,
    balance_start BIGINT,
    balance_end BIGINT,
    deposits_count int,
    deposits_amount BIGINT,
    withdrawals_count int,
    withdrawals_amount BIGINT,
    inclusion_delay_sum int,
    sync_chance decimal(8, 7), -- slots_per_epochs / sum(active_validators)
    block_chance decimal(8, 7), -- size_of_sync / number_of_active_validators * slots_per_sync_period
    primary key (validator_index)
);

CREATE TABLE IF NOT EXISTS validator_dashboard_data_rolling_monthly (
    validator_index int NOT NULL,
    attestations_source_reward BIGINT,
    attestations_target_reward BIGINT,
    attestations_head_reward BIGINT,
    attestations_inactivity_reward BIGINT,
    attestations_inclusion_reward BIGINT,
    attestations_reward BIGINT,
    attestations_ideal_source_reward BIGINT,
    attestations_ideal_target_reward BIGINT,
    attestations_ideal_head_reward BIGINT,
    attestations_ideal_inactivity_reward BIGINT,
    attestations_ideal_inclusion_reward BIGINT,
    attestations_ideal_reward BIGINT,
    blocks_scheduled int,
    blocks_proposed int,
    blocks_cl_reward BIGINT, -- gwei
    blocks_el_reward NUMERIC, -- wei
    sync_scheduled int,
    sync_executed int,
    sync_rewards BIGINT,
    slashed BOOLEAN,
    balance_start BIGINT,
    balance_end BIGINT,
    deposits_count int,
    deposits_amount BIGINT,
    withdrawals_count int,
    withdrawals_amount BIGINT,
    inclusion_delay_sum BIGINT,
    sync_chance decimal(8, 7), -- slots_per_epochs / sum(active_validators)
    block_chance decimal(8, 7), -- size_of_sync / number_of_active_validators * slots_per_sync_period
    primary key (validator_index)
);

CREATE TABLE IF NOT EXISTS validator_dashboard_data_rolling_total (
    validator_index int NOT NULL,
    attestations_source_reward BIGINT,
    attestations_target_reward BIGINT,
    attestations_head_reward BIGINT,
    attestations_inactivity_reward BIGINT,
    attestations_inclusion_reward BIGINT,
    attestations_reward BIGINT,
    attestations_ideal_source_reward BIGINT,
    attestations_ideal_target_reward BIGINT,
    attestations_ideal_head_reward BIGINT,
    attestations_ideal_inactivity_reward BIGINT,
    attestations_ideal_inclusion_reward BIGINT,
    attestations_ideal_reward BIGINT,
    blocks_scheduled BIGINT,
    blocks_proposed BIGINT,
    blocks_cl_reward BIGINT, -- gwei
    blocks_el_reward NUMERIC, -- wei
    sync_scheduled BIGINT,
    sync_executed BIGINT,
    sync_rewards BIGINT,
    slashed BOOLEAN,
    balance_start BIGINT,
    balance_end BIGINT,
    deposits_count BIGINT,
    deposits_amount BIGINT,
    withdrawals_count BIGINT,
    withdrawals_amount BIGINT,
    inclusion_delay_sum BIGINT,
    sync_chance decimal(8, 7), -- slots_per_epochs / sum(active_validators)
    block_chance decimal(8, 7), -- size_of_sync / number_of_active_validators * slots_per_sync_period
    primary key (validator_index)
);



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
