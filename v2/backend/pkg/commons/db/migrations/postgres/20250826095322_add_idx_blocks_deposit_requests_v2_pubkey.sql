-- +goose NO TRANSACTION

-- +goose Up
SELECT 'create idx_blocks_deposit_requests_v2_pubkey';
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_deposit_requests_v2_pubkey ON blocks_deposit_requests_v2 (pubkey);
-- +goose StatementEnd

-- +goose Down
SELECT 'drop idx_blocks_deposit_requests_v2_pubkey';
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_deposit_requests_v2_pubkey;
-- +goose StatementEnd
