-- +goose Up

-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_epoch ADD INDEX IF NOT EXISTS _index_epoch_timestamp_min_max epoch_timestamp type minmax() granularity 1 SETTINGS alter_sync = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_epoch MATERIALIZE INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_data_epoch ADD INDEX IF NOT EXISTS _index_epoch_timestamp_min_max epoch_timestamp type minmax() granularity 1 SETTINGS alter_sync = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_data_epoch MATERIALIZE INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly ADD INDEX IF NOT EXISTS _index_t_min_max t type minmax() granularity 1 SETTINGS alter_sync = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly MATERIALIZE INDEX _index_t_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily ADD INDEX IF NOT EXISTS _index_t_min_max t type minmax() granularity 1 SETTINGS alter_sync = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily MATERIALIZE INDEX _index_t_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly ADD INDEX IF NOT EXISTS _index_t_min_max t type minmax() granularity 1 SETTINGS alter_sync = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly MATERIALIZE INDEX _index_t_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly ADD INDEX IF NOT EXISTS _index_t_min_max t type minmax() granularity 1 SETTINGS alter_sync = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly MATERIALIZE INDEX _index_t_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE validator_attestation_assignments_slot ADD INDEX IF NOT EXISTS _index_epoch_timestamp_min_max epoch_timestamp type minmax() granularity 1 SETTINGS alter_sync = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE validator_attestation_assignments_slot MATERIALIZE INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE validator_proposal_assignments_slot ADD INDEX IF NOT EXISTS _index_epoch_timestamp_min_max epoch_timestamp type minmax() granularity 1 SETTINGS alter_sync = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE validator_proposal_assignments_slot MATERIALIZE INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE validator_proposal_rewards_slot ADD INDEX IF NOT EXISTS _index_epoch_timestamp_min_max epoch_timestamp type minmax() granularity 1 SETTINGS alter_sync = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE validator_proposal_rewards_slot MATERIALIZE INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE validator_sync_committee_rewards_slot ADD INDEX IF NOT EXISTS _index_epoch_timestamp_min_max epoch_timestamp type minmax() granularity 1 SETTINGS alter_sync = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE validator_sync_committee_rewards_slot MATERIALIZE INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE validator_sync_committee_votes_slot ADD INDEX IF NOT EXISTS _index_epoch_timestamp_min_max epoch_timestamp type minmax() granularity 1 SETTINGS alter_sync = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE validator_sync_committee_votes_slot MATERIALIZE INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_epoch DROP INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_data_epoch DROP INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly DROP INDEX _index_t_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily DROP INDEX _index_t_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly DROP INDEX _index_t_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly DROP INDEX _index_t_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_attestation_assignments_slot DROP INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_proposal_assignments_slot DROP INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_proposal_rewards_slot DROP INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_sync_committee_rewards_slot DROP INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_sync_committee_votes_slot DROP INDEX _index_epoch_timestamp_min_max;
-- +goose StatementEnd
