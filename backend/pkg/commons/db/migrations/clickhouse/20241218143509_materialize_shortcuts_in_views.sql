-- +goose Up
-- +goose StatementBegin
alter table _final_validator_dashboard_data_epoch
add column if not exists attestations_reward Int64 Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_epoch
add column if not exists attestations_ideal_reward Int64 Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_hourly
add column if not exists attestations_reward Int64 Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_hourly
add column if not exists attestations_ideal_reward Int64 Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_daily
add column if not exists attestations_reward Int64 Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_daily
add column if not exists attestations_ideal_reward Int64 Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_weekly
add column if not exists attestations_reward Int64 Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_weekly
add column if not exists attestations_ideal_reward Int64 Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_monthly
add column if not exists attestations_reward Int64 Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=1;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_monthly
add column if not exists attestations_ideal_reward Int64 Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=1;
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
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only as sync_reward
    from _final_validator_dashboard_data_epoch;
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_hourly as
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
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only as sync_reward
    from _final_validator_dashboard_data_hourly
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_daily as
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
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only as sync_reward
    from _final_validator_dashboard_data_daily
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_weekly as
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
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only as sync_reward
    from _final_validator_dashboard_data_weekly
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_monthly as
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
    attestations_reward,
    attestations_ideal_reward,
    sync_reward_rewards_only + sync_reward_penalties_only as sync_reward
    from _final_validator_dashboard_data_monthly
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_epoch materialize column attestations_reward;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_data_epoch materialize column attestations_ideal_reward;
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
Select('im not writing a down migration for this, just dont migrate down');
-- +goose StatementEnd
