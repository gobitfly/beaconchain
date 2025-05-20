-- +goose Up
-- +goose StatementBegin
-- +goose ENVSUB ON
ALTER TABLE _final_validator_dashboard_data_epoch
    UPDATE 
        withdrawals_count = withdrawals_count + dictGet(_dict_backfill_electra_fork_epoch_events, 'withdrawals_count', (epoch_timestamp, validator_index)),
        withdrawals_amount = withdrawals_amount + dictGet(_dict_backfill_electra_fork_epoch_events, 'withdrawals_amount', (epoch_timestamp, validator_index))
    where epoch_timestamp = ${ELECTRA_FORK_EPOCH_TIMESTAMP?please set to the epoch timestamp AFTER the dashboard exporter has prepared the dictionary. set to random date if backfill is not required (i.e the overall export was started after this fix was added)} settings alter_sync=2, mutations_sync=2, database_replicated_enforce_synchronous_settings=1;
-- +goose ENVSUB OFF
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- purposefully fail
select * from "downgrade not supported"."manual re-export recommended"
-- +goose StatementEnd
