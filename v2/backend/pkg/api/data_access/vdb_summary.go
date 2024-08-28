package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/juliangruber/go-intersect"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

func (d *DataAccessService) GetValidatorDashboardSummary(ctx context.Context, dashboardId t.VDBId, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	// @DATA-ACCESS incorporate protocolModes
	result := make([]t.VDBSummaryTableRow, 0)
	var paging t.Paging

	wg := errgroup.Group{}

	// Get the table name based on the period
	clickhouseTable, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, nil, err
	}

	// Searching for a group name is not supported when aggregating groups or for guest dashboards
	groupNameSearchEnabled := !dashboardId.AggregateGroups && dashboardId.Validators == nil

	// Analyze the search term
	searchValidator := -1
	if search != "" {
		if strings.HasPrefix(search, "0x") && utils.IsHash(search) {
			search = strings.ToLower(search)

			// Get the current validator state to convert pubkey to index
			validatorMapping, releaseLock, err := d.services.GetCurrentValidatorMapping()
			if err != nil {
				releaseLock()
				return nil, nil, err
			}
			if index, ok := validatorMapping.ValidatorIndices[search]; ok {
				searchValidator = int(index)
			} else {
				// No validator index for pubkey found, return empty results
				releaseLock()
				return result, &paging, nil
			}
			releaseLock()
		} else if number, err := strconv.ParseUint(search, 10, 64); err == nil {
			searchValidator = int(number)
		} else if !groupNameSearchEnabled {
			return result, &paging, nil
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Fill the validators list if we have a guest dashboard
	validators := make([]t.VDBValidator, 0)
	if dashboardId.Validators != nil {
		validatorFound := false
		for _, validator := range dashboardId.Validators {
			if searchValidator != -1 && int(validator) == searchValidator {
				validatorFound = true
			}
			validators = append(validators, validator)
		}
		if searchValidator != -1 && !validatorFound {
			// The searched validator is not part of the dashboard
			return result, &paging, nil
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Get the average network efficiency
	efficiency, releaseEfficiencyLock, err := d.services.GetCurrentEfficiencyInfo()
	if err != nil {
		releaseEfficiencyLock()
		return nil, nil, err
	}
	averageNetworkEfficiency := d.calculateTotalEfficiency(
		efficiency.AttestationEfficiency[period], efficiency.ProposalEfficiency[period], efficiency.SyncEfficiency[period])
	releaseEfficiencyLock()

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main query and get the data
	var queryResult []struct {
		GroupId                int64           `db:"result_group_id"`
		GroupName              string          `db:"group_name"`
		ValidatorIndices       []uint64        `db:"validator_indices"`
		ClRewards              int64           `db:"cl_rewards"`
		AttestationReward      decimal.Decimal `db:"attestations_reward"`
		AttestationIdealReward decimal.Decimal `db:"attestations_ideal_reward"`
		AttestationsExecuted   uint64          `db:"attestations_executed"`
		AttestationsScheduled  uint64          `db:"attestations_scheduled"`
		BlocksProposed         uint64          `db:"blocks_proposed"`
		BlocksScheduled        uint64          `db:"blocks_scheduled"`
		SyncExecuted           uint64          `db:"sync_executed"`
		SyncScheduled          uint64          `db:"sync_scheduled"`
		MinEpochStart          int64           `db:"min_epoch_start"`
		MaxEpochEnd            int64           `db:"max_epoch_end"`
	}

	ds := goqu.Dialect("postgres").
		From(goqu.L(fmt.Sprintf(`%s AS r FINAL`, clickhouseTable))).
		With("validators", goqu.L("(SELECT dashboard_id, group_id, validator_index FROM users_val_dashboards_validators WHERE dashboard_id = ?)", dashboardId.Id)).
		Select(
			goqu.L("ARRAY_AGG(r.validator_index) AS validator_indices"),
			goqu.L("(SUM(COALESCE(r.balance_end,0)) + SUM(COALESCE(r.withdrawals_amount,0)) - SUM(COALESCE(r.deposits_amount,0)) - SUM(COALESCE(r.balance_start,0))) AS cl_rewards"),
			goqu.L("COALESCE(SUM(r.attestations_reward)::decimal, 0) AS attestations_reward"),
			goqu.L("COALESCE(SUM(r.attestations_ideal_reward)::decimal, 0) AS attestations_ideal_reward"),
			goqu.L("COALESCE(SUM(r.attestations_executed), 0) AS attestations_executed"),
			goqu.L("COALESCE(SUM(r.attestations_scheduled), 0) AS attestations_scheduled"),
			goqu.L("COALESCE(SUM(r.blocks_proposed), 0) AS blocks_proposed"),
			goqu.L("COALESCE(SUM(r.blocks_scheduled), 0) AS blocks_scheduled"),
			goqu.L("COALESCE(SUM(r.sync_executed), 0) AS sync_executed"),
			goqu.L("COALESCE(SUM(r.sync_scheduled), 0) AS sync_scheduled"),
			goqu.L("COALESCE(MIN(r.epoch_start), 0) AS min_epoch_start"),
			goqu.L("COALESCE(MAX(r.epoch_end), 0) AS max_epoch_end")).
		GroupBy(goqu.L("result_group_id"))

	if len(validators) > 0 {
		ds = ds.
			SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
			Where(goqu.L("r.validator_index IN ?", validators))
	} else {
		if dashboardId.AggregateGroups {
			ds = ds.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId))
		} else {
			ds = ds.
				SelectAppend(goqu.L("v.group_id AS result_group_id"))
		}

		ds = ds.
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
			Where(goqu.L("r.validator_index IN (SELECT validator_index FROM validators)"))

		if groupNameSearchEnabled && (search != "" || colSort.Column == enums.VDBSummaryColumns.Group) {
			// Get the group names since we can filter and/or sort for them
			ds = ds.
				SelectAppend(goqu.L("g.name AS group_name")).
				InnerJoin(goqu.L("users_val_dashboards_groups g"), goqu.On(goqu.L("v.group_id = g.id AND v.dashboard_id = g.dashboard_id"))).
				GroupByAppend(goqu.L("group_name"))
		}
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, nil, fmt.Errorf("error preparing query: %v", err)
	}

	err = d.clickhouseReader.SelectContext(ctx, &queryResult, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving data from table %s: %v", clickhouseTable, err)
	}

	if len(queryResult) == 0 {
		// No groups to show
		return result, &paging, nil
	}

	epochMin := int64(math.MaxInt32)
	epochMax := int64(0)

	for _, row := range queryResult {
		if row.MinEpochStart < epochMin {
			epochMin = row.MinEpochStart
		}
		if row.MaxEpochEnd > epochMax {
			epochMax = row.MaxEpochEnd
		}
	}
	// ------------------------------------------------------------------------------------------------------------------
	// Get the EL rewards
	elRewards := make(map[int64]decimal.Decimal)
	ds = goqu.Dialect("postgres").
		Select(
			goqu.L("SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)) AS el_rewards")).
		From(goqu.L("blocks b")).
		LeftJoin(goqu.L("execution_payloads ep"), goqu.On(goqu.L("ep.block_hash = b.exec_block_hash"))).
		LeftJoin(
			goqu.Dialect("postgres").
				From("relays_blocks").
				Select(
					goqu.L("exec_block_hash"),
					goqu.MAX("value")).As("value").
				GroupBy("exec_block_hash").As("rb"),
			goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash")),
		).
		Where(goqu.L("b.epoch >= ? AND b.epoch <= ? AND b.status = '1'", epochMin, epochMax)).
		GroupBy(goqu.L("result_group_id"))

	if len(validators) > 0 {
		ds = ds.
			SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
			Where(goqu.L("b.proposer = ANY(?)", pq.Array(validators)))
	} else {
		if dashboardId.AggregateGroups {
			ds = ds.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId))
		} else {
			ds = ds.
				SelectAppend(goqu.L("v.group_id AS result_group_id"))
		}

		ds = ds.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("b.proposer = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))
	}

	var elRewardsQueryResult []struct {
		GroupId   int64           `db:"result_group_id"`
		ElRewards decimal.Decimal `db:"el_rewards"`
	}

	query, args, err = ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, nil, fmt.Errorf("error preparing query: %v", err)
	}

	err = d.alloyReader.SelectContext(ctx, &elRewardsQueryResult, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving data from table blocks: %v", err)
	}

	for _, entry := range elRewardsQueryResult {
		elRewards[entry.GroupId] = entry.ElRewards
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Get the current and next sync committee validators
	latestEpoch := cache.LatestEpoch.Get()
	currentSyncCommitteeValidators := make(map[uint64]bool)
	upcomingSyncCommitteeValidators := make(map[uint64]bool)
	wg.Go(func() error {
		var err error
		currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, err = d.getCurrentAndUpcomingSyncCommittees(ctx, latestEpoch)
		return err
	})

	err = wg.Wait()
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving validator dashboard summary data: %v", err)
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Sort by group name, after this the name is no longer relevant
	if groupNameSearchEnabled && colSort.Column == enums.VDBSummaryColumns.Group {
		sort.Slice(queryResult, func(i, j int) bool {
			if colSort.Desc {
				return queryResult[i].GroupName > queryResult[j].GroupName
			} else {
				return queryResult[i].GroupName < queryResult[j].GroupName
			}
		})
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Calculate the result
	total := struct {
		GroupId                int64
		Status                 t.VDBSummaryStatus
		Validators             t.VDBSummaryValidators
		AttestationReward      decimal.Decimal
		AttestationIdealReward decimal.Decimal
		AttestationsExecuted   uint64
		AttestationsScheduled  uint64
		BlocksProposed         uint64
		BlocksScheduled        uint64
		SyncExecuted           uint64
		SyncScheduled          uint64
		Reward                 t.ClElValue[decimal.Decimal]
	}{
		GroupId: t.AllGroups,
	}

	for _, queryEntry := range queryResult {
		resultEntry := t.VDBSummaryTableRow{
			GroupId:                  queryEntry.GroupId,
			AverageNetworkEfficiency: averageNetworkEfficiency,
		}

		// Status
		for _, validatorIndex := range queryEntry.ValidatorIndices {
			if currentSyncCommitteeValidators[validatorIndex] {
				resultEntry.Status.CurrentSyncCount++
			}
			if upcomingSyncCommitteeValidators[validatorIndex] {
				resultEntry.Status.UpcomingSyncCount++
			}
		}
		total.Status.CurrentSyncCount += resultEntry.Status.CurrentSyncCount
		total.Status.UpcomingSyncCount += resultEntry.Status.UpcomingSyncCount

		// Validator statuses
		validatorMapping, releaseValMapLock, err := d.services.GetCurrentValidatorMapping()
		defer releaseValMapLock()
		if err != nil {
			return nil, nil, err
		}

		for _, validator := range queryEntry.ValidatorIndices {
			metadata := validatorMapping.ValidatorMetadata[validator]

			// As deposited and pending validators are neither online nor offline they are counted as the third state (exited)
			switch constypes.ValidatorDbStatus(metadata.Status) {
			case constypes.DbDeposited:
				resultEntry.Validators.Exited++
			case constypes.DbPending:
				resultEntry.Validators.Exited++
			case constypes.DbActiveOnline, constypes.DbExitingOnline, constypes.DbSlashingOnline:
				resultEntry.Validators.Online++
			case constypes.DbActiveOffline, constypes.DbExitingOffline, constypes.DbSlashingOffline:
				resultEntry.Validators.Offline++
			case constypes.DbSlashed:
				resultEntry.Validators.Exited++
				resultEntry.Status.SlashedCount++
			case constypes.DbExited:
				resultEntry.Validators.Exited++
			}
		}
		total.Validators.Online += resultEntry.Validators.Online
		total.Validators.Offline += resultEntry.Validators.Offline
		total.Validators.Exited += resultEntry.Validators.Exited
		total.Status.SlashedCount += resultEntry.Status.SlashedCount

		// Attestations
		resultEntry.Attestations.Success = queryEntry.AttestationsExecuted
		resultEntry.Attestations.Failed = queryEntry.AttestationsScheduled - queryEntry.AttestationsExecuted

		// Proposals
		resultEntry.Proposals.Success = queryEntry.BlocksProposed
		resultEntry.Proposals.Failed = queryEntry.BlocksScheduled - queryEntry.BlocksProposed

		// Rewards
		resultEntry.Reward.Cl = utils.GWeiToWei(big.NewInt(queryEntry.ClRewards))
		if _, ok := elRewards[queryEntry.GroupId]; ok {
			resultEntry.Reward.El = elRewards[queryEntry.GroupId]
		}
		total.Reward.Cl = total.Reward.Cl.Add(resultEntry.Reward.Cl)
		total.Reward.El = total.Reward.El.Add(resultEntry.Reward.El)

		// Efficiency
		var attestationEfficiency, proposerEfficiency, syncEfficiency sql.NullFloat64
		if !queryEntry.AttestationIdealReward.IsZero() {
			attestationEfficiency.Float64 = queryEntry.AttestationReward.Div(queryEntry.AttestationIdealReward).InexactFloat64()
			attestationEfficiency.Valid = true
		}
		if queryEntry.BlocksScheduled > 0 {
			proposerEfficiency.Float64 = float64(queryEntry.BlocksProposed) / float64(queryEntry.BlocksScheduled)
			proposerEfficiency.Valid = true
		}
		if queryEntry.SyncScheduled > 0 {
			syncEfficiency.Float64 = float64(queryEntry.SyncExecuted) / float64(queryEntry.SyncScheduled)
			syncEfficiency.Valid = true
		}
		resultEntry.Efficiency = d.calculateTotalEfficiency(attestationEfficiency, proposerEfficiency, syncEfficiency)

		// Add the duties info to the total
		total.AttestationReward = total.AttestationReward.Add(queryEntry.AttestationReward)
		total.AttestationIdealReward = total.AttestationIdealReward.Add(queryEntry.AttestationIdealReward)
		total.AttestationsExecuted += queryEntry.AttestationsExecuted
		total.AttestationsScheduled += queryEntry.AttestationsScheduled
		total.BlocksProposed += queryEntry.BlocksProposed
		total.BlocksScheduled += queryEntry.BlocksScheduled
		total.SyncExecuted += queryEntry.SyncExecuted
		total.SyncScheduled += queryEntry.SyncScheduled

		// If the search permits it add the entry to the result
		if search != "" {
			prefixSearch := strings.ToLower(search)
			for _, validatorIndex := range queryEntry.ValidatorIndices {
				if searchValidator != -1 && validatorIndex == uint64(searchValidator) ||
					(groupNameSearchEnabled && strings.HasPrefix(strings.ToLower(queryEntry.GroupName), prefixSearch)) {
					result = append(result, resultEntry)
					break
				}
			}
		} else {
			result = append(result, resultEntry)
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Sort the result
	// For sorting consider a 0/0 => 0% as lower than a 0/5 => 0%
	var sortParam func(resultEntry t.VDBSummaryTableRow) float64
	switch colSort.Column {
	case enums.VDBSummaryColumns.Validators:
		sortParam = func(resultEntry t.VDBSummaryTableRow) float64 {
			divisor := float64(resultEntry.Validators.Online + resultEntry.Validators.Offline)
			if divisor == 0 {
				return -1
			}
			return float64(resultEntry.Validators.Online) / divisor
		}
	case enums.VDBSummaryColumns.Efficiency:
		sortParam = func(resultEntry t.VDBSummaryTableRow) float64 {
			return resultEntry.Efficiency
		}
	case enums.VDBSummaryColumns.Attestations:
		sortParam = func(resultEntry t.VDBSummaryTableRow) float64 {
			divisor := float64(resultEntry.Attestations.Success + resultEntry.Attestations.Failed)
			if divisor == 0 {
				return -1
			}
			return float64(resultEntry.Attestations.Success) / divisor
		}
	case enums.VDBSummaryColumns.Proposals:
		sortParam = func(resultEntry t.VDBSummaryTableRow) float64 {
			divisor := float64(resultEntry.Proposals.Success + resultEntry.Proposals.Failed)
			if divisor == 0 {
				return -1
			}
			return float64(resultEntry.Proposals.Success) / divisor
		}
	case enums.VDBSummaryColumns.Reward:
		rewardSortParam := func(resultEntry t.VDBSummaryTableRow) decimal.Decimal {
			return resultEntry.Reward.Cl.Add(resultEntry.Reward.El)
		}
		sort.Slice(result, func(i, j int) bool {
			if colSort.Desc {
				return rewardSortParam(result[i]).GreaterThan(rewardSortParam(result[j]))
			} else {
				return rewardSortParam(result[i]).LessThan(rewardSortParam(result[j]))
			}
		})
	case enums.VDBSummaryColumns.Group:
	default:
		return nil, nil, fmt.Errorf("error sorting data: unexpected sorting type")
	}

	if sortParam != nil {
		sort.Slice(result, func(i, j int) bool {
			if colSort.Desc {
				return sortParam(result[i]) > sortParam(result[j])
			} else {
				return sortParam(result[i]) < sortParam(result[j])
			}
		})
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Calculate the total
	if len(queryResult) > 1 && len(result) > 0 {
		// We have more than one group and at least one group remains after the filtering so we need to show the total row
		totalEntry := t.VDBSummaryTableRow{
			GroupId:                  total.GroupId,
			Status:                   total.Status,
			Validators:               total.Validators,
			AverageNetworkEfficiency: averageNetworkEfficiency,
			Reward:                   total.Reward,
		}

		// Attestations
		totalEntry.Attestations.Success = total.AttestationsExecuted
		totalEntry.Attestations.Failed = total.AttestationsScheduled - total.AttestationsExecuted

		// Proposals
		totalEntry.Proposals.Success = total.BlocksProposed
		totalEntry.Proposals.Failed = total.BlocksScheduled - total.BlocksProposed

		// Efficiency
		var totalAttestationEfficiency, totalProposerEfficiency, totalSyncEfficiency sql.NullFloat64
		if !total.AttestationIdealReward.IsZero() {
			totalAttestationEfficiency.Float64 = total.AttestationReward.Div(total.AttestationIdealReward).InexactFloat64()
			totalAttestationEfficiency.Valid = true
		}
		if total.BlocksScheduled > 0 {
			totalProposerEfficiency.Float64 = float64(total.BlocksProposed) / float64(total.BlocksScheduled)
			totalProposerEfficiency.Valid = true
		}
		if total.SyncScheduled > 0 {
			totalSyncEfficiency.Float64 = float64(total.SyncExecuted) / float64(total.SyncScheduled)
			totalSyncEfficiency.Valid = true
		}
		totalEntry.Efficiency = d.calculateTotalEfficiency(totalAttestationEfficiency, totalProposerEfficiency, totalSyncEfficiency)

		result = append([]t.VDBSummaryTableRow{totalEntry}, result...)
	}

	paging.TotalCount = uint64(len(result))

	return result, &paging, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupSummary(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod, protocolModes t.VDBProtocolModes) (*t.VDBGroupSummaryData, error) {
	// TODO: implement data retrieval for the following new field
	// Fetch validator list for user dashboard from the dashboard table when querying the past sync committees as the rolling table might miss exited validators
	// TotalMissedRewards
	// @DATA-ACCESS incorporate protocolModes
	// @DATA-ACCESS implement data retrieval for Rocket Pool stats (if present)

	var err error
	ret := &t.VDBGroupSummaryData{}

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	// Get the current and next sync committee validators
	latestEpoch := cache.LatestEpoch.Get()
	currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, err := d.getCurrentAndUpcomingSyncCommittees(ctx, latestEpoch)
	if err != nil {
		return nil, err
	}

	// Get the table names based on the period
	clickhouseTable, hours, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, err
	}

	validators := make([]t.VDBValidator, 0)
	if dashboardId.Validators != nil {
		validators = dashboardId.Validators
	}

	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("validator_index"),
			goqu.L("epoch_start"),
			goqu.L("attestations_reward"),
			goqu.L("attestations_ideal_reward"),
			goqu.L("attestations_scheduled"),
			goqu.L("attestations_executed"),
			goqu.L("attestation_head_executed"),
			goqu.L("attestation_source_executed"),
			goqu.L("attestation_target_executed"),
			goqu.L("blocks_scheduled"),
			goqu.L("blocks_proposed"),
			goqu.L("sync_scheduled"),
			goqu.L("sync_executed"),
			goqu.L("slashed AS slashed_in_period"),
			goqu.L("blocks_slashing_count AS slashed_amount"),
			goqu.L("blocks_expected"),
			goqu.L("inclusion_delay_sum"),
			goqu.L("sync_committees_expected")).
		From(goqu.L(fmt.Sprintf(`%s AS r FINAL`, clickhouseTable)))

	if dashboardId.Validators == nil {
		ds = ds.
			With("validators", goqu.L("(SELECT validator_index as validator_index, group_id FROM users_val_dashboards_validators WHERE dashboard_id = ? AND (group_id = ? OR ?::smallint = -1))", dashboardId.Id, groupId, groupId)).
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
			Where(goqu.L("validator_index IN (SELECT validator_index FROM validators)"))
	} else {
		ds = ds.
			Where(goqu.L("validator_index IN ?", validators))
	}

	type QueryResult struct {
		ValidatorIndex         uint32 `db:"validator_index"`
		EpochStart             uint64 `db:"epoch_start"`
		AttestationReward      int64  `db:"attestations_reward"`
		AttestationIdealReward int64  `db:"attestations_ideal_reward"`

		AttestationsScheduled     int64 `db:"attestations_scheduled"`
		AttestationsExecuted      int64 `db:"attestations_executed"`
		AttestationHeadExecuted   int64 `db:"attestation_head_executed"`
		AttestationSourceExecuted int64 `db:"attestation_source_executed"`
		AttestationTargetExecuted int64 `db:"attestation_target_executed"`

		BlocksScheduled uint32 `db:"blocks_scheduled"`
		BlocksProposed  uint32 `db:"blocks_proposed"`

		SyncScheduled uint32 `db:"sync_scheduled"`
		SyncExecuted  uint32 `db:"sync_executed"`

		SlashedInPeriod bool   `db:"slashed_in_period"`
		SlashedAmount   uint32 `db:"slashed_amount"`

		BlockChance            float64 `db:"blocks_expected"`
		SyncCommitteesExpected float64 `db:"sync_committees_expected"`

		InclusionDelaySum int64 `db:"inclusion_delay_sum"`
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	var rows []*QueryResult
	err = d.clickhouseReader.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard group summary data: %v", err)
	}

	if len(rows) == 0 {
		return ret, nil
	}

	totalAttestationRewards := int64(0)
	totalIdealAttestationRewards := int64(0)
	totalBlockChance := float64(0)
	totalInclusionDelaySum := int64(0)
	totalInclusionDelayDivisor := int64(0)
	totalSyncExpected := float64(0)
	totalProposals := uint32(0)

	validatorArr := make([]t.VDBValidator, 0)
	for _, row := range rows {
		validatorArr = append(validatorArr, t.VDBValidator(row.ValidatorIndex))
		totalAttestationRewards += row.AttestationReward
		totalIdealAttestationRewards += row.AttestationIdealReward

		ret.AttestationsHead.Success += uint64(row.AttestationHeadExecuted)
		ret.AttestationsHead.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationHeadExecuted)

		ret.AttestationsSource.Success += uint64(row.AttestationSourceExecuted)
		ret.AttestationsSource.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationSourceExecuted)

		ret.AttestationsTarget.Success += uint64(row.AttestationTargetExecuted)
		ret.AttestationsTarget.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationTargetExecuted)

		if row.ValidatorIndex == 0 && row.BlocksProposed > 0 && row.BlocksProposed != row.BlocksScheduled {
			row.BlocksProposed-- // subtract the genesis block from validator 0 (TODO: remove when fixed in the dashoard data exporter)
		}

		totalProposals += row.BlocksScheduled
		if row.BlocksScheduled > 0 {
			if ret.ProposalValidators == nil {
				ret.ProposalValidators = make([]t.VDBValidator, 0, 10)
			}
			ret.ProposalValidators = append(ret.ProposalValidators, t.VDBValidator(row.ValidatorIndex))
		}

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
			totalInclusionDelayDivisor += row.AttestationsExecuted
		}
	}

	_, ret.Apr.El, _, ret.Apr.Cl, err = d.internal_getElClAPR(ctx, dashboardId, groupId, hours)
	if err != nil {
		return nil, err
	}

	if len(validators) > 0 {
		validatorArr = validators
	}

	pastSyncPeriodCutoff := utils.SyncPeriodOfEpoch(rows[0].EpochStart)
	currentSyncPeriod := utils.SyncPeriodOfEpoch(latestEpoch)
	err = d.readerDb.GetContext(ctx, &ret.SyncCommitteeCount.PastPeriods, `SELECT COUNT(*) FROM sync_committees WHERE period >= $1 AND period < $2 AND validatorindex = ANY($3)`, pastSyncPeriodCutoff, currentSyncPeriod, validatorArr)
	if err != nil {
		return nil, fmt.Errorf("error retrieving past sync committee count: %v", err)
	}

	ret.AttestationEfficiency = float64(totalAttestationRewards) / float64(totalIdealAttestationRewards) * 100
	if ret.AttestationEfficiency < 0 || math.IsNaN(ret.AttestationEfficiency) {
		ret.AttestationEfficiency = 0
	}

	luckHours := float64(hours)
	if hours == -1 {
		luckHours = time.Since(time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)).Hours()
		if luckHours == 0 {
			luckHours = 24
		}
	}

	if totalBlockChance > 0 {
		ret.Luck.Proposal.Percent = (float64(totalProposals)) / totalBlockChance * 100

		// calculate the average time it takes for the set of validators to propose a single block on average
		ret.Luck.Proposal.Average = time.Duration((luckHours / totalBlockChance) * float64(time.Hour))
	} else {
		ret.Luck.Proposal.Percent = 0
	}

	if totalSyncExpected == 0 {
		ret.Luck.Sync.Percent = 0
	} else {
		totalSyncSlotDuties := float64(ret.SyncCommittee.StatusCount.Failed) + float64(ret.SyncCommittee.StatusCount.Success)
		slotDutiesPerSyncCommittee := float64(utils.SlotsPerSyncCommittee())
		syncCommittees := math.Ceil(totalSyncSlotDuties / slotDutiesPerSyncCommittee) // gets the number of sync committees
		ret.Luck.Sync.Percent = syncCommittees / totalSyncExpected * 100

		// calculate the average time it takes for the set of validators to be elected into a sync committee on average
		ret.Luck.Sync.Average = time.Duration((luckHours / totalSyncExpected) * float64(time.Hour))
	}

	if totalInclusionDelayDivisor > 0 {
		ret.AttestationAvgInclDist = 1.0 + float64(totalInclusionDelaySum)/float64(totalInclusionDelayDivisor)
	} else {
		ret.AttestationAvgInclDist = 0
	}

	return ret, nil
}

