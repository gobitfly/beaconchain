-- +goose Up
-- +goose StatementBegin
ALTER TABLE eth1_consolidation_requests ADD COLUMN IF NOT EXISTS itx_index int4 NOT NULL DEFAULT 0;
ALTER TABLE eth1_consolidation_requests ADD COLUMN IF NOT EXISTS from_address bytea NOT NULL DEFAULT '\x';
ALTER TABLE eth1_consolidation_requests ADD COLUMN IF NOT EXISTS fee BIGINT NOT NULL DEFAULT 0;

ALTER TABLE eth1_consolidation_requests DROP CONSTRAINT IF EXISTS eth1_consolidation_requests_pkey;
ALTER TABLE eth1_consolidation_requests ADD CONSTRAINT eth1_consolidation_requests_pkey UNIQUE (tx_hash, tx_index, itx_index);

ALTER TABLE eth1_consolidation_requests ADD COLUMN id BIGSERIAL;
ALTER TABLE eth1_consolidation_requests ADD CONSTRAINT eth1_consolidation_requests_id_key UNIQUE (id);

ALTER TABLE eth1_withdrawal_requests ADD COLUMN IF NOT EXISTS itx_index int4 NOT NULL DEFAULT 0;
ALTER TABLE eth1_withdrawal_requests ADD COLUMN IF NOT EXISTS from_address bytea NOT NULL DEFAULT '\x';
ALTER TABLE eth1_withdrawal_requests ADD COLUMN IF NOT EXISTS fee BIGINT NOT NULL DEFAULT 0;

ALTER TABLE eth1_withdrawal_requests DROP CONSTRAINT IF EXISTS eth1_withdrawal_requests_pkey;
ALTER TABLE eth1_withdrawal_requests ADD CONSTRAINT eth1_withdrawal_requests_pkey UNIQUE (tx_hash, tx_index, itx_index);
ALTER TABLE eth1_withdrawal_requests ADD COLUMN id BIGSERIAL;
ALTER TABLE eth1_withdrawal_requests ADD CONSTRAINT eth1_withdrawal_requests_id_key UNIQUE (id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE eth1_consolidation_requests DROP COLUMN itx_index;
ALTER TABLE eth1_consolidation_requests DROP COLUMN from_address;
ALTER TABLE eth1_consolidation_requests DROP COLUMN fee;
ALTER TABLE eth1_consolidation_requests DROP COLUMN id;

ALTER TABLE eth1_withdrawal_requests DROP COLUMN itx_index;
ALTER TABLE eth1_withdrawal_requests DROP COLUMN from_address;
ALTER TABLE eth1_withdrawal_requests DROP COLUMN fee;
ALTER TABLE eth1_withdrawal_requests DROP COLUMN id;
-- +goose StatementEnd
