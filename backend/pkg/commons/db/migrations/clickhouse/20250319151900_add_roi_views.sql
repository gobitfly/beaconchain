-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_roi_hourly AS 
SELECT 
    *
FROM 
    _final_validator_dashboard_roi_hourly FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_roi_daily AS 
SELECT 
    *
FROM 
    _final_validator_dashboard_roi_daily FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_roi_weekly AS 
SELECT 
    *
FROM 
    _final_validator_dashboard_roi_weekly FINAL
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW validator_dashboard_roi_monthly AS 
SELECT 
    *
FROM 
    _final_validator_dashboard_roi_monthly FINAL
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW validator_dashboard_roi_hourly
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW validator_dashboard_roi_daily
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW validator_dashboard_roi_weekly
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW validator_dashboard_roi_monthly
-- +goose StatementEnd
