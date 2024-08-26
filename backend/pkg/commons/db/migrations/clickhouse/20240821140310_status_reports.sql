-- +goose Up
-- +goose StatementBegin
CREATE TABLE status_reports
(
    `inserted_at` DateTime64(3) Default now(), -- miliseconds precision
    `expires_at` DateTime Default now() + INTERVAL 1 MINUTE,
    `event_id` LowCardinality(String),
    `emitter` String,
    `status` LowCardinality(String) MATERIALIZED if(has(metadata, 'status'), metadata['status'], 'not_set'),
    `metadata` Map(LowCardinality(String), String),
)
ENGINE = MergeTree()
ORDER BY (inserted_at, event_id, emitter)
TTL toDateTime(inserted_at + toIntervalWeek(1))

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE status_reports IF EXISTS
-- +goose StatementEnd
