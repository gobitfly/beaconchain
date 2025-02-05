package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/big"
	"slices"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/shopspring/decimal"
)

type MinMaxEpochsResult struct {
	MinEpochStart *uint64 `db:"min_epoch_start"`
	MaxEpochEnd   *uint64 `db:"max_epoch_end"`
}

func (d *DataAccessService) getMinMaxEpochs(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (uint64, uint64, error) {
	clickhouseTable, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting table for period: %w", err)
	}

	ds, err := d.buildMinMaxEpochsQuery(dashboardId, groupId, clickhouseTable)
	if err != nil {
		return 0, 0, fmt.Errorf("error building query: %w", err)
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return 0, 0, fmt.Errorf("error preparing query: %w", err)
	}

	row, err := d.executeMinMaxEpochsQuery(ctx, query, args)
	if err != nil {
		return 0, 0, fmt.Errorf("error executing query: %w", err)
	}

	minEpochStart, maxEpochEnd, err := d.processMinMaxEpochsResult(row)
	if err != nil {
		return 0, 0, fmt.Errorf("error processing result: %w", err)
	}

	return minEpochStart, maxEpochEnd, nil
}

func (d *DataAccessService) buildMinMaxEpochsQuery(dashboardId t.VDBId, groupId int64, clickhouseTable string) (*goqu.SelectDataset, error) {
	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("MIN(epoch_start) as min_epoch_start"),
			goqu.L("MAX(epoch_end) as max_epoch_end")).
		From(goqu.L(fmt.Sprintf(`%s AS r FINAL`, clickhouseTable)))

	if dashboardId.Validators == nil {
		ds = ds.
			With("validators", goqu.L("(SELECT validator_index as validator_index, group_id FROM users_val_dashboards_validators WHERE dashboard_id = ? AND (group_id = ? OR ?::smallint = -1))", dashboardId.Id, groupId, groupId)).
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
			Where(goqu.L("validator_index IN (SELECT validator_index FROM validators)"))
	} else {
		ds = ds.
			Where(goqu.L("validator_index IN ?", dashboardId.Validators))
	}

	return ds, nil
}

func (d *DataAccessService) executeMinMaxEpochsQuery(ctx context.Context, query string, args []interface{}) (MinMaxEpochsResult, error) {
	var row MinMaxEpochsResult
	err := d.clickhouseReader.GetContext(ctx, &row, query, args...)
	if err != nil {
		return MinMaxEpochsResult{}, fmt.Errorf("error executing query: %w", err)
	}
	return row, nil
}

func (d *DataAccessService) processMinMaxEpochsResult(row MinMaxEpochsResult) (uint64, uint64, error) {
	if row.MinEpochStart == nil || row.MaxEpochEnd == nil {
		return 0, 0, fmt.Errorf("no epoch data found")
	}
	return *row.MinEpochStart, *row.MaxEpochEnd, nil
}

func (d *DataAccessService) getMissedELRewards(ctx context.Context, dashboardId t.VDBId, groupId int64, epochStart, epochEnd uint64) (float64, error) {
	query, err := d.buildMissedELRewardsQuery(dashboardId, groupId, epochStart, epochEnd)
	if err != nil {
		return 0, fmt.Errorf("error building query: %w", err)
	}

	totalMissedRewardsEl, err := d.executeMissedELRewardsQuery(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("error executing query: %w", err)
	}

	return totalMissedRewardsEl, nil
}