func (d *DataAccessService) internal_getElClAPR(ctx context.Context, dashboardId t.VDBId, groupId int64, hours int) (elIncome decimal.Decimal, elAPR float64, clIncome decimal.Decimal, clAPR float64, err error) {
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
			goqu.L("(SUM(COALESCE(r.balance_end,0)) + SUM(COALESCE(r.withdrawals_amount,0)) - SUM(COALESCE(r.deposits_amount,0)) - SUM(COALESCE(r.balance_start,0))) AS reward"))

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
		return decimal.Zero, 0, decimal.Zero, 0, fmt.Errorf("error preparing query: %v", err)
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
			return decimal.Zero, 0, decimal.Zero, 0, fmt.Errorf("error preparing query: %v", err)
		}

		err = d.clickhouseReader.GetContext(ctx, &rewardsResultTotal, query, args...)
		if err != nil || !rewardsResultTotal.Reward.Valid {
			return decimal.Zero, 0, decimal.Zero, 0, err
		}

		clIncome = decimal.NewFromInt(rewardsResultTotal.Reward.Int64).Mul(decimal.NewFromInt(1e9))
	}

	elDs := goqu.Dialect("postgres").
		Select(goqu.L("COALESCE(SUM(COALESCE(rb.value / 1e18, fee_recipient_reward)), 0) AS el_reward")).
		From(goqu.L("blocks AS b")).
		LeftJoin(goqu.L("execution_payloads AS ep"), goqu.On(goqu.L("b.exec_block_hash = ep.block_hash"))).
		LeftJoin(
			goqu.Dialect("postgres").
				From("relays_blocks").
				Select(
					goqu.L("exec_block_hash"),
					goqu.MAX("value")).As("value").
				GroupBy("exec_block_hash").As("rb"),
			goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash")),
		).
		Where(goqu.L("b.status = '1'"))

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
		return decimal.Zero, 0, decimal.Zero, 0, fmt.Errorf("error preparing query: %v", err)
	}

	err = d.alloyReader.GetContext(ctx, &elIncome, query, args...)
	if err != nil {
		return decimal.Zero, 0, decimal.Zero, 0, err
	}
	elIncomeFloat, _ := elIncome.Float64()
	elAPR = ((elIncomeFloat / float64(aprDivisor)) / (float64(32e18) * float64(rewardsResultTable.ValidatorCount))) * 24.0 * 365.0 * 100.0
	if math.IsNaN(elAPR) {
		elAPR = 0
	}

	if hours == -1 {
		elTotalDs := elDs.
			Where(goqu.L("b.epoch >= ? AND b.epoch <= ?", rewardsResultTotal.EpochStart, rewardsResultTotal.EpochEnd))

		query, args, err = elTotalDs.Prepared(true).ToSQL()
		if err != nil {
			return decimal.Zero, 0, decimal.Zero, 0, fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.GetContext(ctx, &elIncome, query, args...)
		if err != nil {
			return decimal.Zero, 0, decimal.Zero, 0, err
		}
	}
	elIncome = elIncome.Mul(decimal.NewFromInt(1e18))

	return elIncome, elAPR, clIncome, clAPR, nil
}

