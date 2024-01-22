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

	/*
		SELECT
					sum(attestations_source_reward+attestations_target_reward+attestations_head_reward+blocks_cl_reward) as rewards
				FROM %s
				WHERE validatorindex IN (
					select * from create_random_series_big(%d, %d)
				) %s
	*/
	query := fmt.Sprintf(`
		SELECT 
			sum(attestations_source_reward+attestations_target_reward+attestations_head_reward+blocks_cl_reward) as rewards 
		FROM %s
		WHERE validatorindex IN (
			%s
		) %s
	`, b.TableName, createRandomSeries(validatorAmount, b.ValidatorsInDB-1), epoch)

	err := db.DB.Select(&data, query)
	if err != nil {
		panic(err)
	}

	if len(data) != 1 { // validatorAmount*epochs
		panic(fmt.Sprintf("Expected %d rows, got %d", validatorAmount*epochs, len(data)))
	}
}

/*
CREATE OR REPLACE FUNCTION create_random_series_big(amount INT, max INT)
RETURNS SETOF INT AS $$
DECLARE
    start INT := floor(random() * max)::INT;
    count INT := 0;
    rrange INT := floor(max / amount) - 1;
BEGIN
    FOR i IN 1..amount LOOP
        -- Avoid division by zero
        IF rrange <= 1 THEN
            RAISE EXCEPTION 'Invalid input: amount too large for the given max value';
        END IF;

        -- Generate random number and return it
        start := (start + 1 + floor(random() * (rrange - 1)))::INT % max;  -- Update start
        RETURN NEXT start;
        count := count + 1;

        -- Check if enough numbers generated
        IF count = amount THEN
            EXIT;
        END IF;
    END LOOP;

    -- No need for a final RETURN since we use RETURN NEXT within the loop
END;
$$ LANGUAGE plpgsql;

*/

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
