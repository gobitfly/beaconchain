-- +goose Up
-- +goose ENVSUB ON
-- +goose StatementBegin
set
    param_fork_epoch_ts = ${ELECTRA_FORK_EPOCH_TIMESTAMP?missing},
    param_genesis_epoch_ts = ${GENESIS_EPOCH_TIMESTAMP?missing},
    param_epoch_duration = ${EPOCH_DURATION_IN_SECONDS?missing};
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_hourly
    UPDATE
        epoch_start = GREATEST(
            0,
            CEIL(
                (
                    toStartOfHour({fork_epoch_ts:DateTime})::DateTime - {genesis_epoch_ts:DateTime}
                ) / {epoch_duration:Int64}
            )
        )
    WHERE t = toStartOfHour({fork_epoch_ts:DateTime})
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_daily
    UPDATE
        epoch_start = GREATEST(
            0,
            CEIL(
                (
                    toStartOfDay({fork_epoch_ts:DateTime})::DateTime - {genesis_epoch_ts:DateTime}
                ) / {epoch_duration:Int64}
            )
        )
    WHERE t = toStartOfDay({fork_epoch_ts:DateTime})
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_weekly
    UPDATE
        epoch_start = GREATEST(
            0,
            CEIL(
                (
                    toMonday({fork_epoch_ts:DateTime})::DateTime - {genesis_epoch_ts:DateTime}
                ) / {epoch_duration:Int64}
            )
        )
    WHERE t = toMonday({fork_epoch_ts:DateTime})
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_data_monthly
    UPDATE
        epoch_start = GREATEST(
            0,
            CEIL(
                (
                    toStartOfMonth({fork_epoch_ts:DateTime})::DateTime - {genesis_epoch_ts:DateTime}
                ) / {epoch_duration:Int64}
            )
        )
    WHERE t = toStartOfMonth({fork_epoch_ts:DateTime})
-- +goose StatementEnd
-- +goose ENVSUB OFF

-- +goose Down
-- +goose StatementBegin
select * from "downgrade not supported"."manual re-export recommended"
-- +goose StatementEnd