func (d *DataAccessService) buildMissedELRewardsQuery(dashboardId t.VDBId, groupId int64, epochStart, epochEnd uint64) (*goqu.SelectDataset, error) {
	// Define the `targets` CTE
	targets := goqu.Dialect("postgres").
		From("blocks").
		Select(goqu.I("blocks.slot").As("slot")).
		Where(
			goqu.I("blocks.status").Neq("1"),
			goqu.I("epoch").Gte(epochStart),
			goqu.I("epoch").Lte(epochEnd),
		)

	if dashboardId.Validators == nil {
		targets = targets.
			Join(
				goqu.T("users_val_dashboards_validators").As("uvdv"),
				goqu.On(goqu.I("blocks.proposer").Eq(goqu.I("uvdv.validator_index"))),
			).
			Where(
				goqu.And(
					goqu.I("uvdv.dashboard_id").Eq(dashboardId.Id),
					goqu.Or(
						goqu.I("uvdv.group_id").Eq(groupId),
						goqu.L("?::smallint = -1", groupId),
					),
				),
			)
	} else {
		targets = targets.
			Where(
				goqu.I("blocks.proposer").In(dashboardId.Validators),
			)
	}

	slots := utils.Config.Chain.ClConfig.SlotsPerEpoch / 2

	// Define the `res` CTE
	res := goqu.
		From("targets").
		LeftJoin(
			goqu.T("execution_rewards_finalized").As("b"),
			goqu.On(
				goqu.L(fmt.Sprintf(
					`"b"."slot" >= "targets"."slot" - %d AND "b"."slot" < "targets"."slot" + %d`,
					slots, slots,
				)),
			),
		).
		Select(
			goqu.I("targets.slot"),
			goqu.L("percentile_cont(0.5) WITHIN GROUP (ORDER BY b.value)::numeric(76,0)").As("v"),
		).
		GroupBy(goqu.I("targets.slot"))

	// Build the final query
	query := goqu.From("res").
		With("targets", targets).
		With("res", res).
		Select(goqu.L("COALESCE(SUM(v), 0)"))

	return query, nil
}

func (d *DataAccessService) executeMissedELRewardsQuery(ctx context.Context, query *goqu.SelectDataset) (float64, error) {
	// Generate SQL and arguments
	sql, args, err := query.Prepared(true).ToSQL()
	if err != nil {
		return 0, fmt.Errorf("failed to generate SQL: %w", err)
	}

	// Execute the query
	var totalMissedRewardsEl float64
	err = d.readerDb.GetContext(ctx, &totalMissedRewardsEl, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	return totalMissedRewardsEl, nil
}

type QueryResult struct {
	ValidatorIndex          uint32 `db:"validator_index"`
	EpochStart              uint64 `db:"epoch_start"`
	EpochEnd                uint64 `db:"epoch_end"`
	AttestationReward       int64  `db:"attestations_reward"`
	AttestationsIdealReward int64  `db:"attestations_ideal_reward"`

	AttestationsScheduled         int64 `db:"attestations_scheduled"`
	AttestationsObserved          int64 `db:"attestations_observed"`
	AttestationsHeadExecuted      int64 `db:"attestations_head_executed"`
	AttestationsSourceExecuted    int64 `db:"attestations_source_executed"`
	AttestationsTargetExecuted    int64 `db:"attestations_target_executed"`
	AttestationsRewardRewardsOnly int64 `db:"attestations_reward_rewards_only"`

	BlocksScheduled uint32 `db:"blocks_scheduled"`
	BlocksProposed  uint32 `db:"blocks_proposed"`

	SyncScheduled uint32 `db:"sync_scheduled"`
	SyncExecuted  uint32 `db:"sync_executed"`

	SlashedInPeriod bool   `db:"slashed_in_period"`
	SlashedAmount   uint32 `db:"slashed_amount"`

	BlockChance            float64 `db:"blocks_expected"`
	SyncCommitteesExpected float64 `db:"sync_committees_expected"`

	BlocksCLMissedReward    int64 `db:"blocks_cl_missed_median_reward"`
	SyncLocalizedMaxRewards int64 `db:"sync_localized_max_reward"`
	SyncRewardRewardsOnly   int64 `db:"sync_reward_rewards_only"`

	InclusionDelaySum int64 `db:"inclusion_delay_sum"`
}

func (d *DataAccessService) buildValidatorDashboardGroupSummaryQuery(dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*goqu.SelectDataset, error) {
	// Get the table names based on the period
	clickhouseTable, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, err
	}

	// Build the query
	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("validator_index"),
			goqu.L("epoch_start"),
			goqu.L("epoch_end"),
			goqu.L("attestations_reward"),
			goqu.L("attestations_ideal_reward"),
			goqu.L("attestations_scheduled"),
			goqu.L("attestations_observed"),
			goqu.L("attestations_head_executed"),
			goqu.L("attestations_source_executed"),
			goqu.L("attestations_target_executed"),
			goqu.L("attestations_reward_rewards_only"),
			goqu.L("blocks_scheduled"),
			goqu.L("blocks_proposed"),
			goqu.L("blocks_cl_missed_median_reward"),
			goqu.L("sync_scheduled"),
			goqu.L("sync_executed"),
			goqu.L("slashed AS slashed_in_period"),
			goqu.L("blocks_slashing_count AS slashed_amount"),
			goqu.L("blocks_expected"),
			goqu.L("inclusion_delay_sum"),
			goqu.L("sync_localized_max_reward"),
			goqu.L("sync_reward_rewards_only"),
			goqu.L("sync_committees_expected")).
		From(goqu.L(fmt.Sprintf(`%s AS r FINAL`, clickhouseTable)))

	if dashboardId.Validators == nil {
		ds = ds.
			With("validators", goqu.L("(SELECT validator_index as validator_index, group_id FROM users_val_dashboards_validators WHERE dashboard_id = ? AND (group_id = ? OR ?::smallint = -1))", dashboardId.Id, groupId, groupId)).
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
			Where(goqu.L("validator_index IN (SELECT validator_index FROM validators)"))
	} else {
		ds = ds.
			Where(goqu.L("validator_index IN ?", dashboardId.Validators))
	}

	return ds, nil
}

