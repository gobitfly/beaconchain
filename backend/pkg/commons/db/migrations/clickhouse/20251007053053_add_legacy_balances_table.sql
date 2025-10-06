-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS legacy_validator_epoch_balances (
   `validator_index` UInt32,
   `epoch_timestamp` DateTime,
   `balance` UInt64,
   `effective_balance` UInt64
)
ENGINE = ReplacingMergeTree
ORDER BY (validator_index, epoch_timestamp);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS legacy_validator_epoch_balances;
-- +goose StatementEnd
