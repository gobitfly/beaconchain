package db

// TODO - This file have to be reviewed and refactored

import (
	"context"
	"crypto/tls"
	"fmt"
	"reflect"
	"strconv"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"golang.org/x/sync/errgroup"
)

var ClickHouseNativeWriter ch.Conn

func MustInitClickhouseNative(writer *types.DatabaseConfig) ch.Conn {
	if writer.MaxOpenConns == 0 {
		writer.MaxOpenConns = 50
	}
	if writer.MaxIdleConns == 0 {
		writer.MaxIdleConns = 10
	}
	if writer.MaxOpenConns < writer.MaxIdleConns {
		writer.MaxIdleConns = writer.MaxOpenConns
	}
	log.Infof("initializing clickhouse native writer db connection to %v:%v/%v with %v/%v conn limit", writer.Host, writer.Port, writer.Name, writer.MaxIdleConns, writer.MaxOpenConns)
	dbWriter, err := ch.Open(&ch.Options{
		MaxOpenConns: writer.MaxOpenConns,
		MaxIdleConns: writer.MaxIdleConns,
		// ConnMaxLifetime: time.Minute,
		// the following lowers traffic between client and server
		Compression: &ch.Compression{
			Method: ch.CompressionLZ4,
		},
		Addr: []string{fmt.Sprintf("%s:%s", writer.Host, writer.Port)},
		Auth: ch.Auth{
			Username: writer.Username,
			Password: writer.Password,
			Database: writer.Name,
		},
		Debug: false,
		TLS:   &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12},
		// this gets only called when debug is true
		Debugf: func(s string, p ...interface{}) {
			log.Debugf("CH NATIVE WRITER: "+s, p...)
		},
		Settings: ch.Settings{
			"parallel_view_processing": "true",
			//"send_logs_level":          "trace",
			"max_insert_threads": "4",
			"deduplicate_blocks_in_dependent_materialized_views":                "1",
			"update_insert_deduplication_token_in_dependent_materialized_views": "1",
		},
	})
	if err != nil {
		log.Fatal(err, "Error connecting to clickhouse native writer", 0)
	}
	// verify connection
	ClickHouseTestConnection(&dbWriter, writer.Name)

	return dbWriter
}

func ClickHouseTestConnection(db *ch.Conn, dataBaseName string) {
	v, err := (*db).ServerVersion()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to ping clickhouse database %s: %w", dataBaseName, err), "", 0)
	}
	log.Debugf("connected to clickhouse database %s with version %s", dataBaseName, v)
}

type UltraFastClickhouseStruct interface {
	Get(string) any
	Extend(UltraFastClickhouseStruct) error
}

// RewardBreakdownHyperMap struct with fixed responses
type RewardBreakdownHyperMap struct {
	Source     int64
	Target     int64
	Head       int64
	Inactivity int64
	Inclusion  int64
}

// Get method with fixed response
func (f *RewardBreakdownHyperMap) Get(key any) (any, bool) {
	switch key {
	case "source":
		return f.Source, true
	case "target":
		return f.Target, true
	case "head":
		return f.Head, true
	case "inactivity":
		return f.Inactivity, true
	case "inclusion":
		return f.Inclusion, true
	default:
		return nil, false
	}
}

// Put method that does nothing
func (f *RewardBreakdownHyperMap) Put(key any, value any) {
	// Do nothing
}

// Keys method that returns a channel with the keys
func (f *RewardBreakdownHyperMap) Keys() <-chan any {
	ch := make(chan any)
	go func() {
		ch <- "source"
		ch <- "target"
		ch <- "head"
		ch <- "inactivity"
		ch <- "inclusion"
		close(ch)
	}()
	return ch
}

type ExecutionBreakdownHyperMap struct {
	Source int64
	Target int64
	Head   int64
}

