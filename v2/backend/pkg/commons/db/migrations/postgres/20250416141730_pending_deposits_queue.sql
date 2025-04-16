-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS pending_deposits_queue (
    id INT NOT NULL,
    validator_index BIGINT, -- nullable, null = deposit, not null = topup
    pubkey BYTEA NOT NULL,
    withdrawal_credentials BYTEA NOT NULL,
    amount BIGINT NOT NULL, --gwei
    signature BYTEA NOT NULL,
    slot BIGINT NOT NULL,
    queued_balance_ahead BIGINT NOT NULL, --gwei
    est_clear_epoch BIGINT NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_pending_deposits_queue_pubkey ON pending_deposits_queue (pubkey);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE pending_deposits_queue;

-- +goose StatementEnd