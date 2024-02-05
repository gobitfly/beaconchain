package module_validator_stats

import (
	"fmt"
	"math/rand"
	"perftesting/db"
	"perftesting/seeding"
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type SeederData struct {
	ValidatorsInDB int
	EpochsInDB     int
}

func getValiEpochSeeder(tableName string, columnarEngine bool, scheme seeding.SeederScheme, filler seeding.SeederFiller) *seeding.Seeder {
	return seeding.GetSeeder(tableName, columnarEngine, scheme, filler)
}

func generateRandomEntries(count int, validator int) []Entry {
	entries := make([]Entry, count)
	slashed := false
	balanceStart := decimal.NewFromFloat(32e9)
	balanceEnd := decimal.NewFromFloat(32e9)
	blocks := 0

	depositCount := rand.Int63()

	for i := 0; i < count; i++ {
		if !slashed {
			slashed = rand.Intn(20000000) == 1
		}
		missed := rand.Intn(100) == 1
		missedModifier := int64(1)
		if missed {
			missedModifier = -1
		}
		blockProposed := rand.Intn(200) == 1
		if blockProposed {
			blocks++
		}

		attReward := missedModifier * rand.Int63n(10000000)

		entry := Entry{
			ValidatorIndex:                    int64(validator),
			Epoch:                             int64(i),
			AttestationsSourceReward:          attReward * 10 / 219 / 10,
			AttestationsTargetReward:          attReward * 10 / 406 / 10,
			AttestationsHeadReward:            attReward * 10 / 219 / 10,
			AttestationsInactivityReward:      missedModifier * rand.Int63n(10000000),
			AttestationsInclusionReward:       attReward, // meh
			AttestationsReward:                attReward,
			AttestationsIdealSourceReward:     missedModifier * rand.Int63n(10000000),
			AttestationsIdealTargetReward:     missedModifier * rand.Int63n(10000000),
			AttestationsIdealHeadReward:       missedModifier * rand.Int63n(10000000),
			AttestationsIdealInactivityReward: missedModifier * rand.Int63n(10000000),
			AttestationsIdealInclusionReward:  missedModifier * rand.Int63n(10000000),
			AttestationsIdealReward:           missedModifier * rand.Int63n(10000000),
			BlocksScheduled:                   0,
			BlocksProposed:                    blocks,
			BlocksClReward:                    missedModifier * (1000000 + rand.Int63n(500000)),
			BlocksElReward:                    float64(missedModifier * rand.Int63n(100000000)),
			SyncScheduled:                     rand.Intn(100),
			SyncExecuted:                      rand.Intn(100),
			SyncRewards:                       rand.Int63(),
			Slashed:                           slashed,
			BalanceStart:                      balanceStart,
			BalanceEnd:                        balanceEnd,
			DepositsCount:                     depositCount,
			DepositsAmount:                    32e9,
			WithdrawalsCount:                  rand.Int63(), // meh
			WithdrawalsAmount:                 rand.Int63(), // meh
		}
		balanceStart.Add(decimal.NewFromFloat(float64(entry.AttestationsReward)))
		balanceEnd.Add(decimal.NewFromFloat(float64(entry.AttestationsReward)))
		entry.BalanceStart = balanceStart
		entry.BalanceEnd = balanceEnd

		entries[i] = entry
	}
	return entries
}

func (data SeederData) FillTable(s *seeding.Seeder) error {
	iterations := data.ValidatorsInDB // valis
	epochs := data.EpochsInDB         // epochs (one day)
	batchSize := s.BatchSize

	timeStart := time.Now()
	defer func() {
		fmt.Printf("Time taken: %v\n", time.Since(timeStart))
	}()

	var entries = make([]Entry, 0, batchSize)
	for i := 0; i < iterations; i++ {
		valEntries := generateRandomEntries(epochs, i)

		entries = append(entries, valEntries...)

		if len(entries) >= batchSize {
			fmt.Printf("Flushing batch to db %v%% - %v/%v\n", float64(int64(float64(i*epochs)*1000/float64(iterations*epochs)))/10, i*epochs, iterations*epochs)
			err := insertEntries(s.TableName, entries)
			if err != nil {
				return err
			}
			entries = make([]Entry, 0, batchSize)
		}
	}
	return nil
}

func insertEntries(tableName string, entries []Entry) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(pq.CopyIn(tableName,
		"validatorindex",
		"epoch",
		"attestations_source_reward",
		"attestations_target_reward",
		"attestations_head_reward",
		"attestations_inactivity_reward",
		"attestations_inclusion_reward",
		"attestations_reward",
		"attestations_ideal_source_reward",
		"attestations_ideal_target_reward",
		"attestations_ideal_head_reward",
		"attestations_ideal_inactivity_reward",
		"attestations_ideal_inclusion_reward",
		"attestations_ideal_reward",
		"blocks_scheduled",
		"blocks_proposed",
		"blocks_cl_reward",
		"blocks_el_reward",
		"sync_scheduled",
		"sync_executed",
		"sync_rewards",
		"slashed",
		"balance_start",
		"balance_end",
		"deposits_count",
		"deposits_amount",
		"withdrawals_count",
		"withdrawals_amount",
	))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, entry := range entries {
		_, err = stmt.Exec(
			entry.ValidatorIndex,
			entry.Epoch,
			entry.AttestationsSourceReward,
			entry.AttestationsTargetReward,
			entry.AttestationsHeadReward,
			entry.AttestationsInactivityReward,
			entry.AttestationsInclusionReward,
			entry.AttestationsReward,
			entry.AttestationsIdealSourceReward,
			entry.AttestationsIdealTargetReward,
			entry.AttestationsIdealHeadReward,
			entry.AttestationsIdealInactivityReward,
			entry.AttestationsIdealInclusionReward,
			entry.AttestationsIdealReward,
			entry.BlocksScheduled,
			entry.BlocksProposed,
			entry.BlocksClReward,
			entry.BlocksElReward,
			entry.SyncScheduled,
			entry.SyncExecuted,
			entry.SyncRewards,
			entry.Slashed,
			entry.BalanceStart,
			entry.BalanceEnd,
			entry.DepositsCount,
			entry.DepositsAmount,
			entry.WithdrawalsCount,
			entry.WithdrawalsAmount,
		)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

type Entry struct {
	ValidatorIndex                    int64
	Epoch                             int64
	AttestationsSourceReward          int64
	AttestationsTargetReward          int64
	AttestationsHeadReward            int64
	AttestationsInactivityReward      int64
	AttestationsInclusionReward       int64
	AttestationsReward                int64
	AttestationsIdealSourceReward     int64
	AttestationsIdealTargetReward     int64
	AttestationsIdealHeadReward       int64
	AttestationsIdealInactivityReward int64
	AttestationsIdealInclusionReward  int64
	AttestationsIdealReward           int64
	BlocksScheduled                   int
	BlocksProposed                    int
	BlocksClReward                    int64
	BlocksElReward                    float64
	SyncScheduled                     int
	SyncExecuted                      int
	SyncRewards                       int64
	Slashed                           bool
	BalanceStart                      decimal.Decimal
	BalanceEnd                        decimal.Decimal
	DepositsCount                     int64
	DepositsAmount                    int64
	WithdrawalsCount                  int64
	WithdrawalsAmount                 int64
}
