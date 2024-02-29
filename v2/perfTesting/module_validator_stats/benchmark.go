package module_validator_stats

import (
	"bytes"
	"fmt"
	"math/rand"
	"perftesting/benchmarks"
	"perftesting/db"
	"time"

	"github.com/shopspring/decimal"
)

type Data struct {
	ValidatorsInDB  int
	EpochsInDB      int
	UseLatestEpochs bool // when true it will not randomly select epochs but use the latest epoch as base for the requests
	LatestEpoch     int
	EpochDepth      int
}

type ValEpochParallelBenchmark struct {
	Data
}
type ValEpochSequentialBenchmark struct {
	Data
}
type ValEpochDBKillerBenchmark struct {
	Data
}

func (conf *ValEpochParallelBenchmark) RunBenchmark(b benchmarks.Benchmarker) {
	runContext := b.GetContext()

	// 10/s
	// TODO change to 100ms
	runContext.RunSingle("10 Validators", 200*time.Millisecond, func() { RunRandomValis(b, 10, conf.Data) })
	runContext.RunSingle("100 Validators", 200*time.Millisecond, func() { RunRandomValis(b, 100, conf.Data) })
	runContext.RunSingle("1000 Validators", 200*time.Millisecond, func() { RunRandomValis(b, 1000, conf.Data) }) // 100ms

	// 5/s
	runContext.RunSingle("10.000 Validators", 200*time.Millisecond, func() { RunRandomValis(b, 10000, conf.Data) })

	if conf.ValidatorsInDB > 100000 {
		// 1/s
		runContext.RunSingle("100.000 Validators", 1*time.Second, func() { RunRandomValis(b, 100000, conf.Data) })

		// 0.5/s
		runContext.RunSingle("200.000 Validators", 2*time.Second, func() { RunRandomValis(b, 200000, conf.Data) })
	} else {
		fmt.Println("!! Skipping 100.000 Validators")
		fmt.Println("!! Skipping 200.000 Validators")
	}

	// 1/10m
	runContext.RunSingle("ExporterAggr 6 Epochs", 10*time.Minute, func() { RunGetAllForExport(b, 6, conf.Data) })

	runContext.RunSingle("ExporterAggr 31 Epochs", 10*time.Minute, func() { RunGetAllForExport(b, 31, conf.Data) })

	runContext.Wg.Wait()
}

func (conf *ValEpochSequentialBenchmark) RunBenchmark(b benchmarks.Benchmarker) {
	b.GetContext().RunSingle("10 Validators", 10*time.Millisecond, func() { RunRandomValis(b, 10, conf.Data) }).Wg.Wait()
	b.GetContext().RunSingle("100 Validators", 10*time.Millisecond, func() { RunRandomValis(b, 100, conf.Data) }).Wg.Wait()
	b.GetContext().RunSingle("1000 Validators", 10*time.Millisecond, func() { RunRandomValis(b, 1000, conf.Data) }).Wg.Wait() // 100ms
	b.GetContext().RunSingle("10.000 Validators", 10*time.Millisecond, func() { RunRandomValis(b, 10000, conf.Data) }).Wg.Wait()

	if conf.ValidatorsInDB > 100000 {
		b.GetContext().RunSingle("100.000 Validators", 10*time.Millisecond, func() { RunRandomValis(b, 100000, conf.Data) }).Wg.Wait()
		b.GetContext().RunSingle("200.000 Validators", 10*time.Millisecond, func() { RunRandomValis(b, 200000, conf.Data) }).Wg.Wait()
	} else {
		fmt.Println("!! Skipping 100.000 Validators")
		fmt.Println("!! Skipping 200.000 Validators")
	}

	b.GetContext().RunSingle("ExporterAggr 6 Epochs", 10*time.Millisecond, func() { RunGetAllForExport(b, 6, conf.Data) }).Wg.Wait()
	b.GetContext().RunSingle("ExporterAggr 31 Epochs", 10*time.Millisecond, func() { RunGetAllForExport(b, 31, conf.Data) }).Wg.Wait()
}

