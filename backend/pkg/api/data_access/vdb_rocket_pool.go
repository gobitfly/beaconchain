package dataaccess

import (
	"context"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

func (d *DataAccessService) GetValidatorDashboardRocketPool(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRocketPoolColumn], search string, limit uint64) ([]t.VDBRocketPoolTableRow, *t.Paging, error) {
	// Initialize the cursor
	var currentCursor t.RocketPoolCursor
	var err error
	var paging t.Paging

	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.RocketPoolCursor](cursor)
		if err != nil {
			return nil, &paging, fmt.Errorf("failed to parse passed cursor as RocketPoolCursor: %w", err)
		}
	}

	type RocketPoolData struct {
		Node                   []byte          `db:"naddress"`
		StakedETH              decimal.Decimal `db:"staked_eth"`
		StakedRPL              decimal.Decimal `db:"rpl_stake"`
		MinipoolsTotal         uint64          `db:"minipools_total"`
		MinipoolsLeb16         uint64          `db:"minipools_leb16"`
		MinipoolsLeb8          uint64          `db:"minipools_leb8"`
		AvgCommission          float64         `db:"avg_commission"`
		RPLClaimed             decimal.Decimal `db:"rpl_claimed"`
		RPLUnclaimed           decimal.Decimal `db:"rpl_unclaimed"`
		EffectiveRPL           decimal.Decimal `db:"effective_rpl"`
		SmoothingPoolOptIn     bool            `db:"smoothing_pool_opted_in"`
		SmoothingPoolClaimed   decimal.Decimal `db:"claimed_smoothing_pool"`
		SmoothingPoolUnclaimed decimal.Decimal `db:"unclaimed_smoothing_pool"`
		Timezone               string          `db:"timezone_location"`
		RefundBalance          decimal.Decimal `db:"refund_balance"`
		DepositCredit          decimal.Decimal `db:"deposit_credit"`
		RPLStakeMin            decimal.Decimal `db:"min_rpl_stake"`
		RPLStakeMax            decimal.Decimal `db:"max_rpl_stake"`
		NodeDepositBalance     decimal.Decimal `db:"node_deposit_balance"`
		UserDepositBalance     decimal.Decimal `db:"user_deposit_balance"`
	}

	rocketPoolResults := []RocketPoolData{}

	ds := goqu.Dialect("postgres").
		From(goqu.T("rocketpool_minipools").As("mp")).
		Select(
			goqu.T("n").Col("address").As("naddress"),
			goqu.L("SUM(mp.node_deposit_balance) AS staked_eth"),
			goqu.T("n").Col("rpl_stake"),
			goqu.L("COUNT(mp.address) as minipools_total"),
			goqu.L("SUM(CASE WHEN mp.status = 'Staking' THEN mp.node_deposit_balance ELSE 0 END) AS node_deposit_balance"),
			goqu.L("SUM(CASE WHEN mp.status = 'Staking' THEN mp.user_deposit_balance ELSE 0 END) AS user_deposit_balance"),
			goqu.L("SUM(CASE WHEN mp.node_deposit_balance = 8e18 THEN 1 ELSE 0 END) AS minipools_leb8"),
			goqu.L("SUM(CASE WHEN mp.node_deposit_balance = 16e18 THEN 1 ELSE 0 END) AS minipools_leb16"),
			goqu.L(`
				CASE 
					WHEN SUM(CASE WHEN mp.status = 'Staking' THEN mp.user_deposit_balance ELSE 0 END) > 0 
					THEN 
						SUM(CASE WHEN mp.status = 'Staking' THEN mp.node_fee * mp.user_deposit_balance ELSE 0 END) / 
						SUM(CASE WHEN mp.status = 'Staking' THEN mp.user_deposit_balance ELSE 0 END) 
					ELSE 0 
				END AS avg_commission
			`),
			goqu.T("n").Col("unclaimed_rpl_rewards").As("rpl_unclaimed"),
			goqu.L("n.rpl_cumulative_rewards - n.unclaimed_rpl_rewards as rpl_claimed"),
			goqu.T("n").Col("effective_rpl_stake").As("effective_rpl"),
			goqu.T("n").Col("smoothing_pool_opted_in"),
			goqu.T("n").Col("claimed_smoothing_pool"),
			goqu.T("n").Col("unclaimed_smoothing_pool"),
			goqu.T("n").Col("timezone_location"),
			goqu.L("COALESCE(SUM(mp.node_refund_balance),0) AS refund_balance"),
			goqu.T("n").Col("deposit_credit"),
			goqu.L("min_rpl_stake"),
			goqu.L("max_rpl_stake"),
		).
		LeftJoin(
			goqu.T("rocketpool_nodes").As("n"),
			goqu.On(goqu.T("mp").Col("node_address").Eq(goqu.T("n").Col("address"))),
		)

	if len(dashboardId.Validators) > 0 {
		ds = ds.Where(goqu.T("mp").Col("validator_index").In(dashboardId.Validators))
	} else {
		ds = ds.
			InnerJoin(
				goqu.T("users_val_dashboards_validators").As("v"),
				goqu.On(goqu.T("mp").Col("validator_index").Eq(goqu.T("v").Col("validator_index"))),
			).
			Where(goqu.T("v").Col("dashboard_id").Eq(dashboardId.Id))
	}

	if search != "" {
		bytes, err := hex.DecodeString(strings.TrimPrefix(search, "0x"))
		if err == nil {
			ds = ds.Where(goqu.L("n.address = ?", bytes))
		}
	}

	ds = ds.GroupBy(goqu.L("n.address"), goqu.L("n.rpl_stake"), goqu.L("n.unclaimed_rpl_rewards"), goqu.L("n.rpl_cumulative_rewards"), goqu.L("n.effective_rpl_stake"), goqu.L("n.smoothing_pool_opted_in"), goqu.L("n.claimed_smoothing_pool"), goqu.L("n.unclaimed_smoothing_pool"), goqu.L("n.timezone_location"), goqu.L("n.deposit_credit"), goqu.L("min_rpl_stake"), goqu.L("max_rpl_stake"))

	// 3. Sorting and pagination
	defaultColumns := []t.SortColumn{
		{Column: enums.VDBRocketPoolColumns.Node.ToExpr(), Desc: colSort.Desc, Offset: currentCursor.Address},
	}
	var offset any
	switch colSort.Column {
	case enums.VDBRocketPoolNode:
		offset = currentCursor.Address
	case enums.VDBRocketPoolCollateral:
		offset = currentCursor.StakedRpl
	case enums.VDBRocketPoolEffectiveRpl:
		offset = currentCursor.EffectiveRpl
	case enums.VDBRocketPoolSmoothingPool:
		offset = currentCursor.SmoothingpoolOptIn
	case enums.VDBRocketPoolMinipools:
		offset = currentCursor.MinipoolsTotal
	}

	order, directions, err := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToExpr(), Desc: colSort.Desc, Offset: offset}, currentCursor.GenericCursor)
	if err != nil {
		return nil, nil, err
	}
	ds = ds.Order(order...)
	if directions != nil {
		if colSort.Column == enums.VDBRocketPoolMinipools { // minipools is an aggregate hence use having instead of where
			ds = ds.Having(directions)
		} else {
			ds = ds.Where(directions)
		}
	}

	ds = ds.Limit(uint(limit + 1))

	wg := errgroup.Group{}

	var rpNetworkStats *t.RPNetworkStats

	wg.Go(func() error {
		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.alloyReader.SelectContext(ctx, &rocketPoolResults, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving rocketpool data: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		var err error
		rpNetworkStats, err = d.getInternalRpNetworkStats(ctx)
		if err != nil {
			return fmt.Errorf("error retrieving rocketpool network stats: %w", err)
		}
		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, &paging, err
	}

	apr := func(node RocketPoolData) float64 {
		if !rpNetworkStats.EffectiveRPLStaked.IsZero() && !node.EffectiveRPL.IsZero() && !rpNetworkStats.NodeOperatorRewards.IsZero() && rpNetworkStats.ClaimIntervalHours > 0 {
			share := node.EffectiveRPL.Div(rpNetworkStats.EffectiveRPLStaked)

			periodsPerYear := decimal.NewFromFloat(365 / (rpNetworkStats.ClaimIntervalHours / 24))
			return rpNetworkStats.NodeOperatorRewards.
				Mul(share).
				Div(node.StakedRPL).
				Mul(periodsPerYear).
				Mul(decimal.NewFromInt(100)).InexactFloat64()
		}
		return 0
	}

	projectedRplClaim := func(node RocketPoolData) decimal.Decimal {
		if !rpNetworkStats.EffectiveRPLStaked.IsZero() && !node.EffectiveRPL.IsZero() && !rpNetworkStats.NodeOperatorRewards.IsZero() {
			share := node.EffectiveRPL.Div(rpNetworkStats.EffectiveRPLStaked)
			return rpNetworkStats.NodeOperatorRewards.Mul(share).Floor()
		}
		return decimal.Zero
	}

	collPercentage := func(node RocketPoolData) float64 {
		if !node.StakedRPL.IsZero() && !node.RPLStakeMin.IsZero() {
			rplPrice := rpNetworkStats.RPLPrice
			currentETH := node.StakedRPL.Mul(rplPrice)
			minETH := node.RPLStakeMin.Mul(rplPrice).Mul(decimal.NewFromInt(10))

			return currentETH.Div(minETH).Mul(decimal.NewFromInt(100)).InexactFloat64()
		}
		return 0
	}

	var result []t.VDBRocketPoolTableRow
	for _, row := range rocketPoolResults {
		result = append(result, t.VDBRocketPoolTableRow{
			Address:   row.Node,
			Node:      t.Address{Hash: t.Hash(fmt.Sprintf("0x%x", row.Node))},
			StakedEth: row.StakedETH,
			StakedRpl: row.StakedRPL,

			MinipoolsTotal: row.MinipoolsTotal,
			MinipoolsLeb16: row.MinipoolsLeb16,
			MinipoolsLeb8:  row.MinipoolsLeb8,

			Collateral: t.PercentageDetails[decimal.Decimal]{
				Percentage: collPercentage(row),
				MinValue:   row.RPLStakeMin,
				MaxValue:   row.RPLStakeMax,
			},
			AvgCommission:  row.AvgCommission,
			RplClaimed:     row.RPLClaimed,
			RplUnclaimed:   row.RPLUnclaimed,
			EffectiveRpl:   row.EffectiveRPL,
			RplApr:         apr(row),
			RplAprUpdateTs: rpNetworkStats.Ts.Unix(),
			RplEstimate:    projectedRplClaim(row),

			SmoothingpoolOptIn:     row.SmoothingPoolOptIn,
			SmoothingpoolClaimed:   row.SmoothingPoolClaimed,
			SmoothingpoolUnclaimed: row.SmoothingPoolUnclaimed,
			NodeDepositBalance:     row.NodeDepositBalance,
			UserDepositBalance:     row.UserDepositBalance,

			Timezone:      row.Timezone,
			RefundBalance: row.RefundBalance,
			DepositCredit: row.DepositCredit,
		})
	}

	moreDataFlag := len(result) > int(limit)
	if moreDataFlag {
		result = result[:len(result)-1]
	}
	if currentCursor.IsReverse() {
		slices.Reverse(result)
	}

	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return result, &t.Paging{}, nil
	}
	p, err := utils.GetPagingFromData(result, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, err
	}
	return result, p, nil
}

func (d *DataAccessService) GetValidatorDashboardTotalRocketPool(ctx context.Context, dashboardId t.VDBId, search string) (*t.VDBRocketPoolTableRow, error) {
	// TODO @DATA-ACCESS
	return d.dummy.GetValidatorDashboardTotalRocketPool(ctx, dashboardId, search)
}

func (d *DataAccessService) GetValidatorDashboardRocketPoolMinipools(ctx context.Context, dashboardId t.VDBId, node, cursor string, colSort t.Sort[enums.VDBRocketPoolMinipoolsColumn], search string, limit uint64) ([]t.VDBRocketPoolMinipoolsTableRow, *t.Paging, error) {
	// TODO @DATA-ACCESS
	return d.dummy.GetValidatorDashboardRocketPoolMinipools(ctx, dashboardId, node, cursor, colSort, search, limit)
}
