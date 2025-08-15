-- +goose Up

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
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_data_hourly FINAL settings asterisk_include_alias_columns=1,  asterisk_include_materialized_columns=1
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
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_data_daily FINAL settings asterisk_include_alias_columns=1,  asterisk_include_materialized_columns=1
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
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_data_weekly FINAL settings asterisk_include_alias_columns=1,  asterisk_include_materialized_columns=1
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
    sync_reward_rewards_only + sync_reward_penalties_only AS sync_reward
FROM _final_validator_dashboard_data_monthly FINAL  settings asterisk_include_alias_columns=1,  asterisk_include_materialized_columns=1
-- +goose StatementEnd

-- +goose Down

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
FROM _final_validator_dashboard_data_hourly FINAL
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
FROM _final_validator_dashboard_data_daily FINAL
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
FROM _final_validator_dashboard_data_weekly FINAL
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
FROM _final_validator_dashboard_data_monthly FINAL
-- +goose StatementEnd