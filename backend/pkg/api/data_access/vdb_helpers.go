package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/services"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/price"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

//////////////////// 		Helper functions (must be used by more than one VDB endpoint!)

func (d DataAccessService) getDashboardValidators(ctx context.Context, dashboardId t.VDBId, groupIds []uint64) ([]t.VDBValidator, error) {
	if len(dashboardId.Validators) == 0 {
		ds := goqu.Dialect("postgres").
			Select("validator_index").
			From("users_val_dashboards_validators").
			Where(goqu.L("dashboard_id = ?", dashboardId.Id)).
			Order(goqu.I("validator_index").Asc())

		if len(groupIds) > 0 {
			ds = ds.Where(goqu.L("group_id = ANY(?)", pq.Array(groupIds)))
		}

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return nil, err
		}

		var validatorsArray []t.VDBValidator
		err = d.alloyReader.SelectContext(ctx, &validatorsArray, query, args...)
		return validatorsArray, err
	}
	return dashboardId.Validators, nil
}

func (d DataAccessService) calculateChartEfficiency(efficiencyType enums.VDBSummaryChartEfficiencyType, row *t.VDBValidatorSummaryChartRow) (float64, error) {
	efficiency := float64(0)
	switch efficiencyType {
	case enums.VDBSummaryChartAll:
		var attestationEfficiency, proposerEfficiency, syncEfficiency sql.NullFloat64
		if row.AttestationIdealReward > 0 {
			attestationEfficiency.Float64 = row.AttestationReward / row.AttestationIdealReward
			attestationEfficiency.Valid = true
		}
		if row.BlocksScheduled > 0 {
			proposerEfficiency.Float64 = row.BlocksProposed / row.BlocksScheduled
			proposerEfficiency.Valid = true
		}
		if row.SyncScheduled > 0 {
			syncEfficiency.Float64 = row.SyncExecuted / row.SyncScheduled
			syncEfficiency.Valid = true
		}

		efficiency = utils.CalculateTotalEfficiency(attestationEfficiency, proposerEfficiency, syncEfficiency)
	case enums.VDBSummaryChartAttestation:
		if row.AttestationIdealReward > 0 {
			efficiency = (row.AttestationReward / row.AttestationIdealReward) * 100
		} else {
			efficiency = 100
		}
	case enums.VDBSummaryChartProposal:
		if row.BlocksScheduled > 0 {
			efficiency = (row.BlocksProposed / row.BlocksScheduled) * 100
		} else {
			efficiency = 100
		}
	case enums.VDBSummaryChartSync:
		if row.SyncScheduled > 0 {
			efficiency = (row.SyncExecuted / row.SyncScheduled) * 100
		} else {
			efficiency = 100
		}
	default:
		return 0, fmt.Errorf("unexpected efficiency type: %v", efficiency)
	}
	return efficiency, nil
}

func (d *DataAccessService) getWithdrawableCountFromCursor(validatorindex t.VDBValidator, cursor uint64) (uint64, error) {
	// the validators' balance will not be checked here as this is only a rough estimation
	// checking the balance for hundreds of thousands of validators is too expensive

	stats := cache.LatestStats.Get()
	if stats == nil || stats.ActiveValidatorCount == nil || stats.TotalValidatorCount == nil {
		return 0, errors.New("stats not available")
	}

	var maxValidatorIndex t.VDBValidator
	if *stats.TotalValidatorCount > 0 {
		maxValidatorIndex = *stats.TotalValidatorCount - 1
	}
	if maxValidatorIndex == 0 {
		return 0, nil
	}

	activeValidators := *stats.ActiveValidatorCount
	if activeValidators == 0 {
		activeValidators = maxValidatorIndex
	}

	if validatorindex > cursor {
		// if the validatorindex is after the cursor, simply return the number of validators between the cursor and the validatorindex
		// the returned data is then scaled using the number of currently active validators in order to account for exited / entering validators
		return (validatorindex - cursor) * activeValidators / maxValidatorIndex, nil
	} else if validatorindex < cursor {
		// if the validatorindex is before the cursor (wraparound case) return the number of validators between the cursor and the most recent validator plus the amount of validators from the validator 0 to the validatorindex
		// the returned data is then scaled using the number of currently active validators in order to account for exited / entering validators
		return (maxValidatorIndex - cursor + validatorindex) * activeValidators / maxValidatorIndex, nil
	} else {
		return 0, nil
	}
}

