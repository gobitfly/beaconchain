-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS _exporter_rolling_metadata (
    rolling_name String,
    last_epoch Int64,
    updated_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree
ORDER BY (last_epoch, rolling_name)
PRIMARY KEY (last_epoch, rolling_name)
SETTINGS non_replicated_deduplication_window = 2048, replicated_deduplication_window = 2048;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS _exporter_rolling_metadata;
-- +goose StatementEnd
