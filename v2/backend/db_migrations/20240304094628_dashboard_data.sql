-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS dashboard_data_epoch (
    validatorindex int,
    epoch int,
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
    withdrawals_amount BIGINT
) PARTITION BY range (epoch);

create index on dashboard_data_epoch (validatorindex, epoch);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