func (f *ExecutionBreakdownHyperMap) Get(key any) (any, bool) {
	switch key {
	case "source":
		return f.Source, true
	case "target":
		return f.Target, true
	case "head":
		return f.Head, true
	default:
		return nil, false
	}
}

func (f *ExecutionBreakdownHyperMap) Put(key any, value any) {
	// Do nothing
}

func (f *ExecutionBreakdownHyperMap) Keys() <-chan any {
	ch := make(chan any)
	go func() {
		ch <- "source"
		ch <- "target"
		ch <- "head"
		close(ch)
	}()
	return ch
}

type VDBDataEpochColumns struct {
	// this should be the same as Epoch but only once, basically
	EpochsContained                     []uint64 `custom_size:"1"`
	ValidatorIndex                      []uint64
	Epoch                               []int64
	EpochTimestamp                      []*time.Time
	BalanceEffectiveStart               []int64
	BalanceEffectiveEnd                 []int64
	BalanceStart                        []int64
	BalanceEnd                          []int64
	DepositsCount                       []int64
	DepositsAmount                      []int64
	WithdrawalsCount                    []int64
	WithdrawalsAmount                   []int64
	AttestationsScheduled               []int64
	AttestationsObserved                []int64
	AttestationsHeadMatched             []int64
	AttestationsSourceMatched           []int64
	AttestationsTargetMatched           []int64
	AttestationsHeadExecuted            []int64
	AttestationsSourceExecuted          []int64
	AttestationsTargetExecuted          []int64
	AttestationsHeadReward              []int64
	AttestationsSourceReward            []int64
	AttestationsTargetReward            []int64
	AttestationsInactivityReward        []int64
	AttestationsInclusionReward         []int64
	AttestationsIdealHeadReward         []int64
	AttestationsIdealSourceReward       []int64
	AttestationsIdealTargetReward       []int64
	AttestationsIdealInactivityReward   []int64
	AttestationsIdealInclusionReward    []int64
	AttestationsLocalizedMaxReward      []int64
	AttestationsHyperLocalizedMaxReward []int64
	InclusionDelaySum                   []int64
	OptimalInclusionDelaySum            []int64
	BlocksStatusSlot                    [][]int64
	BlocksStatusProposed                [][]bool
	BlockRewardsSlot                    [][]int64
	BlockRewardsAttestationsReward      [][]int64
	BlockRewardsSyncAggregateReward     [][]int64
	BlockRewardsSlasherReward           [][]int64
	BlocksClMissedMedianReward          []int64
	BlocksSlashingCount                 []int64
	BlocksExpected                      []float64
	SyncScheduled                       []int64
	SyncStatusSlot                      [][]int64
	SyncStatusExecuted                  [][]bool
	SyncRewardsSlot                     [][]int64
	SyncRewardsReward                   [][]int64
	SyncLocalizedMaxReward              []int64
	SyncCommitteesExpected              []float64
	Slashed                             []bool
	AttestationAssignmentsSlot          [][]int64
	AttestationAssignmentsCommittee     [][]int64
	AttestationAssignmentsIndex         [][]int64
	SyncCommitteeAssignmentsPeriod      [][]int64
	SyncCommitteeAssignmentsIndex       [][]int64
}

type OrderedMap interface {
	Get(key any) (any, bool)
	Put(key any, value any)
	Keys() <-chan any
}