// for summary charts: series id is group id, no stack

func (d *DataAccessService) GetValidatorDashboardSummaryChart(ctx context.Context, dashboardId t.VDBId, groupIds []int64, efficiency enums.VDBSummaryChartEfficiencyType, aggregation enums.ChartAggregation, afterTs uint64, beforeTs uint64) (*t.ChartData[int, float64], error) {
	ret := &t.ChartData[int, float64]{}

	if len(groupIds) == 0 { // short circuit if no groups are selected
		return ret, nil
	}

	// log.Infof("retrieving data between %v and %v for aggregation %v", time.Unix(int64(afterTs), 0), time.Unix(int64(beforeTs), 0), aggregation)
	dataTable := ""
	dateColumn := ""
	switch aggregation {
	case enums.IntervalEpoch:
		dataTable = "validator_dashboard_data_epoch"
		dateColumn = "epoch_timestamp"
	case enums.IntervalHourly:
		dataTable = "validator_dashboard_data_hourly"
		dateColumn = "hour"
	case enums.IntervalDaily:
		dataTable = "validator_dashboard_data_daily"
		dateColumn = "day"
	case enums.IntervalWeekly:
		dataTable = "validator_dashboard_data_weekly"
		dateColumn = "week"
	default:
		return nil, fmt.Errorf("unexpected aggregation type: %v", aggregation)
	}

	var queryResults []*t.VDBValidatorSummaryChartRow

	containsGroups := false
	requestedGroupsMap := make(map[int64]bool)
	for _, groupId := range groupIds {
		requestedGroupsMap[groupId] = true
		if !containsGroups && groupId >= 0 {
			containsGroups = true
		}
	}

	totalLineRequested := requestedGroupsMap[t.AllGroups]
	averageNetworkLineRequested := requestedGroupsMap[t.NetworkAverage]

	if dashboardId.Validators != nil {
		query := fmt.Sprintf(`
			SELECT
				%[2]s as ts,
				0 AS group_id, 
				COALESCE(SUM(d.attestations_reward), 0) AS attestation_reward,
				COALESCE(SUM(d.attestations_ideal_reward), 0) AS attestations_ideal_reward,
				COALESCE(SUM(d.blocks_proposed), 0) AS blocks_proposed,
				COALESCE(SUM(d.blocks_scheduled), 0) AS blocks_scheduled,
				COALESCE(SUM(d.sync_executed), 0) AS sync_executed,
				COALESCE(SUM(d.sync_scheduled), 0) AS sync_scheduled
			FROM %[1]s d
			WHERE %[2]s >= fromUnixTimestamp($1) AND %[2]s <= fromUnixTimestamp($2) AND validator_index IN ($3)
			GROUP BY %[2]s;
		`, dataTable, dateColumn)
		err := d.clickhouseReader.SelectContext(ctx, &queryResults, query, afterTs, beforeTs, dashboardId.Validators)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table %s: %v", dataTable, err)
		}
	} else {
		query := fmt.Sprintf(`
		WITH validators AS (
			SELECT validator_index as validator_index, group_id FROM users_val_dashboards_validators WHERE dashboard_id = $3 AND (group_id IN ($4) OR $5)
		)		
		SELECT
			%[2]s as ts,
			v.group_id,
			COALESCE(SUM(d.attestations_reward), 0) AS attestation_reward,
			COALESCE(SUM(d.attestations_ideal_reward), 0) AS attestations_ideal_reward,
			COALESCE(SUM(d.blocks_proposed), 0) AS blocks_proposed,
			COALESCE(SUM(d.blocks_scheduled), 0) AS blocks_scheduled,
			COALESCE(SUM(d.sync_executed), 0) AS sync_executed,
			COALESCE(SUM(d.sync_scheduled), 0) AS sync_scheduled
		FROM %[1]s d
		INNER JOIN validators v ON d.validator_index = v.validator_index
		WHERE %[2]s >= fromUnixTimestamp($1) AND %[2]s <= fromUnixTimestamp($2) AND validator_index in (select validator_index from validators)
		GROUP BY 1, 2;`, dataTable, dateColumn)

		err := d.clickhouseReader.SelectContext(ctx, &queryResults, query, afterTs, beforeTs, dashboardId.Id, groupIds, totalLineRequested)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table %s: %v", dataTable, err)
		}
	}

	// convert the returned data to the expected return type (not pretty)
	tsMap := make(map[time.Time]bool)
	data := make(map[time.Time]map[int64]float64)

	totalEfficiencyMap := make(map[time.Time]*t.VDBValidatorSummaryChartRow)
	for _, row := range queryResults {
		tsMap[row.Timestamp] = true

		if data[row.Timestamp] == nil {
			data[row.Timestamp] = make(map[int64]float64)
		}

		if requestedGroupsMap[row.GroupId] {
			groupEfficiency, err := d.calculateChartEfficiency(efficiency, row)
			if err != nil {
				return nil, err
			}

			data[row.Timestamp][row.GroupId] = groupEfficiency
		}

		if totalLineRequested {
			if totalEfficiencyMap[row.Timestamp] == nil {
				totalEfficiencyMap[row.Timestamp] = &t.VDBValidatorSummaryChartRow{
					Timestamp: row.Timestamp,
				}
			}
			totalEfficiencyMap[row.Timestamp].AttestationReward += row.AttestationReward
			totalEfficiencyMap[row.Timestamp].AttestationIdealReward += row.AttestationIdealReward
			totalEfficiencyMap[row.Timestamp].BlocksProposed += row.BlocksProposed
			totalEfficiencyMap[row.Timestamp].BlocksScheduled += row.BlocksScheduled
			totalEfficiencyMap[row.Timestamp].SyncExecuted += row.SyncExecuted
			totalEfficiencyMap[row.Timestamp].SyncScheduled += row.SyncScheduled
		}
	}

	if averageNetworkLineRequested {
		// Get the average network efficiency
		efficiency, releaseEfficiencyLock, err := d.services.GetCurrentEfficiencyInfo()
		if err != nil {
			releaseEfficiencyLock()
			return nil, err
		}
		averageNetworkEfficiency := d.calculateTotalEfficiency(
			efficiency.AttestationEfficiency[enums.Last24h], efficiency.ProposalEfficiency[enums.Last24h], efficiency.SyncEfficiency[enums.Last24h])
		releaseEfficiencyLock()

		for ts := range tsMap {
			data[ts][int64(t.NetworkAverage)] = averageNetworkEfficiency
		}
	}

	if totalLineRequested {
		for _, row := range totalEfficiencyMap {
			totalEfficiency, err := d.calculateChartEfficiency(efficiency, row)
			if err != nil {
				return nil, err
			}

			data[row.Timestamp][t.AllGroups] = totalEfficiency
		}
	}

	tsArray := make([]time.Time, 0, len(tsMap))
	for ts := range tsMap {
		tsArray = append(tsArray, ts)
	}
	sort.Slice(tsArray, func(i, j int) bool {
		return tsArray[i].Before(tsArray[j])
	})

	groupsArray := make([]int64, 0, len(requestedGroupsMap))
	for group := range requestedGroupsMap {
		groupsArray = append(groupsArray, group)
	}
	sort.Slice(groupsArray, func(i, j int) bool {
		return groupsArray[i] < groupsArray[j]
	})

	ret.Categories = make([]uint64, 0, len(tsArray))
	for _, ts := range tsArray {
		ret.Categories = append(ret.Categories, uint64(ts.Unix()))
	}
	ret.Series = make([]t.ChartSeries[int, float64], 0, len(groupsArray))

	seriesMap := make(map[int64]*t.ChartSeries[int, float64])
	for group := range requestedGroupsMap {
		series := t.ChartSeries[int, float64]{
			Id:   int(group),
			Data: make([]float64, 0, len(tsMap)),
		}
		seriesMap[group] = &series
	}

	for _, ts := range tsArray {
		for _, group := range groupsArray {
			seriesMap[group].Data = append(seriesMap[group].Data, data[ts][group])
		}
	}

	for _, series := range seriesMap {
		ret.Series = append(ret.Series, *series)
	}

	sort.Slice(ret.Series, func(i, j int) bool {
		return ret.Series[i].Id < ret.Series[j].Id
	})

	return ret, nil
}

