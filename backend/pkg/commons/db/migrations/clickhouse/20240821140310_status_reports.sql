-- +goose Up
-- +goose StatementBegin
CREATE TABLE status_reports
(
    `insert_id` Int64 CODEC(Delta, ZSTD(1)),
    `inserted_at` DateTime Materialized snowflakeToDateTime(insert_id), -- twitter snowflake epoch
    `timeouts_at` DateTime Default inserted_at + INTERVAL 1 MINUTE,
    `expires_at` DateTime Default timeouts_at + INTERVAL 1 MINUTE,
    `deployment_type` LowCardinality(String),
    `event_id` LowCardinality(String),
    `emitter` UUID,
    `run_id` UUID Materialized if(has(metadata, 'run_id'), metadata['run_id'], null),
    `status` LowCardinality(String) MATERIALIZED if(has(metadata, 'status'), metadata['status'], 'not_set'),
    `metadata` Map(LowCardinality(String), String),
)
ENGINE = MergeTree()
ORDER BY (inserted_at, event_id, emitter)
TTL toDateTime(expires_at + toIntervalMonth(1))

-- +goose StatementEnd  

-- +goose Down
-- +goose StatementBegin
DROP TABLE status_reports IF EXISTS
-- +goose StatementEnd
