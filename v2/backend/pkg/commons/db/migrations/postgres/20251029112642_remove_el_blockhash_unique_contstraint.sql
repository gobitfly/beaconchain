-- +goose Up
-- +goose StatementBegin
ALTER TABLE blocks DROP CONSTRAINT IF EXISTS blocks_exec_block_hash_unique;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE blocks ADD CONSTRAINT IF EXISTS blocks_exec_block_hash_unique UNIQUE (exec_block_hash);
-- +goose StatementEnd
