package cl_transformers

import (
	"slices"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"encoding/json"
)

type WithdrawalQueuedEventTransformer struct{}

func (d *WithdrawalQueuedEventTransformer) Transform(tx *sqlx.Tx, events []types.ConsensusLayerEvent) (int, error) {
	eventName := types.WithdrawalQueuedEventName
	// filter events for the relevant type
	events, err := FilterEventsByTypeAndSort(events, eventName)
	if err != nil {
		return 0, errors.Wrap(err, "error filtering events")
	}
	if len(events) == 0 {
		return 0, nil
	}
	// convert raw data from events into actual structs
	data := make([]types.WithdrawalQueuedEvent, len(events))
	for i, event := range events {
		// unmarshal from data field to DepositQueuedEvent using json
		err = json.Unmarshal(event.RawData, &data[i])
		if err != nil {
			return 0, errors.Wrap(err, "error unmarshalling event data")
		}
	}

	// no pre-state required for these events, they are new additions to the table
	newRows := make([]types.WithdrawalRequestDBRow, len(data))
	for i, event := range data {
		slotQueued := int64(event.Slot)
		indexQueued := int64(event.EventIndex)
		newRows[i] = types.WithdrawalRequestDBRow{
			SlotQueued:      &slotQueued,
			IndexQueued:     &indexQueued,
			BlockQueuedRoot: event.BlockRoot,

			Status: types.GenericEventStatusQueued,

			ValidatorPubkey: event.Pubkey,
			Amount:          event.Amount,
		}
	}

	// chunk to max 2000 rows
	for chunk := range slices.Chunk(newRows, 2000) {
		_, err = tx.NamedExec(`
			INSERT INTO 
			blocks_withdrawal_requests_v2 (
					slot_queued,
					index_queued,
					block_queued_root,
					"status",
					validator_pubkey,
					amount
				)
			VALUES (
				:slot_queued,
				:index_queued,
				:block_queued_root,
				:status,
				:validator_pubkey,
				:amount
			)`, chunk)
		if err != nil {
			return 0, errors.Wrap(err, "error inserting new rows into blocks_withdrawal_requests_v2 table")
		}
	}
	return len(newRows), nil
}