// get by string
func (c *VDBDataEpochColumns) Get(str string) any {
	// test type assertion
	switch str {
	case "validator_index":
		return c.ValidatorIndex
	case "epoch":
		return c.Epoch
	case "epoch_timestamp":
		return c.EpochTimestamp
	case "balance_effective_start":
		return c.BalanceEffectiveStart
	case "balance_effective_end":
		return c.BalanceEffectiveEnd
	case "balance_start":
		return c.BalanceStart
	case "balance_end":
		return c.BalanceEnd
	case "deposits_count":
		return c.DepositsCount
	case "deposits_amount":
		return c.DepositsAmount
	case "withdrawals_count":
		return c.WithdrawalsCount
	case "withdrawals_amount":
		return c.WithdrawalsAmount
	case "attestations_scheduled":
		return c.AttestationsScheduled
	case "attestations_observed":
		return c.AttestationsObserved
	case "attestations_head_executed":
		return c.AttestationsHeadExecuted
	case "attestations_source_executed":
		return c.AttestationsSourceExecuted
	case "attestations_target_executed":
		return c.AttestationsTargetExecuted
	case "attestations_head_matched":
		return c.AttestationsHeadMatched
	case "attestations_source_matched":
		return c.AttestationsSourceMatched
	case "attestations_target_matched":
		return c.AttestationsTargetMatched
	case "attestations_head_reward":
		return c.AttestationsHeadReward
	case "attestations_source_reward":
		return c.AttestationsSourceReward
	case "attestations_target_reward":
		return c.AttestationsTargetReward
	case "attestations_inactivity_reward":
		return c.AttestationsInactivityReward
	case "attestations_inclusion_reward":
		return c.AttestationsInclusionReward
	case "attestations_ideal_head_reward":
		return c.AttestationsIdealHeadReward
	case "attestations_ideal_source_reward":
		return c.AttestationsIdealSourceReward
	case "attestations_ideal_target_reward":
		return c.AttestationsIdealTargetReward
	case "attestations_ideal_inactivity_reward":
		return c.AttestationsIdealInactivityReward
	case "attestations_ideal_inclusion_reward":
		return c.AttestationsIdealInclusionReward
	case "attestations_localized_max_reward":
		return c.AttestationsLocalizedMaxReward
	case "attestations_hyperlocalized_max_reward":
		return c.AttestationsHyperLocalizedMaxReward
	case "inclusion_delay_sum":
		return c.InclusionDelaySum
	case "optimal_inclusion_delay_sum":
		return c.OptimalInclusionDelaySum
	case "blocks_status.slot":
		return c.BlocksStatusSlot
	case "blocks_status.proposed":
		return c.BlocksStatusProposed
	case "block_rewards.slot":
		return c.BlockRewardsSlot
	case "block_rewards.attestations_reward":
		return c.BlockRewardsAttestationsReward
	case "block_rewards.sync_aggregate_reward":
		return c.BlockRewardsSyncAggregateReward
	case "block_rewards.slasher_reward":
		return c.BlockRewardsSlasherReward
	case "blocks_cl_missed_median_reward":
		return c.BlocksClMissedMedianReward
	case "blocks_slashing_count":
		return c.BlocksSlashingCount
	case "blocks_expected":
		return c.BlocksExpected
	case "sync_scheduled":
		return c.SyncScheduled
	case "sync_status.slot":
		return c.SyncStatusSlot
	case "sync_status.executed":
		return c.SyncStatusExecuted
	case "sync_rewards.slot":
		return c.SyncRewardsSlot
	case "sync_rewards.reward":
		return c.SyncRewardsReward
	case "sync_localized_max_reward":
		return c.SyncLocalizedMaxReward
	case "sync_committees_expected":
		return c.SyncCommitteesExpected
	case "slashed":
		return c.Slashed
	case "attestation_assignments.slot":
		return c.AttestationAssignmentsSlot
	case "attestation_assignments.committee":
		return c.AttestationAssignmentsCommittee
	case "attestation_assignments.index":
		return c.AttestationAssignmentsIndex
	case "sync_committee_assignments.period":
		return c.SyncCommitteeAssignmentsPeriod
	case "sync_committee_assignments.index":
		return c.SyncCommitteeAssignmentsIndex
	default:
		return nil
	}
}