func (conf *ValEpochDBKillerBenchmark) RunBenchmark(b benchmarks.Benchmarker) {
	runContext := b.GetContext()

	// 10/s
	// TODO change to 100ms
	for i := 0; i < 10; i++ {
		runContext.RunSingle("10 Validators", 100*time.Millisecond, func() { RunRandomValis(b, 10, conf.Data) })
		runContext.RunSingle("100 Validators", 100*time.Millisecond, func() { RunRandomValis(b, 100, conf.Data) })
		runContext.RunSingle("1000 Validators", 100*time.Millisecond, func() { RunRandomValis(b, 1000, conf.Data) }) // 100ms

		// 5/s
		runContext.RunSingle("10.000 Validators", 100*time.Millisecond, func() { RunRandomValis(b, 10000, conf.Data) })

		if conf.ValidatorsInDB > 100000 {
			// 1/s
			runContext.RunSingle("100.000 Validators", 100*time.Millisecond, func() { RunRandomValis(b, 100000, conf.Data) })

			// 0.5/s
			runContext.RunSingle("200.000 Validators", 100*time.Millisecond, func() { RunRandomValis(b, 200000, conf.Data) })
		} else {
			fmt.Println("!! Skipping 100.000 Validators")
			fmt.Println("!! Skipping 200.000 Validators")
		}
	}

	// 1/10m
	runContext.RunSingle("ExporterAggr 6 Epochs", 5*time.Minute, func() { RunGetAllForExport(b, 6, conf.Data) })

	runContext.RunSingle("ExporterAggr 31 Epochs", 5*time.Minute, func() { RunGetAllForExport(b, 31, conf.Data) })

	runContext.Wg.Wait()
}

// Export

func RunGetAllForExport(b benchmarks.Benchmarker, epochs int, conf Data) {
	var data []RandomValisResponse

	var randomStartEpoch int
	if conf.UseLatestEpochs {
		randomStartEpoch = conf.LatestEpoch - epochs + 1
	} else {
		randomStartEpoch = rand.Intn(conf.EpochsInDB + 1 - epochs)
	}

	randomEndEpoch := randomStartEpoch + epochs

	err := db.DB.Select(&data, fmt.Sprintf(`
		SELECT
			sum(attestations_source_reward+attestations_target_reward+attestations_head_reward+blocks_cl_reward) as rewards
		FROM %s
		WHERE epoch BETWEEN %d AND %d`,
		b.TableName, randomStartEpoch, randomEndEpoch))

	if err != nil {
		panic(err)
	}

}

// Validator

type RandomValisResponse struct {
	Epoch     int             `db:"epoch"`
	Validator int             `db:"validatorindex"`
	Rewards   decimal.Decimal `db:"rewards"`
}

func RunRandomValis(b benchmarks.Benchmarker, validatorAmount int, conf Data) {
	var data []RandomValisResponse

	var randomStartEpoch int
	if conf.UseLatestEpochs {
		randomStartEpoch = conf.LatestEpoch - conf.EpochDepth + 1
	} else {
		randomStartEpoch = rand.Intn(conf.EpochsInDB + 1 - conf.EpochDepth)
	}

	randomEndEpoch := randomStartEpoch + conf.EpochDepth - 1

	epoch := fmt.Sprintf(` AND epoch BETWEEN %d AND %d`, randomStartEpoch, randomEndEpoch)
	if conf.EpochDepth == 0 {
		epoch = ""
	}

	query := fmt.Sprintf(`
		SELECT
			sum(attestations_source_reward+attestations_target_reward+attestations_head_reward+blocks_cl_reward) as rewards
		FROM %s
		WHERE validatorindex IN (
			%s
		) %s
	`, b.TableName, createRandomSeries(validatorAmount, conf.ValidatorsInDB-1), epoch)

	err := db.DB.Select(&data, query)
	if err != nil {
		panic(err)
	}

	if len(data) != 1 { // validatorAmount*epochs
		panic(fmt.Sprintf("Expected %d rows, got %d", validatorAmount*conf.EpochDepth, len(data)))
	}
}

func createRandomSeries(amount, max int) string {
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
