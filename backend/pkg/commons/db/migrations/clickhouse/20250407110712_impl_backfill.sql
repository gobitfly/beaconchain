-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _exporter_backfill_metadata
(
    `epoch` UInt64 COMMENT 'epoch number',
    `backfill_name` String COMMENT 'name of the backfill',
    `backfill_batch_id` Nullable(UUID) COMMENT 'id of the batch the epoch is part of during backfill',
    `successful_backfill` Nullable(DateTime) COMMENT 'if the batch was successfully backfilled. This is set after the exporter has received confirmation from the clickhouse server'
)
ENGINE = ReplacingMergeTree
ORDER BY (epoch, backfill_name)
SETTINGS index_granularity = 8192, non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS _exporter_backfill_metadata;
-- +goose StatementEnd
