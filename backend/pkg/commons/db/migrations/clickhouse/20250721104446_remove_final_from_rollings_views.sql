-- +goose Up
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
FROM _final_validator_dashboard_rolling_1h
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
FROM _final_validator_dashboard_rolling_24h
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
FROM _final_validator_dashboard_rolling_7d
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
FROM _final_validator_dashboard_rolling_30d
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
FROM _final_validator_dashboard_rolling_90d
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
FROM _final_validator_dashboard_rolling_total
-- +goose StatementEnd

-- +goose Down
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
FROM _final_validator_dashboard_rolling_1h FINAL
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
FROM _final_validator_dashboard_rolling_24h FINAL
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
FROM _final_validator_dashboard_rolling_7d FINAL
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
FROM _final_validator_dashboard_rolling_30d FINAL
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
FROM _final_validator_dashboard_rolling_90d FINAL
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
FROM _final_validator_dashboard_rolling_total FINAL
-- +goose StatementEnd#