func (d *DataAccessService) GetLatestExportedChartTs(ctx context.Context, aggregation enums.ChartAggregation) (uint64, error) {
	var table string
	var dateColumn string
	switch aggregation {
	case enums.IntervalEpoch:
		table = "validator_dashboard_data_epoch"
		dateColumn = "epoch_timestamp"
	case enums.IntervalHourly:
		table = "validator_dashboard_data_hourly"
		dateColumn = "hour"
	case enums.IntervalDaily:
		table = "validator_dashboard_data_daily"
		dateColumn = "day"
	case enums.IntervalWeekly:
		table = "validator_dashboard_data_weekly"
		dateColumn = "week"
	default:
		return 0, fmt.Errorf("unexpected aggregation type: %v", aggregation)
	}

	query := fmt.Sprintf(`SELECT max(%s) FROM %s`, dateColumn, table)
	var ts time.Time
	err := d.clickhouseReader.GetContext(ctx, &ts, query)
	if err != nil {
		return 0, fmt.Errorf("error retrieving latest exported chart timestamp: %v", err)
	}

	return uint64(ts.Unix()), nil
}

func (d *DataAccessService) GetValidatorDashboardSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64) (*t.VDBGeneralSummaryValidators, error) {
	result := &t.VDBGeneralSummaryValidators{}

	// Get the validator indices
	var groupIds []uint64
	if !dashboardId.AggregateGroups && groupId != t.AllGroups {
		groupIds = append(groupIds, uint64(groupId))
	}

	validatorIndices, err := d.getDashboardValidators(ctx, dashboardId, groupIds)
	if err != nil {
		return nil, err
	}

	latestEpoch := cache.LatestFinalizedEpoch.Get()
	latestStats := cache.LatestStats.Get()
	var activationChurnRate uint64

	if latestStats.ValidatorActivationChurnLimit == nil {
		activationChurnRate = 4
		log.Warnf("Activation Churn rate not set in config using 4 as default")
	} else {
		activationChurnRate = *latestStats.ValidatorActivationChurnLimit
	}

	stats := cache.LatestStats.Get()
	if stats == nil || stats.LatestValidatorWithdrawalIndex == nil {
		return nil, errors.New("stats not available")
	}

	// Get the current validator state
	validatorMapping, releaseValMapLock, err := d.services.GetCurrentValidatorMapping()
	defer releaseValMapLock()
	if err != nil {
		return nil, err
	}

	// Fill the data
	for _, validatorIndex := range validatorIndices {
		metadata := validatorMapping.ValidatorMetadata[validatorIndex]

		switch constypes.ValidatorDbStatus(metadata.Status) {
		case constypes.DbDeposited:
			result.Deposited = append(result.Deposited, validatorIndex)
		case constypes.DbPending:
			validatorInfo := t.IndexTimestamp{
				Index: validatorIndex,
			}
			if metadata.ActivationEpoch.Valid {
				validatorInfo.Timestamp = uint64(utils.EpochToTime(uint64(metadata.ActivationEpoch.Int64)).Unix())
			} else if metadata.Queues.ActivationIndex.Valid {
				queuePosition := uint64(metadata.Queues.ActivationIndex.Int64)
				epochsToWait := (queuePosition - 1) / activationChurnRate
				// calculate dequeue epoch
				estimatedActivationEpoch := latestEpoch + epochsToWait + 1
				// add activation offset
				estimatedActivationEpoch += utils.Config.Chain.ClConfig.MaxSeedLookahead + 1
				validatorInfo.Timestamp = uint64(utils.EpochToTime(estimatedActivationEpoch).Unix())
			}
			result.Pending = append(result.Pending, validatorInfo)
		case constypes.DbActiveOnline:
			result.Online = append(result.Online, validatorIndex)
		case constypes.DbActiveOffline:
			result.Offline = append(result.Offline, validatorIndex)
		case constypes.DbSlashingOnline, constypes.DbSlashingOffline:
			result.Slashing = append(result.Slashing, validatorIndex)
			if constypes.ValidatorDbStatus(metadata.Status) == constypes.DbSlashingOffline {
				result.Offline = append(result.Offline, validatorIndex)
			} else {
				result.Online = append(result.Online, validatorIndex)
			}
		case constypes.DbExitingOnline, constypes.DbExitingOffline:
			result.Exiting = append(result.Exiting, t.IndexTimestamp{
				Index:     validatorIndex,
				Timestamp: uint64(utils.EpochToTime(uint64(metadata.ExitEpoch.Int64)).Unix()),
			})
			if constypes.ValidatorDbStatus(metadata.Status) == constypes.DbExitingOffline {
				result.Offline = append(result.Offline, validatorIndex)
			} else {
				result.Online = append(result.Online, validatorIndex)
			}
		case constypes.DbExited, constypes.DbSlashed:
			if constypes.ValidatorDbStatus(metadata.Status) == constypes.DbSlashed {
				result.Slashed = append(result.Slashed, validatorIndex)
			} else {
				result.Exited = append(result.Exited, validatorIndex)
			}

			if metadata.WithdrawableEpoch.Valid && metadata.WithdrawableEpoch.Int64 <= int64(latestEpoch) {
				if metadata.Balance != 0 {
					validatorInfo := t.IndexTimestamp{
						Index: validatorIndex,
					}

					if utils.IsValidWithdrawalCredentialsAddress(fmt.Sprintf("%x", metadata.WithdrawalCredentials)) {
						distance, err := d.getWithdrawableCountFromCursor(validatorIndex, *stats.LatestValidatorWithdrawalIndex)
						if err != nil {
							return nil, err
						}

						timeToWithdrawal := d.getTimeToNextWithdrawal(distance)
						validatorInfo.Timestamp = uint64(timeToWithdrawal.Unix())
					}

					result.Withdrawing = append(result.Withdrawing, validatorInfo)
				} else {
					result.Withdrawn = append(result.Withdrawn, validatorIndex)
				}
			}
		}
	}

	return result, nil
}

