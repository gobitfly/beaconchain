package main

import (
	"flag"
	"fmt"
	"perftesting/benchmarks"
	"perftesting/db"
	"perftesting/seeding"
	"perftesting/seeding_strat_valepoch"
	"time"
)

var CONF = GlobalConf{
	// ---- Benchmark specific ---
	BenchmarkDuration:    20 * time.Minute, // duration of benchmark
	BenchUseLatestEpochs: false,            // -- if true it plays with non random epochs to see effects of caching
	BenchLatestEpoch:     215,              // -- head epoch if BenchUseLatestEpochs is true
	BenchEpochDepth:      6,                // how many epochs of data to return per vali

	// ---- Seeder specific ---
	SeederValidatorsInDB: 1000000, // 1m valis
	SeederEpochsInDB:     1,       // 1 day of data (or 225 days of aggregate data)
	SeederNoOfPartitions: 70,      // unused right now
}

func main() {
	var dsn, tableName, cmd string

	flag.StringVar(&tableName, "table.name", "test_ss", "name of table to create")
	flag.StringVar(&cmd, "cmd", "bench", "bench or seed")
	flag.StringVar(&dsn, "db.dsn", "postgres://user:pass@host:port/dbnames", "data-source-name of db, if it starts with projects/ it will use gcp-secretmanager")
	flag.Parse()

	err := db.InitWithDSN(dsn)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Table Name: %v\n", tableName)

	switch cmd {
	case "bench":
		err = RunBenchmark(tableName)
	case "seed":
		//seeder := seeding.GetUnpartitioned(tableName)
		seeder := seeding_strat_valepoch.GetSeederPartitionEpoch(tableName, CONF.SeederNoOfPartitions, true)
		//seeder := seeding.GetSeederPartitionHashIndex(tableName, 64, true)

		//seeder := seeding.GetSeederPartitionExotic(tableName, 30, 6, true)
		//seeder := seeding.GetSeederPartitionExoticReverse(tableName, 30, 6, true)
		err = RunSeeder(tableName, seeder)
	default:
		panic("unknown command")
	}

	if err != nil {
		panic(err)
	}

	fmt.Printf("done")
}

func RunSeeder(tableName string, seeder *seeding.Seeder) error {
	fmt.Printf("Running seeder\n")

	seeder.EpochsInDB = CONF.SeederEpochsInDB
	seeder.ValidatorsInDB = CONF.SeederValidatorsInDB

	fmt.Printf("creating schema\n")
	err := seeder.CreateSchema()
	if err != nil {
		return err
	}

	fmt.Printf("seeding\n")
	err = seeder.FillTable()
	if err != nil {
		return err
	}
	return nil
}

func RunBenchmark(tableName string) error {
	var validatorsInDB, epochsInDB int

	err := db.DB.Get(&validatorsInDB, "SELECT max(validatorindex) FROM "+tableName+" WHERE epoch = 0")
	if err != nil {
		return err
	}
	err = db.DB.Get(&epochsInDB, "SELECT max(epoch) FROM "+tableName)
	if err != nil {
		return err
	}

	fmt.Printf("Running benchmark %v \n", CONF.BenchmarkDuration.String())
	benchmark := benchmarks.Benchmarker{
		TableName:       tableName,
		ValidatorsInDB:  validatorsInDB,
		EpochsInDB:      epochsInDB,
		UseLatestEpochs: CONF.BenchUseLatestEpochs,
		LatestEpoch:     CONF.BenchLatestEpoch,
		EpochDepth:      CONF.BenchEpochDepth,
	}

	//benchmark.RunBenchmarkParallel(CONF.BenchmarkDuration)
	benchmark.RunBenchmarkDBKiller(CONF.BenchmarkDuration)
	//benchmark.RunBenchmarkSequential(CONF.BenchmarkDuration / 8)
	return nil
}

type GlobalConf struct {
	SeederValidatorsInDB int
	SeederEpochsInDB     int
	BenchmarkDuration    time.Duration
	BenchUseLatestEpochs bool
	BenchLatestEpoch     int
	SeederNoOfPartitions int
	BenchEpochDepth      int
}
