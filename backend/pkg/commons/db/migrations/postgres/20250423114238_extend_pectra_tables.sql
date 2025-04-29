-- +goose Up
-- +goose StatementBegin

ALTER TABLE eth1_consolidation_requests ADD COLUMN IF NOT EXISTS id SERIAL UNIQUE;
ALTER TABLE eth1_withdrawal_requests ADD COLUMN IF NOT EXISTS id SERIAL UNIQUE;

CREATE TABLE IF NOT EXISTS blocks_deposit_requests_v2 (
	id SERIAL PRIMARY KEY,                                      -- would ideally match eth1_id if possible ?
	eth1_id bytea REFERENCES eth1_deposits(merkletree_index),   -- if available (= type is genesis or system_excess)
	-- null if never queued bcus rejected
	slot_queued int4,
	index_queued int4,                                          -- fork state transition includes deposits from all non-active validators
	                                                            -- post-fork max is MAX_DEPOSIT_REQUESTS_PER_PAYLOAD (8,192) + MAX_CONSOLIDATION_REQUESTS_PER_PAYLOAD (2)
	block_queued_root bytea,                                    -- can't fk blocks.blockroot bcus field isn't unique
	-- null if not processed yet bcus queued
	slot_processed int4,                                        -- really an epoch property, but it points to the slot containing the transition
	index_processed int2,                                       -- max MAX_PENDING_DEPOSITS_PER_EPOCH (16)
	block_processed_root bytea,

	type text NOT NULL,                                         -- account, system_excess, genesis
	status text NOT NULL,                                       -- queued, completed, rejected
	reject_reason text,                                         -- invalid_signature

	--  kind of duplicated from eth1_deposits
	pubkey bytea NOT NULL,
	withdrawal_credentials bytea NOT NULL,
	amount int8 NOT NULL,
	signature bytea NOT NULL
);



CREATE TABLE IF NOT EXISTS blocks_consolidation_requests_v2 (
	id SERIAL PRIMARY KEY,
	eth1_id integer REFERENCES eth1_consolidation_requests(id), -- nullable to enable async eth1 event matching
	-- null if never queued bcus rejected
	slot_queued int4,
	index_queued int2,                                          -- max MAX_CONSOLIDATION_REQUESTS_PER_PAYLOAD (2)
	block_queued_root bytea,
	-- null if not processed yet bcus queued
	slot_processed int4,                                        -- really an epoch property, but it points to the slot containing the transition
	index_processed int4,                                       -- max PENDING_CONSOLIDATIONS_LIMIT (262,144)
	block_processed_root bytea,
	
	status text NOT NULL,                                       -- queued, completed, rejected
	reject_reason text,                                         -- ~15
	
	amount_consolidated int8 NULL,
	source_pubkey bytea NOT NULL,
	target_pubkey bytea NOT NULL
);



CREATE TABLE IF NOT EXISTS blocks_withdrawal_requests_v2 (
	id SERIAL PRIMARY KEY,
	eth1_id integer REFERENCES eth1_withdrawal_requests(id),    -- nullable to enable async eth1 event matching
	-- null if never queued bcus rejected
	slot_queued int4,
	index_queued int2,                                          -- max MAX_WITHDRAWAL_REQUESTS_PER_PAYLOAD (16)
	block_queued_root bytea,
	-- null if not processed yet bcus queued
	slot_processed int4,                                        -- really an epoch property, but it points to the slot containing the transition
	index_processed int4,                                       -- max PENDING_PARTIAL_WITHDRAWALS_LIMIT (134,217,728)
	block_processed_root bytea,
	
	status text NOT NULL,                                       -- queued, completed, rejected
	reject_reason text,                                         -- ~15

	validator_pubkey bytea NOT NULL,
	amount decimal NOT NULL
);


CREATE TABLE IF NOT EXISTS blocks_switch_to_compounding_requests_v2 (
	id SERIAL PRIMARY KEY,
	eth1_id integer REFERENCES eth1_consolidation_requests(id),      -- no genesis hardcoding here, should always have an eth1 event
	-- never null because handled pre queue
	slot_processed int4 NOT NULL,                                    -- really an epoch property, but it points to the slot containing the transition
	index_processed int4 NOT NULL,                                   -- max PENDING_CONSOLIDATIONS_LIMIT (262,144)
	block_processed_root bytea NOT NULL,

	status text NOT NULL,                                            -- queued, completed, rejected
	reject_reason text,                                              -- ~10

	validator_pubkey bytea NOT NULL
);

CREATE TABLE IF NOT EXISTS blocks_exit_requests (
	id SERIAL PRIMARY KEY,
	eth1_id integer REFERENCES eth1_withdrawal_requests(id),      -- no genesis hardcoding here, should always have an eth1 event
	-- never null because handled pre queue
	slot_processed int4 NOT NULL,                                    -- really an epoch property, but it points to the slot containing the transition
	index_processed int4 NOT NULL,                                   -- max PENDING_CONSOLIDATIONS_LIMIT (262,144)
	block_processed_root bytea NOT NULL,

	status text NOT NULL,                                            -- queued, completed, rejected
	reject_reason text,                                              -- ~10

	validator_pubkey bytea NOT NULL
);

CREATE TABLE IF NOT EXISTS blocks_removed_excess_balance_events (
	id SERIAL PRIMARY KEY,
	eth1_id integer REFERENCES eth1_withdrawal_requests(id),         -- no genesis hardcoding here, should always have an eth1 event
	-- never null because handled pre queue
	slot_processed int4 NOT NULL,                                    -- really an epoch property, but it points to the slot containing the transition
	index_processed int4 NOT NULL,                                   -- max PENDING_CONSOLIDATIONS_LIMIT (262,144)
	block_processed_root bytea NOT NULL,

	status text NOT NULL,                                            -- queued, completed, rejected
	reject_reason text,                                              -- ~10

	validator_pubkey bytea NOT NULL,
	amount decimal NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
