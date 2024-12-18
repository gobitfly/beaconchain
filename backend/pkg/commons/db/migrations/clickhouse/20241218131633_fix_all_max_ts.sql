-- +goose Up
-- +goose StatementBegin
create or replace view view_validator_dashboard_data_hourly_max_ts as
    select max(t) as t
    from _final_validator_dashboard_data_hourly
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view view_validator_dashboard_data_daily_max_ts as
    select max(t) as t
    from _final_validator_dashboard_data_daily
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view view_validator_dashboard_data_weekly_max_ts as
    select max(t) as t
    from _final_validator_dashboard_data_weekly
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view view_validator_dashboard_data_monthly_max_ts as
    select max(t) as t
    from _final_validator_dashboard_data_monthly
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
