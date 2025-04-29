package cl_transformers

import (
	"slices"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"encoding/json"
)

type ExitRequestProcessedEventTransformer struct{}

func (d *ExitRequestProcessedEventTransformer) Transform(tx *sqlx.Tx, events []types.ConsensusLayerEvent) (int, error) {
	eventName := types.ExitRequestProcessedEventName
	// filter events for the relevant type
	events, err := FilterEventsByTypeAndSort(events, eventName)
	if err != nil {
		return 0, errors.Wrap(err, "error filtering events")
	}
	if len(events) == 0 {
		return 0, nil
	}
	// convert raw data from events into actual structs
	data := make([]types.ExitRequestProcessedEvent, len(events))
	for i, event := range events {
		// unmarshal from data field to DepositQueuedEvent using json
		err = json.Unmarshal(event.RawData, &data[i])
		if err != nil {
			return 0, errors.Wrap(err, "error unmarshalling event data")
		}
	}

	// no pre-state required for these events, they are new additions to the table
	newRows := make([]types.ExitRequestDBRow, len(data))
	for i, event := range data {
		slotQueued := int64(event.Slot)
		indexQueued := int64(event.EventIndex)
		newRows[i] = types.ExitRequestDBRow{
			SlotProcessed:      &slotQueued,
			IndexProcessed:     &indexQueued,
			BlockProcessedRoot: event.BlockRoot,

			Status:          types.GenericEventStatusProcessed,
			ValidatorPubkey: event.Pubkey,
		}
	}

	// chunk to max 2000 rows
	for chunk := range slices.Chunk(newRows, 2000) {
		_, err = tx.NamedExec(`
			INSERT INTO 
			blocks_exit_requests (
					slot_processed,
					index_processed,
					block_processed_root,
					"status",
					validator_pubkey
				)
			VALUES (
				:slot_processed,
				:index_processed,
				:block_processed_root,
				:status,
				:validator_pubkey
			)`, chunk)
		if err != nil {
			return 0, errors.Wrap(err, "error inserting new rows into blocks_exit_requests table")
		}
	}
	return len(newRows), nil
}