// GetTimeToNextWithdrawal calculates the time it takes for the validators next withdrawal to be processed.
func (d *DataAccessService) getTimeToNextWithdrawal(distance uint64) time.Time {
	minTimeToWithdrawal := time.Now().Add(time.Second * time.Duration((distance/utils.Config.Chain.ClConfig.MaxValidatorsPerWithdrawalSweep)*utils.Config.Chain.ClConfig.SecondsPerSlot))
	timeToWithdrawal := time.Now().Add(time.Second * time.Duration((float64(distance)/float64(utils.Config.Chain.ClConfig.MaxWithdrawalsPerPayload))*float64(utils.Config.Chain.ClConfig.SecondsPerSlot)))

	if timeToWithdrawal.Before(minTimeToWithdrawal) {
		return minTimeToWithdrawal
	}

	return timeToWithdrawal
}

func (d *DataAccessService) getElClAPR(ctx context.Context, dashboardId t.VDBId, groupId int64, hours int) (elIncome decimal.Decimal, elAPR float64, clIncome decimal.Decimal, clAPR float64, err error) {
	table := ""

	switch hours {
	case 1:
		table = "validator_dashboard_data_rolling_1h"
	case 24:
		table = "validator_dashboard_data_rolling_24h"
	case 7 * 24:
		table = "validator_dashboard_data_rolling_7d"
	case 30 * 24:
		table = "validator_dashboard_data_rolling_30d"
	case -1:
		table = "validator_dashboard_data_rolling_90d"
	default:
		return decimal.Zero, 0, decimal.Zero, 0, fmt.Errorf("invalid hours value: %v", hours)
	}

	type RewardsResult struct {
		EpochStart     uint64        `db:"epoch_start"`
		EpochEnd       uint64        `db:"epoch_end"`
		ValidatorCount uint64        `db:"validator_count"`
		Reward         sql.NullInt64 `db:"reward"`
	}

	var rewardsResultTable RewardsResult
	var rewardsResultTotal RewardsResult

	rewardsDs := goqu.Dialect("postgres").
		From(goqu.L(fmt.Sprintf("%s AS r FINAL", table))).
		With("validators", goqu.L("(SELECT group_id, validator_index FROM users_val_dashboards_validators WHERE dashboard_id = ?)", dashboardId.Id)).
		Select(
			goqu.L("MIN(epoch_start) AS epoch_start"),
			goqu.L("MAX(epoch_end) AS epoch_end"),
			goqu.L("COUNT(*) AS validator_count"),
			goqu.L(`
				(
					SUM(COALESCE(finalizeAggregation(r.balance_end), 0)) +
					SUM(COALESCE(r.withdrawals_amount, 0)) -
					SUM(COALESCE(r.deposits_amount, 0)) -
					SUM(COALESCE(finalizeAggregation(r.balance_start), 0))
				) AS reward
			`))
	if len(dashboardId.Validators) > 0 {
		rewardsDs = rewardsDs.
			Where(goqu.L("validator_index IN ?", dashboardId.Validators))
	} else {
		rewardsDs = rewardsDs.
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
			Where(goqu.L("r.validator_index IN (SELECT validator_index FROM validators)"))

		if groupId != -1 {
			rewardsDs = rewardsDs.
				Where(goqu.L("v.group_id = ?", groupId))
		}
	}

	query, args, err := rewardsDs.Prepared(true).ToSQL()
	if err != nil {
		return decimal.Zero, 0, decimal.Zero, 0, fmt.Errorf("error preparing query: %w", err)
	}

	err = d.clickhouseReader.GetContext(ctx, &rewardsResultTable, query, args...)
	if err != nil || !rewardsResultTable.Reward.Valid {
		return decimal.Zero, 0, decimal.Zero, 0, err
	}

	if rewardsResultTable.ValidatorCount == 0 {
		return decimal.Zero, 0, decimal.Zero, 0, nil
	}

	aprDivisor := hours
	if hours == -1 { // for all time APR
		aprDivisor = 90 * 24
	}

	// invested amount is not post-pectra safe
	investedAmount := d.convertClToMain(decimal.NewFromUint64(d.config.ClConfig.MaxEffectiveBalance))

	clAPR = calcAPR(d.convertClToMain(decimal.NewFromInt(rewardsResultTable.Reward.Int64)), investedAmount, aprDivisor, rewardsResultTable.ValidatorCount)

	clIncome = decimal.NewFromInt(rewardsResultTable.Reward.Int64).Mul(decimal.NewFromInt(1e9))

	if hours == -1 {
		rewardsDs = rewardsDs.
			From(goqu.L("validator_dashboard_data_rolling_total AS r FINAL"))

		query, args, err = rewardsDs.Prepared(true).ToSQL()
		if err != nil {
			return decimal.Zero, 0, decimal.Zero, 0, fmt.Errorf("error preparing query: %w", err)
		}

		err = d.clickhouseReader.GetContext(ctx, &rewardsResultTotal, query, args...)
		if err != nil || !rewardsResultTotal.Reward.Valid {
			return decimal.Zero, 0, decimal.Zero, 0, err
		}

		clIncome = decimal.NewFromInt(rewardsResultTotal.Reward.Int64).Mul(decimal.NewFromInt(1e9))
	}

	elDs := goqu.Dialect("postgres").
		Select(goqu.COALESCE(goqu.SUM(goqu.L("value")), 0)).
		From(goqu.I("execution_rewards_finalized").As("b"))

	if len(dashboardId.Validators) > 0 {
		elDs = elDs.
			Where(goqu.L("b.proposer = ANY(?)", pq.Array(dashboardId.Validators)))
	} else {
		elDs = elDs.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("b.proposer = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

		if groupId != -1 {
			elDs = elDs.
				Where(goqu.L("v.group_id = ?", groupId))
		}
	}

	elTableDs := elDs.
		Where(goqu.L("b.epoch >= ? AND b.epoch <= ?", rewardsResultTable.EpochStart, rewardsResultTable.EpochEnd))

	query, args, err = elTableDs.Prepared(true).ToSQL()
	if err != nil {
		return decimal.Zero, 0, decimal.Zero, 0, fmt.Errorf("error preparing query: %w", err)
	}

	err = d.alloyReader.GetContext(ctx, &elIncome, query, args...)
	if err != nil {
		return decimal.Zero, 0, decimal.Zero, 0, err
	}

	elAPR = calcAPR(d.convertElToMain(elIncome), investedAmount, aprDivisor, rewardsResultTable.ValidatorCount)

	if hours == -1 {
		elTotalDs := elDs.
			Where(goqu.L("b.epoch >= ? AND b.epoch <= ?", rewardsResultTotal.EpochStart, rewardsResultTotal.EpochEnd))

		query, args, err = elTotalDs.Prepared(true).ToSQL()
		if err != nil {
			return decimal.Zero, 0, decimal.Zero, 0, fmt.Errorf("error preparing query: %w", err)
		}

		err = d.alloyReader.GetContext(ctx, &elIncome, query, args...)
		if err != nil {
			return decimal.Zero, 0, decimal.Zero, 0, err
		}
	}

	return elIncome, elAPR, clIncome, clAPR, nil
}

// precondition: invested amount and rewards are in the same currency
func calcAPR(rewards, investedAmount decimal.Decimal, aprDivisor int, validatorCount uint64) float64 {
	if rewards.IsZero() || investedAmount.IsZero() || validatorCount == 0 {
		return 0
	}
	return (rewards.Div(decimal.NewFromInt(int64(aprDivisor))).Div(investedAmount.Mul(decimal.NewFromInt(int64(validatorCount)))).Mul(decimal.NewFromInt(24 * 365 * 100))).InexactFloat64()
}

// converts a cl amount to the main currency
func (d *DataAccessService) convertClToMain(amount decimal.Decimal) decimal.Decimal {
	price := decimal.NewFromFloat(price.GetPrice(d.config.Frontend.MainCurrency, d.config.Frontend.ClCurrency))
	return amount.Div(decimal.NewFromInt(d.config.Frontend.ClCurrencyDivisor)).Div(price)
}

// converts a el amount to the main currency
func (d *DataAccessService) convertElToMain(amount decimal.Decimal) decimal.Decimal {
	price := decimal.NewFromFloat(price.GetPrice(d.config.Frontend.MainCurrency, d.config.Frontend.ElCurrency))
	return amount.Div(decimal.NewFromInt(d.config.Frontend.ElCurrencyDivisor)).Div(price)
}

type RpOperatorInfo struct {
	ValidatorIndex     uint64          `db:"validatorindex"`
	NodeFee            float64         `db:"node_fee"`
	NodeDepositBalance decimal.Decimal `db:"node_deposit_balance"`
	UserDepositBalance decimal.Decimal `db:"user_deposit_balance"`
}

func (d *DataAccessService) getValidatorDashboardRpOperatorInfo(ctx context.Context, dashboardId t.VDBId) ([]RpOperatorInfo, error) {
	var rpOperatorInfo []RpOperatorInfo

	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("v.validatorindex"),
			goqu.L("rplm.node_fee"),
			goqu.L("rplm.node_deposit_balance"),
			goqu.L("rplm.user_deposit_balance")).
		From(goqu.L("rocketpool_minipools AS rplm")).
		LeftJoin(goqu.L("validators AS v"), goqu.On(goqu.L("rplm.pubkey = v.pubkey"))).
		Where(goqu.L("node_deposit_balance IS NOT NULL")).
		Where(goqu.L("user_deposit_balance IS NOT NULL"))

	if len(dashboardId.Validators) == 0 {
		ds = ds.
			LeftJoin(goqu.L("users_val_dashboards_validators uvdv"), goqu.On(goqu.L("uvdv.validator_index = v.validatorindex"))).
			Where(goqu.L("uvdv.dashboard_id = ?", dashboardId.Id))
	} else {
		ds = ds.
			Where(goqu.L("v.validatorindex = ANY(?)", pq.Array(dashboardId.Validators)))
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	err = d.alloyReader.SelectContext(ctx, &rpOperatorInfo, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving rocketpool validators data: %w", err)
	}
	return rpOperatorInfo, nil
}

func (d *DataAccessService) calculateValidatorDashboardBalance(ctx context.Context, rpOperatorInfo []RpOperatorInfo, validators []t.VDBValidator, validatorMapping *services.ValidatorMapping, protocolModes t.VDBProtocolModes) (t.ValidatorBalances, error) {
	balances := t.ValidatorBalances{}

	rpValidators := make(map[uint64]RpOperatorInfo)
	for _, res := range rpOperatorInfo {
		rpValidators[res.ValidatorIndex] = res
	}

	// Create a new sub-dashboard to get the total cl deposits for non-rocketpool validators
	var nonRpDashboardId t.VDBId

	for _, validator := range validators {
		metadata := validatorMapping.ValidatorMetadata[validator]
		validatorBalance := utils.GWeiToWei(big.NewInt(int64(metadata.Balance)))
		effectiveBalance := utils.GWeiToWei(big.NewInt(int64(metadata.EffectiveBalance)))

		if rpValidator, ok := rpValidators[validator]; ok {
			if protocolModes.RocketPool {
				// Calculate the balance of the operator
				fullDeposit := rpValidator.UserDepositBalance.Add(rpValidator.NodeDepositBalance)
				operatorShare := rpValidator.NodeDepositBalance.Div(fullDeposit)
				invOperatorShare := decimal.NewFromInt(1).Sub(operatorShare)

				base := decimal.Min(decimal.Max(decimal.Zero, validatorBalance.Sub(rpValidator.UserDepositBalance)), rpValidator.NodeDepositBalance)
				commission := decimal.Max(decimal.Zero, validatorBalance.Sub(fullDeposit).Mul(invOperatorShare).Mul(decimal.NewFromFloat(rpValidator.NodeFee)))
				reward := decimal.Max(decimal.Zero, validatorBalance.Sub(fullDeposit).Mul(operatorShare).Add(commission))

				operatorBalance := base.Add(reward)

				balances.Total = balances.Total.Add(operatorBalance)
			} else {
				balances.Total = balances.Total.Add(validatorBalance)
			}
			balances.StakedEth = balances.StakedEth.Add(rpValidator.NodeDepositBalance)
		} else {
			balances.Total = balances.Total.Add(validatorBalance)

			nonRpDashboardId.Validators = append(nonRpDashboardId.Validators, validator)
		}
		balances.Effective = balances.Effective.Add(effectiveBalance)
	}

	// Get the total cl deposits for non-rocketpool validators
	if len(nonRpDashboardId.Validators) > 0 {
		totalNonRpDeposits, err := d.GetValidatorDashboardTotalClDeposits(ctx, nonRpDashboardId)
		if err != nil {
			return balances, fmt.Errorf("error retrieving total cl deposits for non-rocketpool validators: %w", err)
		}
		balances.StakedEth = balances.StakedEth.Add(totalNonRpDeposits.TotalAmount)
	}
	return balances, nil
}

// Builds a query to select the maximum timestamp from the validator_dashboard_data view for an interval (epoch, hourly, daily, weekly)
func buildLatestExportedChartQuery(view, dateColumn string) *goqu.SelectDataset {
	return goqu.Dialect("postgres").From(view).Select(goqu.MAX(dateColumn))
}

// Converts a time.Time value into a Unix timestamp (uint64).
func processLastExportedChartResult(ts time.Time) (uint64, error) {
	return uint64(ts.Unix()), nil
}

// Retrieves the latest exported chart timestamp for a given aggregation level
// 1. Determines the appropriate view and date column based on the aggregation type
// 2. Constructs and executes the query to fetch the latest timestamp
// 3. Converts the result into a Unix timestamp and returns it
func (d *DataAccessService) GetLatestExportedChartTs(ctx context.Context, aggregation enums.ChartAggregation) (uint64, error) {
	view, dateColumn, err := getViewAndDateColumn(aggregation)
	if err != nil {
		return 0, err
	}

	ds := buildLatestExportedChartQuery(view, dateColumn)

	ts, err := runQuery[time.Time](ctx, d.clickhouseReader, ds)
	if err != nil {
		return 0, err
	}

	return processLastExportedChartResult(ts)
}

type SyncCommitteeResult struct {
	ValidatorIndex uint64 `db:"validatorindex"`
	Period         uint64 `db:"period"`
}

// Pass the latest (most recently finalized epoch)
// to determine validators current and upcoming sync committees
func (d *DataAccessService) getCurrentAndUpcomingSyncCommittees(ctx context.Context, latestEpoch uint64) (map[uint64]bool, map[uint64]bool, error) {
	currentSyncPeriod := utils.SyncPeriodOfEpoch(latestEpoch)
	ds := buildCurrentAndUpcomingSyncCommitteesQuery(currentSyncPeriod)
	queryResult, err := runQueryRows[[]SyncCommitteeResult](ctx, d.readerDb, ds)
	if err != nil {
		return nil, nil, err
	}
	return processSyncCommitteeResults(queryResult, utils.SyncPeriodOfEpoch(latestEpoch))
}

// Builds a query to fetch sync committee validators for the current and next period
func buildCurrentAndUpcomingSyncCommitteesQuery(currentSyncPeriod uint64) *goqu.SelectDataset {
	return goqu.Dialect("postgres").
		Select(
			goqu.L("validatorindex"),
			goqu.L("period"),
		).
		From("sync_committees").
		Where(goqu.L("period IN (?, ?)", currentSyncPeriod, currentSyncPeriod+1))
}

// Processes query results and categorizes validators into current and upcoming sync committees
func processSyncCommitteeResults(queryResult []SyncCommitteeResult, currentSyncPeriod uint64) (map[uint64]bool, map[uint64]bool, error) {
	currentSyncCommitteeValidators := make(map[uint64]bool)
	upcomingSyncCommitteeValidators := make(map[uint64]bool)

	for _, entry := range queryResult {
		if entry.Period == currentSyncPeriod {
			currentSyncCommitteeValidators[entry.ValidatorIndex] = true
		} else {
			upcomingSyncCommitteeValidators[entry.ValidatorIndex] = true
		}
	}

	return currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, nil
}

// Retrieves the start epoch for a given time period (last 1h, 24h, 7d, 30d)
func (d *DataAccessService) getEpochStart(ctx context.Context, period enums.TimePeriod) (uint64, error) {
	clickhouseTable, _, err := getTablesForPeriod(period)
	if err != nil {
		return 0, err
	}

	ds := buildEpochStartQuery(clickhouseTable)
	epochStart, err := runQuery[uint64](ctx, d.clickhouseReader, ds)
	if err != nil {
		return 0, err
	}

	return epochStart, nil
}

// Builds a query to retrieve the earliest epoch start value
// Orders results in ascending order and limits to the first row.
func buildEpochStartQuery(table string) *goqu.SelectDataset {
	return goqu.Dialect("postgres").
		Select(goqu.L("epoch_start")).
		From(goqu.L(fmt.Sprintf("%s FINAL", table))).
		Order(goqu.L("epoch_start").Asc()).
		Limit(1)
}

// Retrieves past sync committee validators for the given validator indices and epoch range
// 1. Determines the sync period range from the epochStart and latestEpoch values
// 2. Constructs a query to fetch validators who were part of past sync committees within the given range
// 3. Processes the results to count occurrences of each validator in past committees
func (d *DataAccessService) getPastSyncCommittees(ctx context.Context, indices []uint64, epochStart uint64, latestEpoch uint64) (map[uint64]uint64, error) {
	pastSyncPeriodCutoff := utils.SyncPeriodOfEpoch(epochStart)
	currentSyncPeriod := utils.SyncPeriodOfEpoch(latestEpoch)

	ds := buildPastSyncCommitteesQuery(indices, pastSyncPeriodCutoff, currentSyncPeriod)

	validatorIndices, err := runQueryRows[[]uint64](ctx, d.alloyReader, ds)
	if err != nil {
		return nil, err
	}

	return processPastSyncCommitteesResults(validatorIndices)
}

// Builds a query to retrieve past sync committee validators within a given sync period range
func buildPastSyncCommitteesQuery(indices []uint64, pastSyncPeriodCutoff, currentSyncPeriod uint64) *goqu.SelectDataset {
	return goqu.Dialect("postgres").
		Select(goqu.L("sc.validatorindex")).
		From(goqu.L("sync_committees sc")).
		Where(goqu.L("period >= ? AND period < ? AND validatorindex = ANY(?)", pastSyncPeriodCutoff, currentSyncPeriod, pq.Array(indices)))
}

// EpochRange represents the result structure for min and max epochs
type EpochRange struct {
	MinEpochStart *uint64 `db:"min_epoch_start"`
	MaxEpochEnd   *uint64 `db:"max_epoch_end"`
}

// Constructs the SQL query for retrieving min/max epochs
func buildMinMaxEpochsQuery(dashboardId t.VDBId, groupId int64, clickhouseTable string) *goqu.SelectDataset {
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
	return ds
}

// GetMinMaxEpochs is the main function that ties all steps together
func (d *DataAccessService) getMinMaxEpochs(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (uint64, uint64, error) {
	clickhouseTable, _, err := getTablesForPeriod(period)
	if err != nil {
		return 0, 0, err
	}

	ds := buildMinMaxEpochsQuery(dashboardId, groupId, clickhouseTable)

	epochRow, err := runQuery[EpochRange](ctx, d.clickhouseReader, ds)
	if err != nil {
		return 0, 0, err
	}

	return *epochRow.MinEpochStart, *epochRow.MaxEpochEnd, nil
}

type LastScheduledEpoch struct {
	LastScheduledBlockEpoch *int64 `db:"last_scheduled_block_epoch"`
	LastSyncEpoch           *int64 `db:"last_scheduled_sync_epoch"`
}

// Constructs the SQL query for retrieving last scheduled block and sync epochs
func buildLastScheduledBlockAndSyncDateQuery(clickhouseTable string, dashboardId t.VDBId, groupId int64) *goqu.SelectDataset {
	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("MAX(last_scheduled_block_epoch) as last_scheduled_block_epoch"),
			goqu.L("MAX(last_scheduled_sync_epoch) as last_scheduled_sync_epoch")).
		From(goqu.L(fmt.Sprintf(`%s AS r FINAL`, clickhouseTable)))

	if dashboardId.Validators != nil {
		// If Validators are provided, use them directly in the WHERE clause
		return ds.Where(goqu.L("validator_index IN ?", dashboardId.Validators))
	}

	// If Validators are nil, use a subquery
	return ds.
		With("validators", goqu.L("(SELECT validator_index, group_id FROM users_val_dashboards_validators WHERE dashboard_id = ? AND (group_id = ? OR ?::smallint = -1))", dashboardId.Id, groupId, groupId)).
		InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
		Where(goqu.L("validator_index IN (SELECT validator_index FROM validators)"))
}

// Gets last scheduled block/sync committee epoch
func (d *DataAccessService) getLastScheduledBlockAndSyncDate(ctx context.Context, dashboardId t.VDBId, groupId int64) (time.Time, time.Time, error) {
	// we need to use clickhouse table for all_time period
	clickhouseTotalTable, _, err := getTablesForPeriod(enums.AllTime)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	ds := buildLastScheduledBlockAndSyncDateQuery(clickhouseTotalTable, dashboardId, groupId)

	epochRow, err := runQuery[LastScheduledEpoch](ctx, d.clickhouseReader, ds)
	if err != nil {
		return time.Time{}, time.Time{}, err
	} else if epochRow.LastScheduledBlockEpoch == nil || epochRow.LastSyncEpoch == nil {
		return time.Time{}, time.Time{}, nil
	} else {
		return utils.EpochToTime(uint64(*epochRow.LastScheduledBlockEpoch)),
			utils.EpochToTime(uint64(*epochRow.LastSyncEpoch)),
			nil
	}
}

// Builds a query to calculate the total missed EL rewards for a given epoch range
func buildMissedELRewardsQuery(dashboardId t.VDBId, groupId int64, slots, epochStart, epochEnd uint64) *goqu.SelectDataset {
	// Define the `targets` CTE
	targets := goqu.Dialect("postgres").
		From("blocks").
		Select(goqu.I("blocks.slot").As("slot")).
		Where(
			goqu.I("blocks.status").Neq("1"),
			goqu.I("epoch").Gte(epochStart),
			goqu.I("epoch").Lte(epochEnd),
		)

	// Filter validators
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

	// Define the res CTE
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

	return query
}

// Get total median EL rewards
func (d *DataAccessService) getMissedELRewards(ctx context.Context, dashboardId t.VDBId, groupId int64, epochStart, epochEnd uint64) (float64, error) {
	slots := utils.Config.Chain.ClConfig.SlotsPerEpoch / 2
	query := buildMissedELRewardsQuery(dashboardId, groupId, slots, epochStart, epochEnd)

	totalMissedRewardsEl, err := runQuery[float64](ctx, d.readerDb, query)
	if err != nil {
		return 0, err
	}

	return totalMissedRewardsEl, nil
}

// Processes past sync committee results by counting occurrences of each validator
func processPastSyncCommitteesResults(validatorIndices []uint64) (map[uint64]uint64, error) {
	validatorCountMap := make(map[uint64]uint64)
	for _, validatorIndex := range validatorIndices {
		validatorCountMap[validatorIndex]++
	}
	return validatorCountMap, nil
}

// Determines the validator dashboard data table and associated time duration in hours
// based on the given time period (1h, 24h, 7d, 30d, all_time)
func getTablesForPeriod(period enums.TimePeriod) (string, int, error) {
	table := ""
	hours := 0

	switch period {
	case enums.TimePeriods.Last1h:
		table = "validator_dashboard_data_rolling_1h"
		hours = 1
	case enums.TimePeriods.Last24h:
		table = "validator_dashboard_data_rolling_24h"
		hours = 24
	case enums.TimePeriods.Last7d:
		table = "validator_dashboard_data_rolling_7d"
		hours = 7 * 24
	case enums.TimePeriods.Last30d:
		table = "validator_dashboard_data_rolling_30d"
		hours = 30 * 24
	case enums.TimePeriods.AllTime:
		table = "validator_dashboard_data_rolling_total"
		hours = -1
	default:
		return "", 0, fmt.Errorf("not-implemented time period: %v", period)
	}

	return table, hours, nil
}

// Retrieves the validator dashboard data table and corresponding date column
// for a given chart aggregation range (epoch, hourly, daily, weekly)
func getTableAndDateColumn(aggregation enums.ChartAggregation) (string, string, error) {
	var table, dateColumn string

	switch aggregation {
	case enums.IntervalEpoch:
		table = "validator_dashboard_data_epoch"
		dateColumn = "epoch_timestamp"
	case enums.IntervalHourly:
		table = "validator_dashboard_data_hourly"
		dateColumn = "t"
	case enums.IntervalDaily:
		table = "validator_dashboard_data_daily"
		dateColumn = "t"
	case enums.IntervalWeekly:
		table = "validator_dashboard_data_weekly"
		dateColumn = "t"
	default:
		return "", "", fmt.Errorf("unexpected aggregation type: %v", aggregation)
	}

	return table, dateColumn, nil
}

// Retrieves the validator dashboard data view and corresponding date column
// for a given chart aggregation range (epoch, hourly, daily, weekly)
func getViewAndDateColumn(aggregation enums.ChartAggregation) (string, string, error) {
	var view, dateColumn string

	switch aggregation {
	case enums.IntervalEpoch:
		view = "view_validator_dashboard_data_epoch_max_ts"
		dateColumn = "t"
	case enums.IntervalHourly:
		view = "view_validator_dashboard_data_hourly_max_ts"
		dateColumn = "t"
	case enums.IntervalDaily:
		view = "view_validator_dashboard_data_daily_max_ts"
		dateColumn = "t"
	case enums.IntervalWeekly:
		view = "view_validator_dashboard_data_weekly_max_ts"
		dateColumn = "t"
	default:
		return "", "", fmt.Errorf("unexpected aggregation type: %v", aggregation)
	}

	return view, dateColumn, nil
}
