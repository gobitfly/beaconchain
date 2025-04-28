package cl_transformers

import (
	"slices"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"encoding/json"
)

type DepositQueuedEventTransformer struct{}

func (d *DepositQueuedEventTransformer) Transform(tx *sqlx.Tx, events []types.ConsensusLayerEvent) (int, error) {
	eventName := types.DepositQueuedEventName
	// filter events for the relevant type
	events, err := FilterEventsByTypeAndSort(events, eventName)
	if err != nil {
		return 0, errors.Wrap(err, "error filtering events")
	}
	if len(events) == 0 {
		return 0, nil
	}
	// convert raw data from events into actual structs
	data := make([]types.DepositQueuedEvent, len(events))
	for i, event := range events {
		// unmarshal from data field to DepositQueuedEvent using json
		err = json.Unmarshal(event.RawData, &data[i])
		if err != nil {
			return 0, errors.Wrap(err, "error unmarshalling event data")
		}
	}

	// no pre-state required for these events, they are new additions to the table
	newRows := make([]types.DepositRequestDBRow, len(data))
	for i, event := range data {
		slotQueued := int64(event.Slot)
		indexQueued := int64(event.EventIndex)
		newRows[i] = types.DepositRequestDBRow{
			SlotQueued:      &slotQueued,
			IndexQueued:     &indexQueued,
			BlockQueuedRoot: event.BlockRoot,

			Type:   types.DepositRequestAccountType,
			Status: types.GenericEventStatusQueued,

			Pubkey:                event.Pubkey,
			WithdrawalCredentials: event.WithdrawalCredentials,
			Amount:                event.Amount,
			Signature:             event.Signature,
		}
	}

	// apply state to db again. shouldn't be using namedExec directly but it do be what it do be
	// chunk to max 2000 rows
	for chunk := range slices.Chunk(newRows, 2000) {
		_, err = tx.NamedExec(`
			INSERT INTO 
				blocks_deposit_requests_v2 (
					slot_queued,
					index_queued,
					block_queued_root,
					"type",
					"status",
					pubkey,
					withdrawal_credentials,
					amount,
					"signature"
				)
			VALUES (
				:slot_queued,
				:index_queued,
				:block_queued_root,
				:type,
				:status,
				:pubkey,
				:withdrawal_credentials,
				:amount,
				:signature
			)`, chunk)
		if err != nil {
			return 0, errors.Wrap(err, "error inserting new rows into blocks_deposit_requests_v2 table")
		}
	}
	return len(newRows), nil
}
