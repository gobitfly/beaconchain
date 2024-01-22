package benchmarks

import (
	"bytes"
	"fmt"
	"math/rand"
	"perftesting/db"

	"github.com/shopspring/decimal"
)

type RandomValisResponse struct {
	Epoch     int             `db:"epoch"`
	Validator int             `db:"validatorindex"`
	Rewards   decimal.Decimal `db:"rewards"`
}

func (b *Benchmarker) RunRandomValis(validatorAmount int, epochs int) {
	var data []RandomValisResponse

	var randomStartEpoch int
	if b.UseLatestEpochs {
		randomStartEpoch = b.LatestEpoch - epochs + 1
	} else {
		randomStartEpoch = rand.Intn(b.EpochsInDB + 1 - epochs)
	}

	randomEndEpoch := randomStartEpoch + epochs - 1

	epoch := fmt.Sprintf(` AND epoch BETWEEN %d AND %d`, randomStartEpoch, randomEndEpoch)
	if epochs == 0 {
		epoch = ""
	}

	query := fmt.Sprintf(`
		SELECT 
			epoch, 
			validatorindex, 
			sum(attestations_source_reward+attestations_target_reward+attestations_head_reward+blocks_cl_reward) as rewards 
		FROM %s
		WHERE validatorindex IN (
			%s
		) %s
		GROUP BY epoch, validatorindex
	`, b.TableName, createRandomSeries(validatorAmount, b.ValidatorsInDB-1), epoch)

	err := db.DB.Select(&data, query)
	if err != nil {
		panic(err)
	}

	if len(data) != validatorAmount*epochs {
		panic(fmt.Sprintf("Expected %d rows, got %d", validatorAmount*epochs, len(data)))
	}
}

func createRandomSeries(amount, max int) string {

	return createRandomSeriesBig(amount, max)

	// erg := ""
	// randoms := map[int]bool{}
	// count := 0
	// for {
	// 	rand := rand.Intn(max)
	// 	if _, ok := randoms[rand]; !ok {
	// 		randoms[rand] = true
	// 		erg += fmt.Sprintf("%d,", rand)
	// 		count++
	// 	}
	// 	if count == amount {
	// 		break
	// 	}
	// }
	// return erg[:len(erg)-1]
}

func createRandomSeriesBig(amount, max int) string {
	var buffer bytes.Buffer
	start := rand.Intn(max)
	count := 0
	rrange := int(float64(max)/float64(amount)) - 1
	for {
		rand := (start + 1 + rand.Intn(rrange-1)) % max
		start = rand
		buffer.WriteString(fmt.Sprintf("%d,", rand))
		count++

		if count == amount {
			break
		}
	}

	erg := buffer.String()
	return erg[:len(erg)-1]
}
