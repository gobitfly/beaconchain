package cl_transformers

import (
	"bytes"
	"slices"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"encoding/json"
)

type WithdrawalProcessedEventTransformer struct{}

func (d *WithdrawalProcessedEventTransformer) Transform(tx *sqlx.Tx, events []types.ConsensusLayerEvent) (int, error) {
	eventName := types.WithdrawalProcessedEventName
	// filter events for the relevant type
	events, err := FilterEventsByTypeAndSort(events, eventName)
	if err != nil {
		return 0, errors.Wrap(err, "error filtering events")
	}
	if len(events) == 0 {
		return 0, nil
	}
	// convert raw data from events into actual structs
	data := make([]types.WithdrawalProcessedEvent, len(events))
	for i, event := range events {
		// unmarshal from data field to DepositQueuedEvent using json
		err = json.Unmarshal(event.RawData, &data[i])
		if err != nil {
			return 0, errors.Wrap(err, "error unmarshalling event data")
		}
	}
	// fetch pre-state from db
	// we filter by status = queued and (pubkey = event.pubkey and withdrawal_credentials = event.withdrawal_credentials and signature = event.signature)
	var orConditions []goqu.Expression
	for _, event := range data {
		orConditions = append(orConditions, goqu.Ex{
			"validator_pubkey": event.Pubkey,
			"amount":           event.OriginalAmount,
		})
	}
	ds := goqu.Dialect("postgres").From("blocks_withdrawal_requests_v2").Select(goqu.Star()).Where(
		goqu.And(
			goqu.C("status").Eq(types.GenericEventStatusQueued),
			goqu.Or(orConditions...),
		)).Order(goqu.C("slot_queued").Asc(), goqu.C("index_queued").Asc())
	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return 0, errors.Wrap(err, "error creating sql query")
	}
	// execute query
	var preState []types.WithdrawalRequestDBRow
	err = tx.Select(&preState, query, args...)
	if err != nil {
		return 0, errors.Wrap(err, "error executing sql query")
	}
	var updatedRows []types.WithdrawalRequestDBRow
	// update in-memory state with events
outer:
	for _, event := range data {
		for i, preStateRow := range preState {
			if bytes.Equal(preStateRow.ValidatorPubkey, event.Pubkey) &&
				preStateRow.Amount == event.OriginalAmount {
				preStateRow.Status = types.GenericEventStatusProcessed
				slotProcessed := int64(event.Slot)
				indexProcessed := int64(event.EventIndex)
				preStateRow.SlotProcessed = &slotProcessed
				preStateRow.IndexProcessed = &indexProcessed
				preStateRow.BlockProcessedRoot = event.BlockRoot
				preStateRow.Amount = event.Amount
				// add row to updatedRows
				updatedRows = append(updatedRows, preStateRow)
				// remove row from preState
				preState = append(preState[:i], preState[i+1:]...)
				continue outer
			}
		}
		log.Warnf("no prestate found for event with pubkey %x and amount %d",
			event.Pubkey, event.OriginalAmount)
		return 0, errors.Errorf("event %v not found in pre-state", event)
	}
	// apply state to db again. shouldn't be using namedExec directly but it do be what it do be
	// chunk to max 2000 rows
	_, err = tx.Exec(`
		CREATE TEMP TABLE tmp_update_table (
			id int4 NOT NULL,
			slot_processed int4 NULL,
			index_processed int2 NULL,
			block_processed_root bytea NULL,
			status text NOT NULL,
			amount int8 NULL
			)
	`)
	if err != nil {
		return 0, errors.Wrap(err, "error creating temp table")
	}

	for chunk := range slices.Chunk(updatedRows, 2000) {
		// use update from. create temporary table with the chunk and then update the main table
		// truncate
		_, err = tx.Exec(`
			TRUNCATE TABLE tmp_update_table
		`)
		if err != nil {
			return 0, errors.Wrap(err, "error truncating temp table")
		}

		// insert chunk into temp table
		_, err = tx.NamedExec(`
			INSERT INTO tmp_update_table (
				id,
				slot_processed,
				index_processed,
				block_processed_root,
				status,
				amount
			)
			VALUES (
				:id,
				:slot_processed,
				:index_processed,
				:block_processed_root,
				:status,
				:amount
			)
		`, chunk)
		if err != nil {
			return 0, errors.Wrap(err, "error inserting new rows into temp table")
		}
		// update main table from temp table
		_, err = tx.Exec(`
			UPDATE blocks_withdrawal_requests_v2
			SET
				slot_processed = temp.slot_processed,
				index_processed = temp.index_processed,
				block_processed_root = temp.block_processed_root,
				status = temp.status,
				amount = temp.amount
			FROM tmp_update_table temp
			WHERE blocks_withdrawal_requests_v2.id = temp.id
		`)
		if err != nil {
			return 0, errors.Wrap(err, "error updating main table from temp table")
		}
	}
	// drop temp table
	_, err = tx.Exec(`
		DROP TABLE tmp_update_table
	`)
	if err != nil {
		return 0, errors.Wrap(err, "error dropping temp table")
	}
	return len(updatedRows), nil
}