func (d *DataAccessService) GetValidatorDashboardSyncSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSyncSummaryValidators, error) {
	// possible periods are: all_time, last_30d, last_7d, last_24h, last_1h
	result := &t.VDBSyncSummaryValidators{}
	var resultMutex = &sync.RWMutex{}
	wg := errgroup.Group{}

	// Get the table name based on the period
	clickhouseTable, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, err
	}

	// Get the validator indices
	var groupIds []uint64
	if !dashboardId.AggregateGroups && groupId != t.AllGroups {
		groupIds = append(groupIds, uint64(groupId))
	}

	validatorIndices, err := d.getDashboardValidators(ctx, dashboardId, groupIds)
	if err != nil {
		return nil, err
	}

	// Get the current and next sync committee validators
	latestEpoch := cache.LatestEpoch.Get()
	wg.Go(func() error {
		currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, err := d.getCurrentAndUpcomingSyncCommittees(ctx, latestEpoch)
		if err != nil {
			return err
		}

		resultMutex.Lock()
		for _, validatorIndex := range validatorIndices {
			if currentSyncCommitteeValidators[validatorIndex] {
				result.Current = append(result.Current, validatorIndex)
			}
			if upcomingSyncCommitteeValidators[validatorIndex] {
				result.Upcoming = append(result.Upcoming, validatorIndex)
			}
		}
		resultMutex.Unlock()

		return nil
	})

	// Get the past sync committee validators
	wg.Go(func() error {
		// Get the cutoff period for past sync committees
		ds := goqu.Dialect("postgres").
			Select(
				goqu.L("epoch_start")).
			From(goqu.L(fmt.Sprintf("%s FINAL", clickhouseTable))).
			Limit(1)

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		var epochStart uint64
		err = d.clickhouseReader.GetContext(ctx, &epochStart, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving cutoff epoch for past sync committees: %w", err)
		}
		pastSyncPeriodCutoff := utils.SyncPeriodOfEpoch(epochStart)

		// Get the past sync committee validators
		currentSyncPeriod := utils.SyncPeriodOfEpoch(latestEpoch)
		ds = goqu.Dialect("postgres").
			Select(
				goqu.L("sc.validatorindex")).
			From(goqu.L("sync_committees sc")).
			Where(goqu.L("period >= ? AND period < ? AND validatorindex = ANY(?)", pastSyncPeriodCutoff, currentSyncPeriod, pq.Array(validatorIndices)))

		query, args, err = ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		var validatorIndices []uint64
		err = d.alloyReader.SelectContext(ctx, &validatorIndices, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data for past sync committees: %w", err)
		}

		validatorCountMap := make(map[uint64]uint64)
		for _, validatorIndex := range validatorIndices {
			validatorCountMap[validatorIndex]++
		}

		resultMutex.Lock()
		for validatorIndex, count := range validatorCountMap {
			result.Past = append(result.Past, t.VDBValidatorSyncPast{
				Index: validatorIndex,
				Count: count,
			})
		}
		resultMutex.Unlock()

		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DataAccessService) GetValidatorDashboardSlashingsSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSlashingsSummaryValidators, error) {
	// possible periods are: all_time, last_30d, last_7d, last_24h, last_1h
	result := &t.VDBSlashingsSummaryValidators{}

	// Get the table names based on the period
	clickhouseTable, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, err
	}

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	var queryResult []struct {
		EpochStart     uint64 `db:"epoch_start"`
		EpochEnd       uint64 `db:"epoch_end"`
		ValidatorIndex uint64 `db:"validator_index"`
		Slashed        bool   `db:"slashed"`
		SlashedAmount  uint32 `db:"slashed_amount"`
	}

	// Build the query
	ds := goqu.Dialect("postgres").
		From(goqu.L(fmt.Sprintf("%s AS r FINAL", clickhouseTable))).
		With("validators", goqu.L("(SELECT group_id, validator_index FROM users_val_dashboards_validators WHERE dashboard_id = ?)", dashboardId.Id)).
		Select(
			goqu.L("r.epoch_start"),
			goqu.L("r.epoch_end"),
			goqu.L("r.validator_index"),
			goqu.L("r.slashed"),
			goqu.L("COALESCE(r.blocks_slashing_count, 0) AS slashed_amount")).
		Where(goqu.L("(r.slashed OR r.blocks_slashing_count > 0)"))

	// handle the case when we have a list of validators
	if len(dashboardId.Validators) > 0 {
		ds = ds.
			Where(goqu.L("r.validator_index IN ?", dashboardId.Validators))
	} else {
		ds = ds.
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
			Where(goqu.L("r.validator_index IN (SELECT validator_index FROM validators)"))

		if groupId != t.AllGroups {
			ds = ds.Where(goqu.L("v.group_id = ?", groupId))
		}
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = d.clickhouseReader.SelectContext(ctx, &queryResult, query, args...)
	if err != nil {
		log.Error(err, "error while getting validator dashboard slashed validators list", 0)
		return nil, err
	}

	// Process the data and get the slashing validators
	var slashingValidators []uint64
	var slashedValidators []uint64
	for _, queryEntry := range queryResult {
		if queryEntry.SlashedAmount > 0 {
			slashingValidators = append(slashingValidators, queryEntry.ValidatorIndex)
		}

		if queryEntry.Slashed {
			slashedValidators = append(slashedValidators, queryEntry.ValidatorIndex)
		}
	}

	if len(slashingValidators) == 0 && len(slashedValidators) == 0 {
		// We don't have any slashing or slashed validators so we can return early
		return result, nil
	}

	slashingValidatorsMap := utils.SliceToMap(slashingValidators)
	slashedValidatorsMap := utils.SliceToMap(slashedValidators)

	// If we have slashing validators then get the validators that got slashed
	proposalSlashings := make(map[uint64][]uint64)
	proposalSlashed := make(map[uint64]uint64)
	attestationSlashings := make(map[uint64][]uint64)
	attestationSlashed := make(map[uint64]uint64)

	slotStart := queryResult[0].EpochStart * utils.Config.Chain.ClConfig.SlotsPerEpoch
	slotEnd := (queryResult[0].EpochEnd+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1

	wg := errgroup.Group{}

	// Get the proposal slashings
	wg.Go(func() error {
		var queryResult []struct {
			ProposerSlashing uint64 `db:"proposer"`
			ProposerSlashed  uint64 `db:"proposerindex"`
		}

		ds := goqu.Dialect("postgres").
			Select(
				goqu.L("b.proposer"),
				goqu.L("bps.proposerindex")).
			From(goqu.L("blocks_proposerslashings bps")).
			LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("b.slot = bps.block_slot"))).
			Where(goqu.L("bps.block_slot >= ? AND bps.block_slot <= ?", slotStart, slotEnd)).
			Where(goqu.L("(b.proposer = ANY(?) OR bps.proposerindex = ANY(?))", pq.Array(slashingValidators), pq.Array(slashedValidators)))

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.SelectContext(ctx, &queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data from table blocks_proposerslashings: %v", err)
		}

		for _, queryEntry := range queryResult {
			if _, ok := slashingValidatorsMap[queryEntry.ProposerSlashing]; ok {
				if _, ok := proposalSlashings[queryEntry.ProposerSlashing]; !ok {
					proposalSlashings[queryEntry.ProposerSlashing] = make([]uint64, 0)
				}
				proposalSlashings[queryEntry.ProposerSlashing] = append(proposalSlashings[queryEntry.ProposerSlashing], queryEntry.ProposerSlashed)
			}
			if _, ok := slashedValidatorsMap[queryEntry.ProposerSlashed]; ok {
				proposalSlashed[queryEntry.ProposerSlashed] = queryEntry.ProposerSlashing
			}
		}
		return nil
	})

	// Get the attestation slashings
	wg.Go(func() error {
		var queryResult []struct {
			Proposer               uint64        `db:"proposer"`
			Attestestation1Indices pq.Int64Array `db:"attestation1_indices"`
			Attestestation2Indices pq.Int64Array `db:"attestation2_indices"`
		}

		ds := goqu.Dialect("postgres").
			Select(
				goqu.L("b.proposer"),
				goqu.L("bas.attestation1_indices"),
				goqu.L("bas.attestation2_indices")).
			From(goqu.L("blocks_attesterslashings bas")).
			LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("b.slot = bas.block_slot"))).
			Where(goqu.L("bas.block_slot >= ? AND bas.block_slot <= ?", slotStart, slotEnd))

		if len(slashedValidators) == 0 {
			// If we don't have any slashed validators then we can just get the slashing validators
			ds = ds.
				Where(goqu.L("b.proposer = ANY(?)", pq.Array(slashingValidators)))
		}

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.SelectContext(ctx, &queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data from table blocks_attesterslashings: %v", err)
		}

		for _, queryEntry := range queryResult {
			inter := intersect.Simple(queryEntry.Attestestation1Indices, queryEntry.Attestestation2Indices)
			if len(inter) == 0 {
				log.WarnWithStackTrace(nil, "No intersection found for attestation violation", 0)
			}
			for _, v := range inter {
				if _, ok := slashingValidatorsMap[queryEntry.Proposer]; ok {
					if _, ok := attestationSlashings[queryEntry.Proposer]; !ok {
						attestationSlashings[queryEntry.Proposer] = make([]uint64, 0)
					}
					attestationSlashings[queryEntry.Proposer] = append(attestationSlashings[queryEntry.Proposer], uint64(v.(int64)))
				}
				if _, ok := slashedValidatorsMap[uint64(v.(int64))]; ok {
					attestationSlashed[uint64(v.(int64))] = queryEntry.Proposer
				}
			}
		}
		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, err
	}

	// Combine the proposal and attestation slashings
	slashings := make(map[uint64][]uint64)
	for slashingIdx, slashedIdxs := range proposalSlashings {
		if _, ok := slashings[slashingIdx]; !ok {
			slashings[slashingIdx] = make([]uint64, 0)
		}
		slashings[slashingIdx] = append(slashings[slashingIdx], slashedIdxs...)
	}
	for slashingIdx, slashedIdxs := range attestationSlashings {
		if _, ok := slashings[slashingIdx]; !ok {
			slashings[slashingIdx] = make([]uint64, 0)
		}
		slashings[slashingIdx] = append(slashings[slashingIdx], slashedIdxs...)
	}

	// Process the data
	for slashingIdx, slashedIdxs := range slashings {
		result.HasSlashed = append(result.HasSlashed, t.VDBValidatorHasSlashed{
			Index:          slashingIdx,
			SlashedIndices: slashedIdxs,
		})
	}

	// Fill the slashed validators
	for slashedIdx, slashingIdx := range proposalSlashed {
		result.GotSlashed = append(result.GotSlashed, t.VDBValidatorGotSlashed{
			Index:     slashedIdx,
			SlashedBy: slashingIdx,
		})
	}
	for slashedIdx, slashingIdx := range attestationSlashed {
		result.GotSlashed = append(result.GotSlashed, t.VDBValidatorGotSlashed{
			Index:     slashedIdx,
			SlashedBy: slashingIdx,
		})
	}

	return result, nil
}

