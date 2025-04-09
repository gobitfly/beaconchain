package db2

import (
	"github.com/jmoiron/sqlx"
)

type BridgeStore interface {
	SetConsolidationRequests(request ConsolidationRequest) error
	SetWithdrawalRequest(request WithdrawalRequest) error
}

type BridgeSQLStore struct {
	db *sqlx.DB
}

func NewBridgeStore(db *sqlx.DB) *BridgeSQLStore {
	return &BridgeSQLStore{
		db: db,
	}
}

func (store BridgeSQLStore) SetConsolidationRequests(request ConsolidationRequest) error {
	_, err := store.db.Exec(`
		INSERT INTO eth1_consolidation_requests (tx_hash, tx_index, block_number, block_ts, source_address, source_pubkey, target_pubkey) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tx_hash, tx_index) DO UPDATE
		SET block_number = EXCLUDED.block_number, block_ts = EXCLUDED.block_ts, source_address = EXCLUDED.source_address, source_pubkey = EXCLUDED.source_pubkey, target_pubkey = EXCLUDED.target_pubkey
	`, request.TxHash, request.TxIndex, request.BlockNumber, request.BlockTimestamp, request.SourceAddress, request.SourcePubKey, request.TargetPubKey)
	if err != nil {
		return err
	}
	return nil
}

func (store BridgeSQLStore) SetWithdrawalRequest(request WithdrawalRequest) error {
	_, err := store.db.Exec(`
		INSERT INTO eth1_withdrawal_requests (tx_hash, tx_index, block_number, block_ts, source_address, validator_pubkey, amount) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tx_hash, tx_index) DO UPDATE
		SET block_number = EXCLUDED.block_number, block_ts = EXCLUDED.block_ts, source_address = EXCLUDED.source_address, validator_pubkey = EXCLUDED.validator_pubkey, amount = EXCLUDED.amount
	`, request.TxHash, request.TxIndex, request.BlockNumber, request.BlockTimestamp, request.SourceAddress, request.ValidatorPubKey, request.Amount)
	if err != nil {
		return err
	}
	return nil
}
