-- +goose Up
-- +goose StatementBegin
create or replace view validator_dashboard_rolling_1h as
    select * 
    from _final_validator_dashboard_rolling_1h
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_rolling_24h as
    select * 
    from _final_validator_dashboard_rolling_24h
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_rolling_7d as
    select * 
    from _final_validator_dashboard_rolling_7d
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_rolling_30d as
    select * 
    from _final_validator_dashboard_rolling_30d
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_rolling_90d as
    select * 
    from _final_validator_dashboard_rolling_90d
-- +goose StatementEnd
-- +goose StatementBegin
create or replace view validator_dashboard_rolling_total as
    select * 
    from _final_validator_dashboard_rolling_total
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop view if exists validator_dashboard_rolling_1h
-- +goose StatementEnd
-- +goose StatementBegin
drop view if exists validator_dashboard_rolling_24h
-- +goose StatementEnd
-- +goose StatementBegin
drop view if exists validator_dashboard_rolling_7d
-- +goose StatementEnd
-- +goose StatementBegin
drop view if exists validator_dashboard_rolling_30d
-- +goose StatementEnd
-- +goose StatementBegin
drop view if exists validator_dashboard_rolling_90d
-- +goose StatementEnd
-- +goose StatementBegin
drop view if exists validator_dashboard_rolling_total
-- +goose StatementEnd
