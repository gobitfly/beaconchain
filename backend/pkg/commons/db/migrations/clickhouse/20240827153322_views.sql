-- +goose Up
-- +goose StatementBegin
create or replace view validator_dashboard_data_epoch as 
    select * 
    from _final_validator_dashboard_data_epoch
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_hourly as
    select * 
    from _final_validator_dashboard_data_hourly
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_daily as
    select * 
    from _final_validator_dashboard_data_daily
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_weekly as
    select * 
    from _final_validator_dashboard_data_weekly
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_monthly as
    select * 
    from _final_validator_dashboard_data_monthly
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_epoch_max_ts as
    select * 
    from _final_validator_dashboard_data_epoch
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_hourly_max_ts as
    select * 
    from _final_validator_dashboard_data_hourly
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_daily_max_ts as
    select * 
    from _final_validator_dashboard_data_daily
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_weekly_max_ts as
    select * 
    from _final_validator_dashboard_data_weekly
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_data_monthly_max_ts as
    select * 
    from _final_validator_dashboard_data_monthly
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