func (d *DataAccessService) GetValidatorDashboardProposalSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBProposalSummaryValidators, error) {
	// possible periods are: all_time, last_30d, last_7d, last_24h, last_1h
	result := &t.VDBProposalSummaryValidators{}

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	// Get the table name based on the period
	clickhouseTable, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, err
	}

	var epochQueryResult struct {
		EpochStart uint64 `db:"epoch_start"`
		EpochEnd   uint64 `db:"epoch_end"`
	}

	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("epoch_start"),
			goqu.L("epoch_end")).
		From(goqu.L(fmt.Sprintf("%s FINAL", clickhouseTable))).
		Limit(1)

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	err = d.clickhouseReader.GetContext(ctx, &epochQueryResult, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving epoch info for proposals: %w", err)
	}

	// Build the query and get the data
	var queryResult []struct {
		Slot           uint64        `db:"slot"`
		Block          sql.NullInt64 `db:"exec_block_number"`
		Status         string        `db:"status"`
		ValidatorIndex uint64        `db:"proposer"`
	}

	ds = goqu.Dialect("postgres").
		Select(
			goqu.L("b.slot"),
			goqu.L("b.exec_block_number"),
			goqu.L("b.status"),
			goqu.L("b.proposer")).
		From(goqu.L("blocks b")).
		Where(goqu.L("b.epoch >= ? AND b.epoch <= ?", epochQueryResult.EpochStart, epochQueryResult.EpochEnd))

	if len(dashboardId.Validators) > 0 {
		ds = ds.
			Where(goqu.L("b.proposer = ANY(?)", pq.Array(dashboardId.Validators)))
	} else {
		ds = ds.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("b.proposer = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

		if groupId != t.AllGroups {
			ds = ds.Where(goqu.L("v.group_id = ?", groupId))
		}
	}

	query, args, err = ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %v", err)
	}

	err = d.alloyReader.SelectContext(ctx, &queryResult, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving data from table blocks: %v", err)
	}

	// Process the data
	proposedValidatorMap := make(map[uint64][]uint64)
	missedValidatorMap := make(map[uint64][]uint64)
	for _, row := range queryResult {
		if row.Status == "1" {
			if _, ok := proposedValidatorMap[row.ValidatorIndex]; !ok {
				proposedValidatorMap[row.ValidatorIndex] = make([]uint64, 0)
			}
			if !row.Block.Valid {
				return nil, fmt.Errorf("error no block number for slot %v found", row.Slot)
			}
			proposedValidatorMap[row.ValidatorIndex] = append(proposedValidatorMap[row.ValidatorIndex], uint64(row.Block.Int64))
		} else {
			if _, ok := missedValidatorMap[row.ValidatorIndex]; !ok {
				missedValidatorMap[row.ValidatorIndex] = make([]uint64, 0)
			}
			missedValidatorMap[row.ValidatorIndex] = append(missedValidatorMap[row.ValidatorIndex], row.Slot)
		}
	}

	for validatorIndex, blockNumbers := range proposedValidatorMap {
		result.Proposed = append(result.Proposed, t.IndexBlocks{
			Index:  validatorIndex,
			Blocks: blockNumbers,
		})
	}
	for validatorIndex, slotNumbers := range missedValidatorMap {
		result.Missed = append(result.Missed, t.IndexBlocks{
			Index:  validatorIndex,
			Blocks: slotNumbers,
		})
	}

	return result, nil
}

