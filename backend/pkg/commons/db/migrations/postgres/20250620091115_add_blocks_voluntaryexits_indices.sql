-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_blocks_voluntaryexits_validatorindex ON blocks_voluntaryexits (validatorindex);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_blocks_voluntaryexits_validatorindex;
-- +goose StatementEnd
