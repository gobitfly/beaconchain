-- +goose Up
-- +goose StatementBegin
alter table _final_validator_dashboard_data_hourly
add column if not exists attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_hourly
add column if not exists attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_daily
add column if not exists attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_daily
add column if not exists attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_weekly
add column if not exists attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_weekly
add column if not exists attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_monthly
add column if not exists attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_monthly
add column if not exists attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_hourly materialize column attestations_reward;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_hourly materialize column attestations_ideal_reward;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_daily materialize column attestations_reward;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_daily materialize column attestations_ideal_reward;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_weekly materialize column attestations_reward;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_weekly materialize column attestations_ideal_reward;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_monthly materialize column attestations_reward;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_monthly materialize column attestations_ideal_reward;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
