-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS
    blocks_deposits (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        proof bytea[],
        publickey bytea NOT NULL,
        withdrawalcredentials bytea NOT NULL,
        amount BIGINT NOT NULL,
        signature bytea NOT NULL,
        valid_signature bool NOT NULL DEFAULT TRUE,
        PRIMARY KEY (block_slot, block_index)
    );
CREATE INDEX IF NOT EXISTS idx_blocks_deposits_publickey ON blocks_deposits (publickey);
CREATE INDEX IF NOT EXISTS idx_blocks_deposits_block_slot_block_root ON public.blocks_deposits USING btree (block_slot, block_root);
CREATE INDEX IF NOT EXISTS idx_blocks_deposits_block_root_publickey ON public.blocks_deposits USING btree (block_root, publickey);

DROP TABLE IF EXISTS validators;
CREATE TABLE IF NOT EXISTS
    validators (
        validatorindex INT NOT NULL,
        pubkey bytea NOT NULL,
        pubkeyhex TEXT NOT NULL DEFAULT '',
        withdrawableepoch BIGINT NOT NULL,
        withdrawalcredentials bytea NOT NULL,
        balance BIGINT NOT NULL,
        balanceactivation BIGINT,
        effectivebalance BIGINT NOT NULL,
        slashed bool NOT NULL,
        activationeligibilityepoch BIGINT NOT NULL,
        activationepoch BIGINT NOT NULL,
        exitepoch BIGINT NOT NULL,
        lastattestationslot BIGINT,
        status VARCHAR(20) NOT NULL DEFAULT '',
        PRIMARY KEY (validatorindex)
    );
CREATE INDEX IF NOT EXISTS idx_validators_pubkey ON validators (pubkey);
CREATE INDEX IF NOT EXISTS idx_validators_pubkeyhex ON validators (pubkeyhex);
CREATE INDEX IF NOT EXISTS idx_validators_pubkeyhex_pattern_pos ON validators (pubkeyhex varchar_pattern_ops);
CREATE INDEX IF NOT EXISTS idx_validators_status ON validators (status);
CREATE INDEX IF NOT EXISTS idx_validators_balanceactivation ON validators (balanceactivation);
CREATE INDEX IF NOT EXISTS idx_validators_activationepoch ON validators (activationepoch);
CREATE INDEX IF NOT EXISTS validators_is_offline_vali_idx ON validators (validatorindex, lastattestationslot, pubkey);
CREATE INDEX IF NOT EXISTS idx_validators_withdrawalcredentials ON validators (withdrawalcredentials, validatorindex);
CREATE INDEX IF NOT EXISTS idx_validators_exitepoch ON validators (exitepoch);
CREATE INDEX IF NOT EXISTS idx_validators_withdrawableepoch ON validators (withdrawableepoch);
CREATE INDEX IF NOT EXISTS idx_validators_lastattestationslot ON validators (lastattestationslot);
CREATE INDEX IF NOT EXISTS idx_validators_activationepoch_status ON public.validators USING btree (activationepoch, status);
CREATE INDEX IF NOT EXISTS idx_validators_activationeligibilityepoch ON public.validators USING btree (activationeligibilityepoch);

do
$$
begin
  if not exists (select * from pg_roles where rolname = 'alloydbsuperuser') then
     create role alloydbsuperuser;
  end if;
  if not exists (select * from pg_roles where rolname = 'readaccess') then
     create role readaccess;
  end if;
end
$$
;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS blocks_deposits;
DROP INDEX IF EXISTS idx_blocks_deposits_publickey;
DROP INDEX IF EXISTS idx_blocks_deposits_block_slot_block_root;
DROP INDEX IF EXISTS idx_blocks_deposits_block_root_publickey;

DROP TABLE IF EXISTS validators;
DROP INDEX IF EXISTS idx_validators_pubkey;
DROP INDEX IF EXISTS idx_validators_pubkeyhex;
DROP INDEX IF EXISTS idx_validators_pubkeyhex_pattern_pos;
DROP INDEX IF EXISTS idx_validators_status;
DROP INDEX IF EXISTS idx_validators_balanceactivation;
DROP INDEX IF EXISTS idx_validators_activationepoch;
DROP INDEX IF EXISTS validators_is_offline_vali_idx;
DROP INDEX IF EXISTS idx_validators_withdrawalcredentials;
DROP INDEX IF EXISTS idx_validators_exitepoch;
DROP INDEX IF EXISTS idx_validators_withdrawableepoch;
DROP INDEX IF EXISTS idx_validators_lastattestationslot;
DROP INDEX IF EXISTS idx_validators_activationepoch_status;
DROP INDEX IF EXISTS idx_validators_activationeligibilityepoch;

-- +goose StatementEnd
