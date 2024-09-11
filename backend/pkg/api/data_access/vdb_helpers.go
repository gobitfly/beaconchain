package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
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

func (d *DataAccessService) getRocketPoolInfos(ctx context.Context, dashboardId t.VDBId, groupId int64) (*t.RPInfo, error) {
	wg := errgroup.Group{}

	queryResult := []struct {
		ValidatorIndex       uint64           `db:"validatorindex"`
		NodeAddress          []byte           `db:"node_address"`
		NodeFee              float64          `db:"node_fee"`
		NodeDepositBalance   decimal.Decimal  `db:"node_deposit_balance"`
		UserDepositBalance   decimal.Decimal  `db:"user_deposit_balance"`
		EndTime              sql.NullTime     `db:"end_time"`
		SmoothingPoolAddress []byte           `db:"smoothing_pool_address"`
		SmoothingPoolEth     *decimal.Decimal `db:"smoothing_pool_eth"`
	}{}

	wg.Go(func() error {
		ds := goqu.Dialect("postgres").
			Select(
				goqu.L("v.validatorindex"),
				goqu.L("rplm.node_address"),
				goqu.L("rplm.node_fee"),
				goqu.L("rplm.node_deposit_balance"),
				goqu.L("rplm.user_deposit_balance"),
				goqu.L("rplrs.end_time"),
				goqu.L("rploc.smoothing_pool_address"),
				goqu.L("rplrs.smoothing_pool_eth"),
			).
			From(goqu.L("rocketpool_minipools AS rplm")).
			LeftJoin(goqu.L("validators AS v"), goqu.On(goqu.L("rplm.pubkey = v.pubkey"))).
			LeftJoin(goqu.L("rocketpool_rewards_summary AS rplrs"), goqu.On(goqu.L("rplm.node_address = rplrs.node_address"))).
			LeftJoin(goqu.L("rocketpool_onchain_configs AS rploc"), goqu.On(goqu.L("rplm.rocketpool_storage_address = rploc.rocketpool_storage_address"))).
			Where(goqu.L("rplm.node_deposit_balance IS NOT NULL")).
			Where(goqu.L("rplm.user_deposit_balance IS NOT NULL"))

		if len(dashboardId.Validators) == 0 {
			ds = ds.
				LeftJoin(goqu.L("users_val_dashboards_validators uvdv"), goqu.On(goqu.L("uvdv.validator_index = v.validatorindex"))).
				Where(goqu.L("uvdv.dashboard_id = ?", dashboardId.Id))

			if groupId != t.AllGroups {
				ds = ds.
					Where(goqu.L("uvdv.group_id = ?", groupId))
			}
		} else {
			ds = ds.
				Where(goqu.L("v.validatorindex = ANY(?)", pq.Array(dashboardId.Validators)))
		}

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.alloyReader.SelectContext(ctx, &queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving rocketpool validators data: %w", err)
		}

		return nil
	})

	nodeMinipoolCount := make(map[string]uint64)
	wg.Go(func() error {
		queryResult := []struct {
			NodeAddress   []byte `db:"node_address"`
			MinipoolCount uint64 `db:"minipool_count"`
		}{}

		err := d.alloyReader.SelectContext(ctx, &queryResult, `
			SELECT
				node_address,
				COUNT(node_address) AS minipool_count
			FROM rocketpool_minipools
			GROUP BY node_address`)
		if err != nil {
			return fmt.Errorf("error retrieving rocketpool node deposits data: %w", err)
		}

		for _, res := range queryResult {
			node := hexutil.Encode(res.NodeAddress)
			nodeMinipoolCount[node] = res.MinipoolCount
		}

		return nil
	})

	err := wg.Wait()
	if err != nil {
		return nil, err
	}

	if len(queryResult) == 0 {
		return nil, nil
	}

	rpInfo := t.RPInfo{
		Minipool: make(map[uint64]t.RPMinipoolInfo),
		// Smoothing pool address is the same for all nodes on the network so take the first result
		SmoothingPoolAddress: queryResult[0].SmoothingPoolAddress,
	}

	for _, res := range queryResult {
		if _, ok := rpInfo.Minipool[res.ValidatorIndex]; !ok {
			rpInfo.Minipool[res.ValidatorIndex] = t.RPMinipoolInfo{
				NodeFee:              res.NodeFee,
				NodeDepositBalance:   res.NodeDepositBalance,
				UserDepositBalance:   res.UserDepositBalance,
				SmoothingPoolRewards: make(map[uint64]decimal.Decimal),
			}
		}

		node := hexutil.Encode(res.NodeAddress)
		if res.EndTime.Valid && res.SmoothingPoolEth != nil {
			epoch := uint64(utils.TimeToEpoch(res.EndTime.Time))
			splitReward := res.SmoothingPoolEth.Div(decimal.NewFromUint64(nodeMinipoolCount[node]))
			rpInfo.Minipool[res.ValidatorIndex].SmoothingPoolRewards[epoch] =
				rpInfo.Minipool[res.ValidatorIndex].SmoothingPoolRewards[epoch].Add(splitReward)
		}
	}

	return &rpInfo, nil
}

func (d *DataAccessService) getRocketPoolOperatorFactor(minipool t.RPMinipoolInfo) decimal.Decimal {
	fullDeposit := minipool.UserDepositBalance.Add(minipool.NodeDepositBalance)
	operatorShare := minipool.NodeDepositBalance.Div(fullDeposit)
	invOperatorShare := decimal.NewFromInt(1).Sub(operatorShare)

	commissionReward := invOperatorShare.Mul(decimal.NewFromFloat(minipool.NodeFee))
	operatorFactor := operatorShare.Add(commissionReward)

	return operatorFactor
}
