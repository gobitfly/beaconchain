-- +goose Up
-- +goose StatementBegin
ALTER TABLE pending_deposits_queue ADD COLUMN IF NOT EXISTS request_id BIGINT REFERENCES blocks_deposit_requests_v2(id) UNIQUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE pending_deposits_queue DROP COLUMN IF EXISTS request_id;
-- +goose StatementEnd
