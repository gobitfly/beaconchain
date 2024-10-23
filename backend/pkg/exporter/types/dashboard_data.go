package types

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type VDBDataEpochColumns struct {
	// this should be the same as Epoch but only once, basically
	EpochsContained                     []uint64    `custom_size:"1"`
	InsertBatchID                       []uuid.UUID `custom_size:"1"`
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
func (c *VDBDataEpochColumns) Extend(cOther db.UltraFastClickhouseStruct) error {
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
