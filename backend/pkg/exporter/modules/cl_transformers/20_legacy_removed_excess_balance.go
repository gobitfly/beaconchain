package cl_transformers

import (
	"fmt"
	"slices"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"encoding/json"
)

type LegacyRemovedExcessBalanceEventTransformer struct{}

func (d *LegacyRemovedExcessBalanceEventTransformer) Transform(tx *sqlx.Tx, events []types.ConsensusLayerEvent) (int, error) {
	eventName := types.RemovedExcessBalanceEventName
	// filter events for the relevant type
	events, err := FilterEventsByTypeAndSort(events, eventName)
	if err != nil {
		return 0, errors.Wrap(err, "error filtering events")
	}
	if len(events) == 0 {
		return 0, nil
	}
	// convert raw data from events into actual structs
	data := make([]types.RemovedExcessBalance, len(events))
	for i, event := range events {
		// unmarshal from data field to DepositQueuedEvent using json
		err = json.Unmarshal(event.RawData, &data[i])
		if err != nil {
			return 0, errors.Wrap(err, "error unmarshalling event data")
		}
	}

	// get validator pubkey => validator index from postgres
	var validatorMapping []struct {
		Pubkey []byte `db:"pubkey"`
		Index  uint64 `db:"validatorindex"`
	}

	err = tx.Select(&validatorMapping, `
		SELECT pubkey, validatorindex
		FROM validators
		`)
	if err != nil {
		return 0, fmt.Errorf("error getting validator mapping: %w", err)
	}
	// create a map of pubkey => validator index
	pubkeyToIndex := make(map[string]uint64)
	for _, mapping := range validatorMapping {
		pubkeyToIndex[string(mapping.Pubkey)] = mapping.Index
	}
	type BlocksWithdrawalRow struct {
		BlockSlot       uint64 `db:"block_slot"`
		BlockRoot       []byte `db:"block_root"`
		WithdrawalIndex int64  `db:"withdrawalindex"`
		ValidatorIndex  uint64 `db:"validatorindex"`
		Address         []byte `db:"address"`
		Amount          int64  `db:"amount"`
	}
	newRows := make([]BlocksWithdrawalRow, len(events))
	for i, event := range data {
		// get the validator index from the pubkey
		pubkey := string(event.Pubkey)
		validatorIndex, ok := pubkeyToIndex[pubkey]
		if !ok {
			return 0, fmt.Errorf("error getting validator index for pubkey %s", pubkey)
		}
		// create a new row for the blocks_withdrawals table
		newRows[i] = BlocksWithdrawalRow{
			BlockSlot:       event.Slot,
			BlockRoot:       event.BlockRoot,
			WithdrawalIndex: -20000 + int64(event.EventIndex),
			ValidatorIndex:  validatorIndex,
			Address:         []byte{},
			Amount:          int64(event.Amount),
		}
	}

	// chunk to max 2000 rows
	for chunk := range slices.Chunk(newRows, 2000) {
		log.Debugf("inserting %d new rows into blocks_withdrawals table", len(chunk))
		_, err = tx.NamedExec(`
			INSERT INTO
			blocks_withdrawals (
					block_slot,
					block_root,
					withdrawalindex,
					validatorindex,
					address,
					amount
				)
			VALUES (
				:block_slot,
				:block_root,
				:withdrawalindex,
				:validatorindex,
				:address,
				:amount
			) ON CONFLICT DO NOTHING`, chunk)
		if err != nil {
			return 0, errors.Wrap(err, "error inserting new rows into blocks_removed_excess_balance_events table")
		}
	}

	return len(newRows), nil
}