func (d *DataAccessService) executeValidatorDashboardGroupSummaryQuery(ctx context.Context, query *goqu.SelectDataset) ([]*QueryResult, error) {
	querySQL, args, err := query.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	var rows []*QueryResult
	err = d.clickhouseReader.SelectContext(ctx, &rows, querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	return rows, nil
}

func (d *DataAccessService) processValidatorDashboardGroupSummaryResults(ctx context.Context, rows []*QueryResult, dashboardId t.VDBId, groupId int64, period enums.TimePeriod, currentSyncCommitteeValidators, upcomingSyncCommitteeValidators map[uint64]bool, protocolModes t.VDBProtocolModes) (*t.VDBGroupSummaryData, error) {
	ret := &t.VDBGroupSummaryData{}
	if len(rows) == 0 {
		return ret, nil
	}

	_, hours, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, err
	}

	// Initialize variables for calculations
	totalAttestationRewards := int64(0)
	totalIdealAttestationRewards := int64(0)
	totalBlockChance := float64(0)
	totalInclusionDelaySum := int64(0)
	totalInclusionDelayDivisor := int64(0)

	totalSyncExpected := float64(0)
	totalSyncScheduled := uint32(0)
	totalSyncExecuted := uint32(0)

	totalBlocksScheduled := uint32(0)
	totalBlocksProposed := uint32(0)

	totalMissedRewardsCl := int64(0)
	totalMissedRewardsAttestations := int64(0)
	totalMissedRewardsSync := int64(0)

	validators := make([]t.VDBValidator, 0)
	for _, row := range rows {
		validators = append(validators, t.VDBValidator(row.ValidatorIndex))
		totalAttestationRewards += row.AttestationReward
		totalIdealAttestationRewards += row.AttestationsIdealReward

		ret.AttestationsHead.Success += uint64(row.AttestationsHeadExecuted)
		ret.AttestationsHead.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationsHeadExecuted)

		ret.AttestationsSource.Success += uint64(row.AttestationsSourceExecuted)
		ret.AttestationsSource.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationsSourceExecuted)

		ret.AttestationsTarget.Success += uint64(row.AttestationsTargetExecuted)
		ret.AttestationsTarget.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationsTargetExecuted)

		totalMissedRewardsCl += row.BlocksCLMissedReward
		totalMissedRewardsAttestations += row.AttestationsIdealReward - row.AttestationsRewardRewardsOnly
		totalMissedRewardsSync += row.SyncLocalizedMaxRewards - row.SyncRewardRewardsOnly

		if row.ValidatorIndex == 0 && row.BlocksProposed > 0 && row.BlocksProposed != row.BlocksScheduled {
			row.BlocksProposed-- // subtract the genesis block from validator 0 (TODO: remove when fixed in the dashoard data exporter)
		}
		totalBlocksProposed += row.BlocksProposed
		totalBlocksScheduled += row.BlocksScheduled
		if row.BlocksScheduled > 0 {
			if ret.ProposalValidators == nil {
				ret.ProposalValidators = make([]t.VDBValidator, 0, 10)
			}
			ret.ProposalValidators = append(ret.ProposalValidators, t.VDBValidator(row.ValidatorIndex))
		}

		totalSyncScheduled += row.SyncScheduled
		totalSyncExecuted += row.SyncExecuted

		ret.SyncCommittee.StatusCount.Success += uint64(row.SyncExecuted)
		ret.SyncCommittee.StatusCount.Failed += uint64(row.SyncScheduled) - uint64(row.SyncExecuted)

		if row.SyncScheduled > 0 {
			if ret.SyncCommittee.Validators == nil {
				ret.SyncCommittee.Validators = make([]t.VDBValidator, 0, 10)
			}
			ret.SyncCommittee.Validators = append(ret.SyncCommittee.Validators, t.VDBValidator(row.ValidatorIndex))

			if currentSyncCommitteeValidators[uint64(row.ValidatorIndex)] {
				ret.SyncCommitteeCount.CurrentValidators++
			}
			if upcomingSyncCommitteeValidators[uint64(row.ValidatorIndex)] {
				ret.SyncCommitteeCount.UpcomingValidators++
			}
		}

		if row.SlashedInPeriod {
			ret.Slashings.StatusCount.Failed++
			ret.Slashings.Validators = append(ret.Slashings.Validators, t.VDBValidator(row.ValidatorIndex))
		}
		if row.SlashedAmount > 0 {
			ret.Slashings.StatusCount.Success += uint64(row.SlashedAmount)
			ret.Slashings.Validators = append(ret.Slashings.Validators, t.VDBValidator(row.ValidatorIndex))
		}

		totalBlockChance += row.BlockChance
		totalInclusionDelaySum += row.InclusionDelaySum
		totalSyncExpected += row.SyncCommitteesExpected

		if row.InclusionDelaySum > 0 {
			totalInclusionDelayDivisor += row.AttestationsObserved
		}
	}

	// Calculate missed rewards and other metrics
	totalMissedRewardsEl, err := d.getMissedELRewards(ctx, dashboardId, groupId, rows[0].EpochStart, rows[len(rows)-1].EpochEnd)
	if err != nil {
		return nil, err
	}

	ret.MissedRewards.Attestations = utils.GWeiToWei(big.NewInt(totalMissedRewardsAttestations))
	ret.MissedRewards.Sync = utils.GWeiToWei(big.NewInt(totalMissedRewardsSync))
	ret.MissedRewards.ProposerRewards.Cl = utils.GWeiToWei(big.NewInt(totalMissedRewardsCl))
	ret.MissedRewards.ProposerRewards.El = decimal.NewFromFloat(totalMissedRewardsEl)

	// Calculate APR and other metrics
	ret.Rewards.El, ret.Apr.El, ret.Rewards.Cl, ret.Apr.Cl, err = d.getElClAPR(ctx, dashboardId, groupId, hours)
	if err != nil {
		return nil, err
	}

	// Calculate sync committee counts
	pastSyncPeriodCutoff := utils.SyncPeriodOfEpoch(rows[0].EpochStart)
	currentSyncPeriod := utils.SyncPeriodOfEpoch(cache.LatestEpoch.Get())
	err = d.readerDb.GetContext(ctx, &ret.SyncCommitteeCount.PastPeriods, `SELECT COUNT(*) FROM sync_committees WHERE period >= $1 AND period < $2 AND validatorindex = ANY($3)`, pastSyncPeriodCutoff, currentSyncPeriod, validators)
	if err != nil {
		return nil, fmt.Errorf("error retrieving past sync committee count: %w", err)
	}

	// Calculate luck metrics
	luckHours := float64(hours)
	if hours == -1 {
		luckHours = time.Since(time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)).Hours()
		if luckHours == 0 {
			luckHours = 24
		}
	}

	if totalBlockChance > 0 {
		ret.Luck.Proposal.Percent = (float64(totalBlocksScheduled)) / totalBlockChance * 100
		ret.Luck.Proposal.AverageIntervalSeconds = uint64(time.Duration((luckHours / totalBlockChance) * float64(time.Hour)).Seconds())
		ret.Luck.Proposal.ExpectedTimestamp = uint64(time.Now().Unix()) + ret.Luck.Proposal.AverageIntervalSeconds
	} else {
		ret.Luck.Proposal.Percent = 0
	}

	if totalSyncExpected == 0 {
		ret.Luck.Sync.Percent = 0
	} else {
		totalSyncSlotDuties := float64(ret.SyncCommittee.StatusCount.Failed) + float64(ret.SyncCommittee.StatusCount.Success)
		slotDutiesPerSyncCommittee := float64(utils.SlotsPerSyncCommittee())
		syncCommittees := math.Ceil(totalSyncSlotDuties / slotDutiesPerSyncCommittee)
		ret.Luck.Sync.Percent = syncCommittees / totalSyncExpected * 100
		ret.Luck.Sync.AverageIntervalSeconds = uint64(time.Duration((luckHours / totalSyncExpected) * float64(time.Hour)).Seconds())
		ret.Luck.Sync.ExpectedTimestamp = uint64(time.Now().Unix()) + ret.Luck.Sync.AverageIntervalSeconds
	}

	if totalInclusionDelayDivisor > 0 {
		ret.AttestationAvgInclDist = 1.0 + float64(totalInclusionDelaySum)/float64(totalInclusionDelayDivisor)
	} else {
		ret.AttestationAvgInclDist = 0
	}

	ret.ProposalValidatorCount = uint64(len(ret.ProposalValidators))
	ret.Slashings.ValidatorCount = uint64(len(ret.Slashings.Validators))
	ret.SyncCommittee.ValidatorCount = uint64(len(ret.SyncCommittee.Validators))

	// Sort and limit validators
	if len(ret.ProposalValidators) > 0 {
		slices.Sort(ret.ProposalValidators)
		if len(ret.ProposalValidators) > 3 {
			ret.ProposalValidators = ret.ProposalValidators[:3]
		}
	}
	if len(ret.Slashings.Validators) > 0 {
		slices.Sort(ret.Slashings.Validators)
		if len(ret.Slashings.Validators) > 3 {
			ret.Slashings.Validators = ret.Slashings.Validators[:3]
		}
	}
	if len(ret.SyncCommittee.Validators) > 0 {
		slices.Sort(ret.SyncCommittee.Validators)
		if len(ret.SyncCommittee.Validators) > 3 {
			ret.SyncCommittee.Validators = ret.SyncCommittee.Validators[:3]
		}
	}

	// Calculate efficiency
	var attestationEfficiency, proposerEfficiency, syncEfficiency sql.NullFloat64
	if totalIdealAttestationRewards > 0 {
		attestationEfficiency.Float64 = decimal.NewFromInt(totalAttestationRewards).Div(decimal.NewFromInt(totalIdealAttestationRewards)).InexactFloat64()
		attestationEfficiency.Valid = true
		ret.AttestationEfficiency = max(attestationEfficiency.Float64*100, 0)
	}
	if totalBlocksScheduled > 0 {
		proposerEfficiency.Float64 = float64(totalBlocksProposed) / float64(totalBlocksScheduled)
		proposerEfficiency.Valid = true
	}
	if totalSyncScheduled > 0 {
		syncEfficiency.Float64 = float64(totalSyncExecuted) / float64(totalSyncScheduled)
		syncEfficiency.Valid = true
	}
	ret.Efficiency = utils.CalculateTotalEfficiency(attestationEfficiency, proposerEfficiency, syncEfficiency)

	// Fetch additional data
	rpOperatorInfo, err := d.getValidatorDashboardRpOperatorInfo(ctx, dashboardId)
	if err != nil {
		return nil, err
	}
	validatorMapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, err
	}
	balances, err := d.calculateValidatorDashboardBalance(ctx, rpOperatorInfo, validators, validatorMapping, protocolModes)
	if err != nil {
		return nil, err
	}
	ret.Balances = balances

	return ret, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupSummary(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod, protocolModes t.VDBProtocolModes) (*t.VDBGroupSummaryData, error) {
	// TODO: implement data retrieval for the following new field
	// Fetch validator list for user dashboard from the dashboard table when querying the past sync committees as the rolling table might miss exited validators
	// TotalMissedRewards
	// @DATA-ACCESS incorporate protocolModes
	// @DATA-ACCESS implement data retrieval for Rocket Pool stats (if present)
	if dashboardId.AggregateGroups {
		groupId = t.AllGroups
	}

	latestEpoch := cache.LatestEpoch.Get()
	currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, err := d.getCurrentAndUpcomingSyncCommittees(ctx, latestEpoch)
	if err != nil {
		return nil, err
	}

	ds, err := d.buildValidatorDashboardGroupSummaryQuery(dashboardId, groupId, period)
	if err != nil {
		return nil, err
	}

	rows, err := d.executeValidatorDashboardGroupSummaryQuery(ctx, ds)
	if err != nil {
		return nil, err
	}

	return d.processValidatorDashboardGroupSummaryResults(ctx, rows, dashboardId, groupId, period, currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, protocolModes)
}
