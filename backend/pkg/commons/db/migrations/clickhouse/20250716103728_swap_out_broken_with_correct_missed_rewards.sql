-- +goose Up
-- +goose StatementBegin
DROP VIEW IF EXISTS _mv_backfill_final_validator_dashboard_data_hourly_missing_missed_rewards
SETTINGS alter_sync = 2, mutations_sync = 3, database_replicated_enforce_synchronous_settings = 1;
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW IF EXISTS _mv_backfill_final_validator_dashboard_data_daily_missing_missed_rewards
SETTINGS alter_sync = 2, mutations_sync = 3, database_replicated_enforce_synchronous_settings = 1
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW IF EXISTS _mv_backfill_final_validator_dashboard_data_weekly_missing_missed_rewards
SETTINGS alter_sync = 2, mutations_sync = 3, database_replicated_enforce_synchronous_settings = 1;
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW IF EXISTS _mv_backfill_final_validator_dashboard_data_monthly_missing_missed_rewards
SETTINGS alter_sync = 2, mutations_sync = 3, database_replicated_enforce_synchronous_settings = 1;
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW IF EXISTS _mv_backfill_forwards_sync_final_validator_dashboard_data_hourly_missing_missed_rewards
SETTINGS alter_sync = 2, mutations_sync = 3, database_replicated_enforce_synchronous_settings = 1;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS _insert_sink_backfill_missing_missed_rewards
SETTINGS alter_sync = 2, mutations_sync = 3, database_replicated_enforce_synchronous_settings = 1;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    (RENAME COLUMN blocks_cl_missed_median_reward TO _trash),
    (RENAME COLUMN _fixed_blocks_cl_missed_median_reward TO blocks_cl_missed_median_reward),
    (MODIFY COLUMN `efficiency_proposals_divisor` SimpleAggregateFunction(sum, Int64) MATERIALIZED blocks_cl_reward + blocks_cl_missed_median_reward)
SETTINGS alter_sync = 2, mutations_sync = 3, database_replicated_enforce_synchronous_settings = 1;
-- +goose StatementEnd
-- +goose StatementBegin
-- dont wait for mutation, this can complete in the background
ALTER TABLE _final_validator_dashboard_data_hourly
    (DROP COLUMN IF EXISTS _trash),
    (CLEAR COLUMN efficiency_proposals_divisor) -- make the new fixed data instantly available
SETTINGS alter_sync = 2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    (MATERIALIZE COLUMN efficiency_proposals_divisor)
SETTINGS alter_sync=2;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    (RENAME COLUMN blocks_cl_missed_median_reward TO _trash),
    (RENAME COLUMN _fixed_blocks_cl_missed_median_reward TO blocks_cl_missed_median_reward),
    (MODIFY COLUMN `efficiency_proposals_divisor` SimpleAggregateFunction(sum, Int64) MATERIALIZED blocks_cl_reward + blocks_cl_missed_median_reward)
SETTINGS alter_sync = 2, mutations_sync = 3, database_replicated_enforce_synchronous_settings = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    (DROP COLUMN IF EXISTS _trash),
    (CLEAR COLUMN efficiency_proposals_divisor)
SETTINGS alter_sync = 2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    (MATERIALIZE COLUMN efficiency_proposals_divisor)
SETTINGS alter_sync=2;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    (RENAME COLUMN blocks_cl_missed_median_reward TO _trash),
    (RENAME COLUMN _fixed_blocks_cl_missed_median_reward TO blocks_cl_missed_median_reward),
    (MODIFY COLUMN `efficiency_proposals_divisor` SimpleAggregateFunction(sum, Int64) MATERIALIZED blocks_cl_reward + blocks_cl_missed_median_reward)
SETTINGS alter_sync = 2, mutations_sync = 3, database_replicated_enforce_synchronous_settings = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    (DROP COLUMN IF EXISTS _trash),
    (CLEAR COLUMN efficiency_proposals_divisor)
SETTINGS alter_sync = 2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    (MATERIALIZE COLUMN efficiency_proposals_divisor)
SETTINGS alter_sync=2;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    (RENAME COLUMN blocks_cl_missed_median_reward TO _trash),
    (RENAME COLUMN _fixed_blocks_cl_missed_median_reward TO blocks_cl_missed_median_reward),
    (MODIFY COLUMN `efficiency_proposals_divisor` SimpleAggregateFunction(sum, Int64) MATERIALIZED blocks_cl_reward + blocks_cl_missed_median_reward)
SETTINGS alter_sync = 2, mutations_sync = 3, database_replicated_enforce_synchronous_settings = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    (DROP COLUMN IF EXISTS _trash),
    (CLEAR COLUMN efficiency_proposals_divisor)
SETTINGS alter_sync = 2;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    (MATERIALIZE COLUMN efficiency_proposals_divisor)
SETTINGS alter_sync=2;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
