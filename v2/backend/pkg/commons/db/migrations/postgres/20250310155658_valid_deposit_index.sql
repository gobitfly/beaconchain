-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_blocks_deposits_valid_signature ON blocks_deposits (publickey, block_slot, block_index) WHERE valid_signature;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_blocks_deposits_valid_signature;
-- +goose StatementEnd
