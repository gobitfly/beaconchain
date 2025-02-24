package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

func (rp *RocketpoolExporter) TagValidators() error {
	if len(rp.MinipoolsByAddress) == 0 {
		return nil
	}

	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(timeStart)}, "saved rocketpool-validator-tags")
	}(timeStart)

	data := rp.prepareValidatorTagData()

	return rp.saveValidatorTags(data)
}

func (rp *RocketpoolExporter) prepareValidatorTagData() []*RocketpoolMinipool {
	data := make([]*RocketpoolMinipool, len(rp.MinipoolsByAddress))
	i := 0
	for _, mp := range rp.MinipoolsByAddress {
		data[i] = mp
		i++
	}
	return data
}

func (rp *RocketpoolExporter) saveValidatorTags(data []*RocketpoolMinipool) error {
	batchSize := 5000

	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings, valueArgs := rp.prepareValidatorTagBatch(data[start:end])
		if err := db.SaveValidatorTags(valueStrings, valueArgs); err != nil {
			return fmt.Errorf("error inserting into validator_tags: %w", err)
		}

		if err := db.SaveValidatorPool(valueStrings, valueArgs); err != nil {
			return fmt.Errorf("error inserting into validator_pool: %w", err)
		}
	}
	return nil
}

func (rp *RocketpoolExporter) prepareValidatorTagBatch(data []*RocketpoolMinipool) ([]string, []interface{}) {
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data))

	for i, d := range data {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, 'rocketpool')", i+1))
		valueArgs = append(valueArgs, d.Pubkey)
	}

	return valueStrings, valueArgs
}
