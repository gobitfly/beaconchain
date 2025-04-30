package cl_transformers

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"encoding/json"
)

type LegacySwitchToCompoundingEventTransformer struct{}

func (d *LegacySwitchToCompoundingEventTransformer) Transform(tx *sqlx.Tx, events []types.ConsensusLayerEvent) (int, error) {
	eventName := types.SwitchToCompoundingEventName
	// filter events for the relevant type
	events, err := FilterEventsByTypeAndSort(events, eventName)
	if err != nil {
		return 0, errors.Wrap(err, "error filtering events")
	}
	if len(events) == 0 {
		return 0, nil
	}
	// convert raw data from events into actual structs
	data := make([]types.SwitchToCompoundingEvent, len(events))
	for i, event := range events {
		// unmarshal from data field to DepositQueuedEvent using json
		err = json.Unmarshal(event.RawData, &data[i])
		if err != nil {
			return 0, errors.Wrap(err, "error unmarshalling event data")
		}
	}

	filters := make([]types.ConsensusLayerEventFilter, 0)
	for _, event := range events {
		filters = append(filters, types.ConsensusLayerEventFilter{
			Slot:      uint64(event.Slot),
			BlockRoot: event.BlockRoot,
		})
	}

	var orConditions []goqu.Expression
	for _, filter := range filters {
		orConditions = append(orConditions, goqu.Ex{
			"slot":       filter.Slot,
			"block_root": filter.BlockRoot,
		})
	}

	ds := goqu.Dialect("postgres").Insert("blocks_switch_to_compounding_requests").
		Cols(
			"block_slot",
			"block_root",
			"request_index",
			"address",
			"validator_index").
		FromQuery(
			goqu.From("consensus_layer_events").As("cle").
				Select(
					"cle.slot",
					"cle.block_root",
					"cle.event_index",
					goqu.L("decode((cle.data->>'address'), 'base64') AS address"),
					goqu.L("vali.validatorindex AS validator_index"),
				).
				Join(goqu.T("validators"), goqu.On(goqu.Ex{
					"vali.pubkey": goqu.L(`decode(cle.data ->> 'pubkey', 'base64')`),
				})).As("vali").
				Where(goqu.And(
					goqu.Ex{"event_name": eventName},
					goqu.Or(orConditions...),
					goqu.Ex{"version": types.ConsensusLayerEventVersion},
				)),
		)
	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return 0, fmt.Errorf("error preparing query: %w", err)
	}

	res, err := tx.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("error executing query: %w", err)
	}
	c, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting rows affected: %w", err)
	}

	return int(c), nil
}
