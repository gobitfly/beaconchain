package benchmarks

import (
	"fmt"
	"math/rand"
	"perftesting/db"
)

func (b *Benchmarker) RunGetAllForExport(epochs int) {
	var data []RandomValisResponse

	var randomStartEpoch int
	if b.UseLatestEpochs {
		randomStartEpoch = b.LatestEpoch - epochs + 1
	} else {
		randomStartEpoch = rand.Intn(b.EpochsInDB + 1 - epochs)
	}

	randomEndEpoch := randomStartEpoch + epochs

	err := db.DB.Select(&data, fmt.Sprintf(`
		SELECT 
			epoch, 
			validatorindex, 
			sum(attestations_source_reward+attestations_target_reward+attestations_head_reward+blocks_cl_reward) as rewards 
		FROM %s
		WHERE epoch BETWEEN %d AND %d
		GROUP BY epoch, validatorindex`,
		b.TableName, randomStartEpoch, randomEndEpoch))

	if err != nil {
		panic(err)
	}

	// if len(data) != b.ValidatorsInDB*epochs {
	// 	panic(fmt.Sprintf("Expected %d rows, got %d", b.ValidatorsInDB*epochs, len(data)))
	// }
}
