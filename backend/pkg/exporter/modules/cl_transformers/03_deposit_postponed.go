package cl_transformers

import (
	"bytes"
	"slices"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"encoding/json"
)

type DepositPostponedEventTransformer struct{}

func (d *DepositPostponedEventTransformer) Transform(tx *sqlx.Tx, events []types.ConsensusLayerEvent) (int, error) {
	eventName := types.DepositPostponeEventName
	// filter events for the relevant type
	events, err := FilterEventsByTypeAndSort(events, eventName)
	if err != nil {
		return 0, errors.Wrap(err, "error filtering events")
	}
	if len(events) == 0 {
		return 0, nil
	}
	// convert raw data from events into actual structs
	data := make([]types.DepositPostponedEvent, len(events))
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
			"pubkey":                 event.Pubkey,
			"withdrawal_credentials": event.WithdrawalCredentials,
			"signature":              event.Signature,
			"amount":                 event.Amount,
		})
	}
	ds := goqu.Dialect("postgres").From("blocks_deposit_requests_v2").Select(goqu.Star()).Where(
		goqu.And(
			goqu.C("status").Eq(types.GenericEventStatusQueued),
			goqu.Or(orConditions...),
		),
	)
	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return 0, errors.Wrap(err, "error creating sql query")
	}
	// execute query
	var preState []types.DepositRequestDBRow
	err = tx.Select(&preState, query, args...)
	if err != nil {
		return 0, errors.Wrap(err, "error executing sql query")
	}
	var updatedRows []types.DepositRequestDBRow
	// update in-memory state with events
outer:
	for _, event := range data {
		for i, preStateRow := range preState {
			if bytes.Equal(preStateRow.Pubkey, event.Pubkey) &&
				bytes.Equal(preStateRow.WithdrawalCredentials, event.WithdrawalCredentials) &&
				bytes.Equal(preStateRow.Signature, event.Signature) &&
				preStateRow.Amount == event.Amount {
				// update status to postponed
				preStateRow.Status = types.GenericEventStatusPostponed
				slotQueued := int64(event.Slot)
				indexQueued := int64(event.EventIndex)
				preStateRow.SlotQueued = &slotQueued
				preStateRow.IndexQueued = &indexQueued
				preStateRow.BlockQueuedRoot = event.BlockRoot
				// add row to updatedRows
				updatedRows = append(updatedRows, preStateRow)
				// remove row from preState
				preState = append(preState[:i], preState[i+1:]...)
				continue outer
			}
		}
		return 0, errors.Errorf("event %v not found in pre-state", event)
	}
	// apply state to db again. shouldn't be using namedExec directly but it do be what it do be
	// chunk to max 2000 rows
	for chunk := range slices.Chunk(updatedRows, 2000) {
		_, err = tx.NamedExec(`
			UPDATE blocks_deposit_requests_v2
			SET
				slot_queued = :slot_queued,
				index_queued = :index_queued,
				block_queued_root = :block_queued_root,
				status = :status
			WHERE
				id = :id
		`, chunk)
		if err != nil {
			return 0, errors.Wrap(err, "error inserting new rows into blocks_deposit_requests table")
		}
	}
	return len(updatedRows), nil
}
