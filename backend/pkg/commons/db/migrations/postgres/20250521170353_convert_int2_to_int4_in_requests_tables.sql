-- +goose Up
-- +goose StatementBegin
ALTER TABLE blocks_deposit_requests_v2 ALTER COLUMN index_processed TYPE int4;
ALTER TABLE blocks_consolidation_requests_v2 ALTER COLUMN index_processed TYPE int4;
ALTER TABLE blocks_withdrawal_requests_v2 ALTER COLUMN index_queued TYPE int4;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE blocks_deposit_requests_v2 ALTER COLUMN index_processed TYPE int2;
ALTER TABLE blocks_consolidation_requests_v2 ALTER COLUMN index_processed TYPE int2;
ALTER TABLE blocks_withdrawal_requests_v2 ALTER COLUMN index_queued TYPE int2;
-- +goose StatementEnd
