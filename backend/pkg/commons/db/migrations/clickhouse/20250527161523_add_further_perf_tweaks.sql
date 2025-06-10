-- +goose Up
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_1h MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_1h ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_1h MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_1h ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
CREATE VIEW IF NOT EXISTS view_validator_dashboard_data_rolling_1h_epoch_minmax AS
SELECT
    min(epoch_start) as epoch_start, max(epoch_end) as epoch_end
FROM
    _final_validator_dashboard_rolling_1h
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_24h MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_24h ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_24h MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_24h ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
CREATE VIEW IF NOT EXISTS view_validator_dashboard_data_rolling_24h_epoch_minmax AS
SELECT
    min(epoch_start) as epoch_start, max(epoch_end) as epoch_end
FROM
    _final_validator_dashboard_rolling_24h
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_7d MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_7d ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_7d MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_7d ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
CREATE VIEW IF NOT EXISTS view_validator_dashboard_data_rolling_7d_epoch_minmax AS
SELECT
    min(epoch_start) as epoch_start, max(epoch_end) as epoch_end
FROM
    _final_validator_dashboard_rolling_7d
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_30d MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_30d ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_30d MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_30d ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
CREATE VIEW IF NOT EXISTS view_validator_dashboard_data_rolling_30d_epoch_minmax AS
SELECT
    min(epoch_start) as epoch_start, max(epoch_end) as epoch_end
FROM
    _final_validator_dashboard_rolling_30d
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_90d MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_90d ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_90d MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_90d ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
CREATE VIEW IF NOT EXISTS view_validator_dashboard_data_rolling_90d_epoch_minmax AS
SELECT
    min(epoch_start) as epoch_start, max(epoch_end) as epoch_end
FROM
    _final_validator_dashboard_rolling_90d
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_total MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _unsafe_validator_dashboard_rolling_total ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_total MODIFY SETTING deduplicate_merge_projection_mode='rebuild'
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE _final_validator_dashboard_rolling_total ADD PROJECTION IF NOT EXISTS epoch_minmax (select min(epoch_start), max(epoch_end))
-- +goose StatementEnd
-- +goose StatementBegin
CREATE VIEW IF NOT EXISTS view_validator_dashboard_data_rolling_total_epoch_minmax AS
SELECT
    min(epoch_start) as epoch_start, max(epoch_end) as epoch_end
FROM
    _final_validator_dashboard_rolling_total
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