// factory that pre allocates slices with a given capacity
func NewVDBDataEpochColumns(capacity int) (VDBDataEpochColumns, error) {
	start := time.Now()
	c := VDBDataEpochColumns{}
	ct := reflect.TypeOf(c)
	cv := reflect.ValueOf(&c).Elem()
	var g errgroup.Group

	for i := 0; i < ct.NumField(); i++ {
		i := i // capture loop variable
		g.Go(func() error {
			f := cv.Field(i)
			// check for custom_size tag
			if tag := ct.Field(i).Tag.Get("custom_size"); tag != "" {
				cap, err := strconv.Atoi(tag)
				if err != nil {
					return fmt.Errorf("failed to parse custom_size tag: %w", err)
				}
				f.Set(reflect.MakeSlice(f.Type(), cap, cap))
			} else {
				f.Set(reflect.MakeSlice(f.Type(), capacity, capacity))
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return c, err
	}

	log.Debugf("allocated in %s", time.Since(start))
	return c, nil
}

// util function to combine two VDBDataEpochColumns
func (c *VDBDataEpochColumns) Extend(cOther UltraFastClickhouseStruct) error {
	// ype assert
	c2, ok := (cOther).(*VDBDataEpochColumns)
	if !ok {
		return fmt.Errorf("type assertion failed")
	}
	// use reflection baby
	start := time.Now()
	ct := reflect.TypeOf(*c)
	cv := reflect.ValueOf(c).Elem()
	for i := 0; i < ct.NumField(); i++ {
		f := cv.Field(i)
		f2 := reflect.ValueOf(c2).Elem().Field(i)
		// append
		cv.Field(i).Set(reflect.AppendSlice(f, f2))
	}
	log.Debugf("extended in %s", time.Since(start))
	return nil
}

func UltraFastDumpToClickhouse[T UltraFastClickhouseStruct](data T, target_table string, insert_uuid string) error {
	start := time.Now()
	// get column order & names from clickhouse
	var columns []string
	err := ClickHouseReader.Select(&columns, "SELECT name FROM system.columns where table=$1 and database=currentDatabase() order by position;", target_table)
	if err != nil {
		return err
	}
	log.Debugf("got columns in %s", time.Since(start))
	start = time.Now()
	//query_uuid, _ := uuid.NewV6()
	// log query id
	//log.Debugf("query id: %s", query_uuid.String())
	// prepare batch
	abortCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	ctx := ch.Context(abortCtx, ch.WithSettings(ch.Settings{
		"insert_deduplication_token": insert_uuid, // 重复数据插入时，会根据这个字段进行去重
		"insert_deduplicate":         false,
		//"query_profiler_cpu_time_period_ns":  "1000000",
		//"memory_profiler_sample_probability": "0.1",
	}), ch.WithLogs(func(l *ch.Log) {
		log.Debugf("CH NATIVE WRITER: %s", l.Text)
	}),
	// ch.WithQueryID(query_uuid.String())
	)
	batch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `INSERT INTO `+target_table)
	if err != nil {
		return err
	}
	log.Debugf("prepared batch in %s", time.Since(start))
	start = time.Now()
	defer func() {
		if batch.IsSent() {
			return
		}
		err := batch.Abort()
		if err != nil {
			log.Warnf("failed to abort batch: %v", err)
		}
	}()
	var g errgroup.Group
	g.SetLimit(5)

	// iterate columns retrieved from clickhouse
	for i, n := range columns {
		// Capture the loop variable
		col_index := i
		col_name := n
		if col_name == "_inserted_at" {
			continue
		}
		// Start a new goroutine for each column
		g.Go(func() error {
			// get it from the struct
			column := data.Get(col_name)
			if column == nil {
				return fmt.Errorf("column %s not found in struct", col_name)
			}
			// Perform the type assertion and append operation
			err = batch.Column(col_index).Append(column)
			log.Debugf("appended column %s in %s", col_name, time.Since(start))
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	log.Debugf("appended all columns to batch in %s", time.Since(start))
	start = time.Now()
	err = batch.Send()
	if err != nil {
		return err
	}
	log.Debugf("sent batch in %s", time.Since(start))
	return nil
}
