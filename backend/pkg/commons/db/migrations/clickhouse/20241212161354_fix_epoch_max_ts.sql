-- +goose Up
-- +goose StatementBegin
create or replace view view_validator_dashboard_data_epoch_max_ts as
    select max(epoch_timestamp) as t
    from _final_validator_dashboard_data_epoch
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