func (d *DataAccessService) getCurrentAndUpcomingSyncCommittees(ctx context.Context, latestEpoch uint64) (map[uint64]bool, map[uint64]bool, error) {
	currentSyncCommitteeValidators := make(map[uint64]bool)
	upcomingSyncCommitteeValidators := make(map[uint64]bool)

	currentSyncPeriod := utils.SyncPeriodOfEpoch(latestEpoch)
	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("validatorindex"),
			goqu.L("period")).
		From("sync_committees").
		Where(goqu.L("period IN (?, ?)", currentSyncPeriod, currentSyncPeriod+1))

	var queryResult []struct {
		ValidatorIndex uint64 `db:"validatorindex"`
		Period         uint64 `db:"period"`
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, nil, fmt.Errorf("error preparing query: %w", err)
	}

	err = d.readerDb.SelectContext(ctx, &queryResult, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving sync committee current and next period data: %w", err)
	}

	for _, queryEntry := range queryResult {
		if queryEntry.Period == currentSyncPeriod {
			currentSyncCommitteeValidators[queryEntry.ValidatorIndex] = true
		} else {
			upcomingSyncCommitteeValidators[queryEntry.ValidatorIndex] = true
		}
	}

	return currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, nil
}

func (d *DataAccessService) getTablesForPeriod(period enums.TimePeriod) (string, int, error) {
	clickhouseTable := ""
	hours := 0

	switch period {
	case enums.TimePeriods.Last1h:
		clickhouseTable = "validator_dashboard_data_rolling_1h"
		hours = 1
	case enums.TimePeriods.Last24h:
		clickhouseTable = "validator_dashboard_data_rolling_24h"
		hours = 24
	case enums.TimePeriods.Last7d:
		clickhouseTable = "validator_dashboard_data_rolling_7d"
		hours = 7 * 24
	case enums.TimePeriods.Last30d:
		clickhouseTable = "validator_dashboard_data_rolling_30d"
		hours = 30 * 24
	case enums.TimePeriods.AllTime:
		clickhouseTable = "validator_dashboard_data_rolling_total"
		hours = -1
	default:
		return "", 0, fmt.Errorf("not-implemented time period: %v", period)
	}

	return clickhouseTable, hours, nil
}
