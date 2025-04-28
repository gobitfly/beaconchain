package cl_transformers

import (
	"slices"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
)

type EventTransformer interface {
	// - filter events for the relevant type
	//
	// - sort by slot, event_index
	//
	// - convert raw data from events into actual structs
	//
	// - load required pre-state from db into memory
	//
	// - update in-memory state with events
	//
	// - apply state to db again
	Transform(tx *sqlx.Tx, events []types.ConsensusLayerEvent) (int, error)
}

func FilterEventsByTypeAndSort(events []types.ConsensusLayerEvent, eventName types.ConsensusLayerEventName) ([]types.ConsensusLayerEvent, error) {
	res := make([]types.ConsensusLayerEvent, 0)
	for _, event := range events {
		if event.EventName != eventName {
			continue
		}
		res = append(res, event)
	}
	// sort by slot, event_index
	slices.SortFunc(res, func(a, b types.ConsensusLayerEvent) int {
		if a.Slot == b.Slot {
			return int(a.EventIndex) - int(b.EventIndex)
		}
		return int(a.Slot) - int(b.Slot)
	})
	return res, nil
}
