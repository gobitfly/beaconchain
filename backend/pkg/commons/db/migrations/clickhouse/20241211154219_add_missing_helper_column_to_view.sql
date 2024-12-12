-- +goose Up
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_data_epoch DROP COLUMN IF EXISTS sync_reward;  -- whoopsie
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_epoch DROP COLUMN IF EXISTS sync_reward;  -- whoopsie
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_epoch as 
    select *, 
    attestations_head_reward_penalties_only + attestations_head_reward_rewards_only as attestations_head_reward,
    attestations_source_reward_penalties_only + attestations_source_reward_rewards_only as attestations_source_reward,
    attestations_target_reward_penalties_only + attestations_target_reward_rewards_only as attestations_target_reward,
    attestations_inclusion_reward_penalties_only + attestations_inclusion_reward_rewards_only as attestations_inclusion_reward,
    attestations_inactivity_reward_penalties_only + attestations_inactivity_reward_rewards_only as attestations_inactivity_reward,
    (
        attestations_head_reward_rewards_only +
        attestations_source_reward_rewards_only +
        attestations_target_reward_rewards_only +
        attestations_inclusion_reward_rewards_only +
        attestations_inactivity_reward_rewards_only
    ) as attestations_reward_rewards_only,
    (
        attestations_head_reward_penalties_only +
        attestations_source_reward_penalties_only +
        attestations_target_reward_penalties_only +
        attestations_inclusion_reward_penalties_only +
        attestations_inactivity_reward_penalties_only
    ) as attestations_reward_penalties_only,
    attestations_reward_penalties_only + attestations_reward_rewards_only as attestations_reward,
    (
        attestations_ideal_head_reward + 
        attestations_ideal_source_reward +
        attestations_ideal_target_reward +
        attestations_ideal_inclusion_reward +
        attestations_ideal_inactivity_reward
    ) as attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only as sync_reward
    from _final_validator_dashboard_data_epoch;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- not needed

-- +goose StatementEnd