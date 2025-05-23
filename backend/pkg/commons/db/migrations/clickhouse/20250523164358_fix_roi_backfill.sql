-- +goose Up
-- +goose StatementBegin
DROP TABLE _insert_sink_backfill_validator_dashboard_data_roi;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE _insert_sink_backfill_validator_dashboard_data_roi
(
    `validator_index` UInt64 COMMENT 'validator index',
    `t` DateTime COMMENT 'timestamp of the epoch',
    `roi_dividend` Int128 COMMENT 'sum of roi_dividend',
    `roi_divisor` Int128 COMMENT 'sum of roi_divisor'
)
ENGINE = Null
-- +goose StatementEnd
-- +goose StatementBegin
DROP VIEW _mv_tmp_final_validator_dashboard_roi_hourly;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE MATERIALIZED VIEW _mv_tmp_final_validator_dashboard_roi_hourly TO _final_validator_dashboard_roi_hourly
(
    `validator_index` UInt64,
    `t` DateTime,
    `roi_dividend` Int128,
    `roi_divisor` Int128
)
AS SELECT
    validator_index AS validator_index,
    toStartOfHour(t) AS t,
    sum(CAST(foo.roi_dividend, 'Int128')) AS roi_dividend,
    sum(CAST(foo.roi_divisor, 'Int128')) AS roi_divisor
FROM _insert_sink_backfill_validator_dashboard_data_roi AS foo
GROUP BY
    t,
    validator_index;
-- +goose StatementEnd
-- +goose StatementBegin
TRUNCATE TABLE _final_validator_dashboard_roi_hourly;
-- +goose StatementEnd
-- +goose StatementBegin
TRUNCATE TABLE _final_validator_dashboard_roi_daily;
-- +goose StatementEnd
-- +goose StatementBegin
TRUNCATE TABLE _final_validator_dashboard_roi_weekly;
-- +goose StatementEnd
-- +goose StatementBegin
TRUNCATE TABLE _final_validator_dashboard_roi_monthly;
-- +goose StatementEnd
ALTER TABLE _exporter_backfill_metadata DELETE WHERE backfill_name = 'roi';
-- +goose StatementBegin
/* set the highest epoch the backfill process should backfill */
INSERT INTO _exporter_backfill_metadata
    SELECT
        epoch,
        'roi' as backfill_name,
        transfer_batch_id as backfill_batch_id,
        NULL AS successful_backfill
    FROM
    (
        SELECT *
        FROM _exporter_metadata
        WHERE successful_transfer IS NOT NULL
        ORDER BY epoch DESC
        LIMIT 1
    )
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
