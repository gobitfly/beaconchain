-- +goose Up
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_1h
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_1h
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_1h
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_1h
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_24h
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_24h
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_24h
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_24h
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_7d
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_7d
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_7d
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_7d
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_30d
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_30d
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_30d
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_30d
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_90d
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_90d
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_90d
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_90d
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_total
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_total
add column attestations_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_head_reward_penalties_only + attestations_source_reward_penalties_only + attestations_target_reward_penalties_only + attestations_inclusion_reward_penalties_only + attestations_inactivity_reward_penalties_only +attestations_head_reward_rewards_only + attestations_source_reward_rewards_only + attestations_target_reward_rewards_only + attestations_inclusion_reward_rewards_only + attestations_inactivity_reward_rewards_only
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _unsafe_validator_dashboard_rolling_total
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
-- +goose StatementBegin
alter table _final_validator_dashboard_rolling_total
add column attestations_ideal_reward SimpleAggregateFunction(sum, Int64) Materialized attestations_ideal_head_reward + attestations_ideal_source_reward + attestations_ideal_target_reward + attestations_ideal_inclusion_reward + attestations_ideal_inactivity_reward
settings mutations_sync=2;
-- +goose StatementEnd
--- we dont bother materializing as it will be materialized on the next run anyways
--- we do have to update the views tho
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_1h as
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
    from _final_validator_dashboard_rolling_1h FINAL
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_24h as 
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
    from _final_validator_dashboard_rolling_24h FINAL
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_7d as 
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
    from _final_validator_dashboard_rolling_7d FINAL
    
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_30d as 
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
    from _final_validator_dashboard_rolling_30d FINAL
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_90d as 
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
    from _final_validator_dashboard_rolling_90d FINAL
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_rolling_total as 
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
    from _final_validator_dashboard_rolling_total FINAL
-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
