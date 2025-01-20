-- +goose Up
-- +goose StatementBegin
-- hourly
create table if not exists _final_validator_dashboard_rolling_1h as _final_validator_dashboard_data_hourly ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose StatementBegin
ALTER table _final_validator_dashboard_rolling_1h MODIFY COLUMN t SimpleAggregateFunction(max, DateTime)
-- +goose StatementEnd
-- +goose StatementBegin
create table if not exists _unsafe_validator_dashboard_rolling_1h as _final_validator_dashboard_rolling_1h ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose StatementBegin
create table if not exists _final_validator_dashboard_rolling_24h as _final_validator_dashboard_rolling_1h ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose StatementBegin
create table if not exists _unsafe_validator_dashboard_rolling_24h as _final_validator_dashboard_rolling_1h ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose StatementBegin
create table if not exists _final_validator_dashboard_rolling_7d as _final_validator_dashboard_rolling_1h ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose StatementBegin
create table if not exists _unsafe_validator_dashboard_rolling_7d as _final_validator_dashboard_rolling_1h ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose StatementBegin
create table if not exists _final_validator_dashboard_rolling_30d as _final_validator_dashboard_rolling_1h ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose StatementBegin
create table if not exists _unsafe_validator_dashboard_rolling_30d as _final_validator_dashboard_rolling_1h ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose StatementBegin
create table if not exists _final_validator_dashboard_rolling_90d as _final_validator_dashboard_rolling_1h ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose StatementBegin
create table if not exists _unsafe_validator_dashboard_rolling_90d as _final_validator_dashboard_rolling_1h ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose StatementBegin
create table if not exists _final_validator_dashboard_rolling_total as _final_validator_dashboard_rolling_1h ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose StatementBegin
create table if not exists _unsafe_validator_dashboard_rolling_total as _final_validator_dashboard_rolling_1h ENGINE AggregatingMergeTree ORDER BY (validator_index)
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table if exists _final_validator_dashboard_rolling_1h
-- +goose StatementEnd
-- +goose StatementBegin
drop table if exists _unsafe_validator_dashboard_rolling_1h
-- +goose StatementEnd
-- +goose StatementBegin
drop table if exists _final_validator_dashboard_rolling_24h
-- +goose StatementEnd
-- +goose StatementBegin
drop table if exists _unsafe_validator_dashboard_rolling_24h
-- +goose StatementEnd
-- +goose StatementBegin
drop table if exists _final_validator_dashboard_rolling_7d
-- +goose StatementEnd
-- +goose StatementBegin
drop table if exists _unsafe_validator_dashboard_rolling_7d
-- +goose StatementEnd
-- +goose StatementBegin
drop table if exists _final_validator_dashboard_rolling_30d
-- +goose StatementEnd
-- +goose StatementBegin
drop table if exists _unsafe_validator_dashboard_rolling_30d
-- +goose StatementEnd
-- +goose StatementBegin
drop table if exists _final_validator_dashboard_rolling_90d
-- +goose StatementEnd
-- +goose StatementBegin
drop table if exists _unsafe_validator_dashboard_rolling_90d
-- +goose StatementEnd
-- +goose StatementBegin
drop table if exists _final_validator_dashboard_rolling_total
-- +goose StatementEnd
-- +goose StatementBegin
drop table if exists _unsafe_validator_dashboard_rolling_total
-- +goose StatementEnd
