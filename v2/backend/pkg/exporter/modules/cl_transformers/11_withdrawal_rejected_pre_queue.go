package cl_transformers

import (
	"slices"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"encoding/json"
)

type WithdrawalRejectedPreQueueEventTransformer struct{}

func (d *WithdrawalRejectedPreQueueEventTransformer) Transform(tx *sqlx.Tx, events []types.ConsensusLayerEvent) (int, error) {
	eventName := types.WithdrawalRejectedEventName
	// filter events for the relevant type
	events, err := FilterEventsByTypeAndSort(events, eventName)
	if err != nil {
		return 0, errors.Wrap(err, "error filtering events")
	}
	if len(events) == 0 {
		return 0, nil
	}
	// convert raw data from events into actual structs
	data := make([]types.WithdrawalRejectedEvent, len(events))
	for i, event := range events {
		// unmarshal from data field to DepositQueuedEvent using json
		err = json.Unmarshal(event.RawData, &data[i])
		if err != nil {
			return 0, errors.Wrap(err, "error unmarshalling event data")
		}
	}

	// additionally filter for events that have PreQueue == true
	var filteredData []types.WithdrawalRejectedEvent
	for _, event := range data {
		if event.PreQueue {
			filteredData = append(filteredData, event)
		}
	}
	if len(filteredData) == 0 {
		return 0, nil
	}

	// same as queued, no pre-state required for these events, they are new additions to the table
	newRows := make([]types.WithdrawalRequestDBRow, len(data))
	for i, event := range data {
		slotProcessed := int64(event.Slot)
		indexProcessed := int64(event.EventIndex)
		newRows[i] = types.WithdrawalRequestDBRow{
			SlotProcessed:      &slotProcessed,
			IndexProcessed:     &indexProcessed,
			BlockProcessedRoot: event.BlockRoot,

			Status:       types.GenericEventStatusRejected,
			RejectReason: &event.Reason,

			ValidatorPubkey: event.Pubkey,
			Amount:          event.Amount,
		}
	}

	// apply state to db again. shouldn't be using namedExec directly but it do be what it do be
	// chunk to max 2000 rows
	for chunk := range slices.Chunk(newRows, 2000) {
		_, err = tx.NamedExec(`
			INSERT INTO 
			blocks_withdrawal_requests_v2 (
					slot_processed,
					index_processed,
					block_processed_root,
					"status",
					reject_reason,
					validator_pubkey,
					amount
				)
			VALUES (
				:slot_processed,
				:index_processed,
				:block_processed_root,
				:status,
				:reject_reason,
				:validator_pubkey,
				:amount
			)`, chunk)
		if err != nil {
			return 0, errors.Wrap(err, "error inserting new rows into blocks_withdrawal_requests_v2 table")
		}
	}
	return len(newRows), nil
}
