package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
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
	clAPR = ((float64(rewardsResultTable.Reward.Int64) / float64(aprDivisor)) / (float64(32e9) * float64(rewardsResultTable.ValidatorCount))) * 24.0 * 365.0 * 100.0
	if math.IsNaN(clAPR) {
		clAPR = 0
	}

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
		Select(goqu.COALESCE(goqu.SUM(goqu.L("value / 1e18")), 0)).
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
	elIncomeFloat, _ := elIncome.Float64() // EL income is in ETH
	elAPR = ((elIncomeFloat / float64(aprDivisor)) / (float64(32) * float64(rewardsResultTable.ValidatorCount))) * 24.0 * 365.0 * 100.0
	if math.IsNaN(elAPR) {
		elAPR = 0
	}

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
	elIncome = elIncome.Mul(decimal.NewFromInt(1e18))

	return elIncome, elAPR, clIncome, clAPR, nil
}
