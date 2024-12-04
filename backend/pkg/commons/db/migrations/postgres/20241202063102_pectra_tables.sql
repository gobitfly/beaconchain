-- +goose Up
-- +goose StatementBegin
SELECT 'creating blocks_consolidation_requests table';
CREATE TABLE IF NOT EXISTS blocks_consolidation_requests (
    block_slot INT NOT NULL,
    block_root BYTEA NOT NULL,
    request_index INT NOT NULL,
    source_address BYTEA NOT NULL,
    source_pubkey BYTEA NOT NULL,
    target_pubkey BYTEA NOT NULL,
    PRIMARY KEY (block_slot, block_root, request_index)
);
SELECT 'creating blocks_withdrawal_requests table';
CREATE TABLE IF NOT EXISTS blocks_withdrawal_requests (
    block_slot INT NOT NULL,
    block_root BYTEA NOT NULL,
    request_index INT NOT NULL,
    source_address BYTEA NOT NULL,
    validator_pubkey BYTEA NOT NULL,
    amount BIGINT NOT NULL,
    PRIMARY KEY (block_slot, block_root, request_index)
);
SELECT 'creating blocks_deposit_requests table';
CREATE TABLE IF NOT EXISTS blocks_deposit_requests (
    block_slot INT NOT NULL,
    block_root BYTEA NOT NULL,
    request_index INT NOT NULL,
    pubkey BYTEA NOT NULL,
    withdrawal_credentials BYTEA NOT NULL,
    amount BIGINT NOT NULL,
    signature BYTEA NOT NULL,
    index INT NOT NULL,
    PRIMARY KEY (block_slot, block_root, request_index)
);

SELECT 'updating blocks_attestations table';
ALTER TABLE blocks_attestations ADD COLUMN IF NOT EXISTS committeebits BYTEA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'dropping blocks_consolidation_requests table';
DROP TABLE IF EXISTS blocks_consolidation_requests;
-- +goose StatementEnd
