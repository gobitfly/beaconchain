-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS eth1_consolidation_requests (
	tx_hash bytea NOT NULL,
	tx_index int4 NOT NULL,
	block_number int4 NOT NULL,
	block_ts timestamp NOT NULL,
	source_address bytea NOT NULL,
	source_pubkey bytea NOT NULL,
	target_pubkey bytea NOT NULL,
	CONSTRAINT eth1_consolidation_requests_pkey PRIMARY KEY (tx_hash, tx_index)
);

CREATE TABLE IF NOT EXISTS eth1_withdrawal_requests (
	tx_hash bytea NOT NULL,
	tx_index int4 NOT NULL,
	block_number int4 NOT NULL,
	block_ts timestamp NOT NULL,
	source_address bytea NOT NULL,
	validator_pubkey bytea NOT NULL,
	amount int8 NOT NULL,
	CONSTRAINT eth1_withdrawal_requests_pkey PRIMARY KEY (tx_hash, tx_index)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF NOT EXISTS eth1_withdrawal_requests;
DROP TABLE IF NOT EXISTS eth1_consolidation_requests;
-- +goose StatementEnd
