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
	"github.com/gobitfly/beaconchain/pkg/commons/db"
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
	table, _, _, err := d.getTablesForPeriod(period)
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
		ValidatorIndices       pq.Int64Array   `db:"validator_indices"`
		ClRewards              int64           `db:"cl_rewards"`
		AttestationReward      decimal.Decimal `db:"attestations_reward"`
		AttestationIdealReward decimal.Decimal `db:"attestations_ideal_reward"`
		AttestationsExecuted   uint64          `db:"attestations_executed"`
		AttestationsScheduled  uint64          `db:"attestations_scheduled"`
		BlocksProposed         uint64          `db:"blocks_proposed"`
		BlocksScheduled        uint64          `db:"blocks_scheduled"`
		SyncExecuted           uint64          `db:"sync_executed"`
		SyncScheduled          uint64          `db:"sync_scheduled"`
	}

	wg.Go(func() error {
		ds := goqu.Dialect("postgres").
			Select(
				goqu.L("ARRAY_AGG(r.validator_index) AS validator_indices"),
				goqu.L("SUM(COALESCE(r.attestations_reward, 0) + COALESCE(r.blocks_cl_reward, 0) + COALESCE(r.sync_rewards, 0) + COALESCE(r.slasher_reward, 0)) AS cl_rewards"),
				goqu.L("COALESCE(SUM(r.attestations_reward)::decimal, 0) AS attestations_reward"),
				goqu.L("COALESCE(SUM(r.attestations_ideal_reward)::decimal, 0) AS attestations_ideal_reward"),
				goqu.L("COALESCE(SUM(r.attestations_executed), 0) AS attestations_executed"),
				goqu.L("COALESCE(SUM(r.attestations_scheduled), 0) AS attestations_scheduled"),
				goqu.L("COALESCE(SUM(r.blocks_proposed), 0) AS blocks_proposed"),
				goqu.L("COALESCE(SUM(r.blocks_scheduled), 0) AS blocks_scheduled"),
				goqu.L("COALESCE(SUM(r.sync_executed), 0) AS sync_executed"),
				goqu.L("COALESCE(SUM(r.sync_scheduled), 0) AS sync_scheduled")).
			From(goqu.T(table).As("r")).
			GroupBy(goqu.L("result_group_id"))

		if len(validators) > 0 {
			ds = ds.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
				Where(goqu.L("r.validator_index = ANY(?)", pq.Array(validators)))
		} else {
			if dashboardId.AggregateGroups {
				ds = ds.
					SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId))
			} else {
				ds = ds.
					SelectAppend(goqu.L("v.group_id AS result_group_id"))
			}

			ds = ds.
				InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
				Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

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
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Select(&queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data from table %s: %v", table, err)
		}
		return nil
	})

	// ------------------------------------------------------------------------------------------------------------------
	// Get the EL rewards
	elRewards := make(map[int64]decimal.Decimal)
	wg.Go(func() error {
		ds := goqu.Dialect("postgres").
			Select(
				goqu.L("SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)) AS el_rewards")).
			From(goqu.T(table).As("r")).
			LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("b.epoch >= r.epoch_start AND b.epoch <= r.epoch_end AND r.validator_index = b.proposer AND b.status = '1'"))).
			LeftJoin(goqu.L("execution_payloads ep"), goqu.On(goqu.L("ep.block_hash = b.exec_block_hash"))).
			LeftJoin(goqu.L("relays_blocks rb"), goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash"))).
			GroupBy(goqu.L("result_group_id"))

		if len(validators) > 0 {
			ds = ds.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
				Where(goqu.L("r.validator_index = ANY(?)", pq.Array(validators)))
		} else {
			if dashboardId.AggregateGroups {
				ds = ds.
					SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId))
			} else {
				ds = ds.
					SelectAppend(goqu.L("v.group_id AS result_group_id"))
			}

			ds = ds.
				InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
				Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))
		}

		var queryResult []struct {
			GroupId   int64           `db:"result_group_id"`
			ElRewards decimal.Decimal `db:"el_rewards"`
		}

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Select(&queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data from table %s: %v", table, err)
		}

		for _, entry := range queryResult {
			elRewards[entry.GroupId] = entry.ElRewards
		}
		return nil
	})

	// ------------------------------------------------------------------------------------------------------------------
	// Get the current and next sync committee validators
	latestEpoch := cache.LatestEpoch.Get()
	currentSyncCommitteeValidators := make(map[uint64]bool)
	upcomingSyncCommitteeValidators := make(map[uint64]bool)
	wg.Go(func() error {
		var err error
		currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, err = d.getCurrentAndUpcomingSyncCommittees(latestEpoch)
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
		uiValidatorIndices := make([]uint64, len(queryEntry.ValidatorIndices))
		for i, validatorIndex := range queryEntry.ValidatorIndices {
			uiValidatorIndices[i] = uint64(validatorIndex)
		}

		resultEntry := t.VDBSummaryTableRow{
			GroupId:                  queryEntry.GroupId,
			AverageNetworkEfficiency: averageNetworkEfficiency,
		}

		// Status
		for _, validatorIndex := range uiValidatorIndices {
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
		validatorStatuses, err := d.getValidatorStatuses(uiValidatorIndices)
		if err != nil {
			return nil, nil, err
		}
		for _, status := range validatorStatuses {
			if status == enums.ValidatorStatuses.Online {
				resultEntry.Validators.Online++
			} else if status == enums.ValidatorStatuses.Offline {
				resultEntry.Validators.Offline++
			} else {
				if status == enums.ValidatorStatuses.Slashed {
					resultEntry.Status.SlashedCount++
				}
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
			for _, validatorIndex := range uiValidatorIndices {
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

	var err error
	ret := &t.VDBGroupSummaryData{}

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	// Get the current and next sync committee validators
	latestEpoch := cache.LatestEpoch.Get()
	currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, err := d.getCurrentAndUpcomingSyncCommittees(latestEpoch)
	if err != nil {
		return nil, err
	}

	// Get the table names based on the period
	table, slashedByCountTable, days, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, err
	}

	query := `select
			users_val_dashboards_validators.validator_index,
			epoch_start,
			COALESCE(attestations_reward, 0) as attestations_reward,
			COALESCE(attestations_ideal_reward, 0) as attestations_ideal_reward,
			COALESCE(attestations_scheduled, 0) as attestations_scheduled,
			COALESCE(attestation_head_executed, 0) as attestation_head_executed,
			COALESCE(attestation_source_executed, 0) as attestation_source_executed,
			COALESCE(attestation_target_executed, 0) as attestation_target_executed,
			COALESCE(blocks_scheduled, 0) as blocks_scheduled,
			COALESCE(blocks_proposed, 0) as blocks_proposed,
			COALESCE(sync_scheduled, 0) as sync_scheduled,
			COALESCE(sync_executed, 0) as sync_executed,
			%[1]s.slashed_by IS NOT NULL AS slashed_in_period,
			COALESCE(%[2]s.slashed_amount, 0) AS slashed_amount,
			COALESCE(blocks_expected, 0) as blocks_expected,
			COALESCE(inclusion_delay_sum, 0) as inclusion_delay_sum,
			COALESCE(sync_committees_expected, 0) as sync_committees_expected
		from users_val_dashboards_validators
		inner join %[1]s on %[1]s.validator_index = users_val_dashboards_validators.validator_index
		left join %[2]s on %[2]s.slashed_by = users_val_dashboards_validators.validator_index
		where (dashboard_id = $1 and (group_id = $2 OR $2 = -1))
		`

	if dashboardId.Validators != nil {
		query = `select
			validator_index,
			epoch_start,
			COALESCE(attestations_reward, 0) as attestations_reward,
			COALESCE(attestations_ideal_reward, 0) as attestations_ideal_reward,
			COALESCE(attestations_scheduled, 0) as attestations_scheduled,
			COALESCE(attestation_head_executed, 0) as attestation_head_executed,
			COALESCE(attestation_source_executed, 0) as attestation_source_executed,
			COALESCE(attestation_target_executed, 0) as attestation_target_executed,
			COALESCE(blocks_scheduled, 0) as blocks_scheduled,
			COALESCE(blocks_proposed, 0) as blocks_proposed,
			COALESCE(sync_scheduled, 0) as sync_scheduled,
			COALESCE(sync_executed, 0) as sync_executed,
			%[1]s.slashed_by IS NOT NULL AS slashed_in_period,
			COALESCE(%[2]s.slashed_amount, 0) AS slashed_amount,
			COALESCE(blocks_expected, 0) as blocks_expected,
			COALESCE(inclusion_delay_sum, 0) as inclusion_delay_sum,
			COALESCE(sync_committees_expected, 0) as sync_committees_expected
		from %[1]s
		left join %[2]s on %[2]s.slashed_by = %[1]s.validator_index
		where %[1]s.validator_index = ANY($1)
	`
	}

	validators := make([]t.VDBValidator, 0)
	if dashboardId.Validators != nil {
		validators = dashboardId.Validators
	}

	type queryResult struct {
		ValidatorIndex         uint32 `db:"validator_index"`
		EpochStart             uint64 `db:"epoch_start"`
		AttestationReward      int64  `db:"attestations_reward"`
		AttestationIdealReward int64  `db:"attestations_ideal_reward"`

		AttestationsScheduled     int64 `db:"attestations_scheduled"`
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

	var rows []*queryResult

	if len(validators) > 0 {
		err = d.alloyReader.Select(&rows, fmt.Sprintf(query, table, slashedByCountTable), validators)
	} else {
		err = d.alloyReader.Select(&rows, fmt.Sprintf(query, table, slashedByCountTable), dashboardId.Id, groupId)
	}

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard group summary data: %v", err)
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
			totalInclusionDelayDivisor += row.AttestationsScheduled
		}
	}

	_, ret.Apr.El, _, ret.Apr.Cl, err = d.internal_getElClAPR(ctx, validatorArr, days)
	if err != nil {
		return nil, err
	}

	if len(validators) > 0 {
		validatorArr = validators
	}

	pastSyncPeriodCutoff := utils.SyncPeriodOfEpoch(rows[0].EpochStart)
	currentSyncPeriod := utils.SyncPeriodOfEpoch(latestEpoch)
	err = d.readerDb.Get(&ret.SyncCommitteeCount.PastPeriods, `SELECT COUNT(*) FROM sync_committees WHERE period >= $1 AND period < $2 AND validatorindex = ANY($3)`, pastSyncPeriodCutoff, currentSyncPeriod, validatorArr)
	if err != nil {
		return nil, fmt.Errorf("error retrieving past sync committee count: %v", err)
	}

	ret.AttestationEfficiency = float64(totalAttestationRewards) / float64(totalIdealAttestationRewards) * 100
	if ret.AttestationEfficiency < 0 || math.IsNaN(ret.AttestationEfficiency) {
		ret.AttestationEfficiency = 0
	}

	luckDays := float64(days)
	if days == -1 {
		luckDays = time.Since(time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)).Hours() / 24
		if luckDays == 0 {
			luckDays = 1
		}
	}

	if totalBlockChance > 0 {
		ret.Luck.Proposal.Percent = (float64(totalProposals)) / totalBlockChance * 100

		// calculate the average time it takes for the set of validators to propose a single block on average
		ret.Luck.Proposal.Average = time.Duration((luckDays / totalBlockChance) * float64(utils.Day))
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
		ret.Luck.Sync.Average = time.Duration((luckDays / totalSyncExpected) * float64(utils.Day))
	}

	if totalInclusionDelayDivisor > 0 {
		ret.AttestationAvgInclDist = 1.0 + float64(totalInclusionDelaySum)/float64(totalInclusionDelayDivisor)
	} else {
		ret.AttestationAvgInclDist = 0
	}

	return ret, nil
}

func (d *DataAccessService) internal_getElClAPR(ctx context.Context, validators []t.VDBValidator, days int) (elIncome decimal.Decimal, elAPR float64, clIncome decimal.Decimal, clAPR float64, err error) {
	var reward sql.NullInt64
	table := ""

	switch days {
	case 1:
		table = "validator_dashboard_data_rolling_daily"
	case 7:
		table = "validator_dashboard_data_rolling_weekly"
	case 30:
		table = "validator_dashboard_data_rolling_monthly"
	case -1:
		table = "validator_dashboard_data_rolling_90d"
	default:
		return decimal.Zero, 0, decimal.Zero, 0, fmt.Errorf("invalid days value: %v", days)
	}

	query := `select (SUM(COALESCE(balance_end,0)) + SUM(COALESCE(withdrawals_amount,0)) - SUM(COALESCE(deposits_amount,0)) - SUM(COALESCE(balance_start,0))) reward FROM %s WHERE validator_index = ANY($1)`

	err = db.AlloyReader.Get(&reward, fmt.Sprintf(query, table), validators)
	if err != nil || !reward.Valid {
		return decimal.Zero, 0, decimal.Zero, 0, err
	}

	aprDivisor := days
	if days == -1 { // for all time APR
		aprDivisor = 90
	}
	clAPR = ((float64(reward.Int64) / float64(aprDivisor)) / (float64(32e9) * float64(len(validators)))) * 365.0 * 100.0
	if math.IsNaN(clAPR) {
		clAPR = 0
	}
	if days == -1 {
		err = db.AlloyReader.Get(&reward, fmt.Sprintf(query, "validator_dashboard_data_rolling_total"), validators)
		if err != nil || !reward.Valid {
			return decimal.Zero, 0, decimal.Zero, 0, err
		}
	}
	clIncome = decimal.NewFromInt(reward.Int64).Mul(decimal.NewFromInt(1e9))

	query = `
	SELECT 
		COALESCE(SUM(COALESCE(rb.value / 1e18, fee_recipient_reward)), 0)
	FROM blocks 
	LEFT JOIN execution_payloads ON blocks.exec_block_hash = execution_payloads.block_hash
	LEFT JOIN relays_blocks rb ON blocks.exec_block_hash = rb.exec_block_hash
	WHERE proposer = ANY($1) AND status = '1' AND slot >= (SELECT MIN(epoch_start) * $2 FROM %s WHERE validator_index = ANY($1));`
	err = db.AlloyReader.Get(&elIncome, fmt.Sprintf(query, table), validators, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if err != nil {
		return decimal.Zero, 0, decimal.Zero, 0, err
	}
	elIncomeFloat, _ := elIncome.Float64()
	elAPR = ((elIncomeFloat / float64(aprDivisor)) / (float64(32e18) * float64(len(validators)))) * 365.0 * 100.0

	if days == -1 {
		err = db.AlloyReader.Get(&elIncome, fmt.Sprintf(query, "validator_dashboard_data_rolling_total"), validators, utils.Config.Chain.ClConfig.SlotsPerEpoch)
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

	type queryResult struct {
		StartEpoch            uint64          `db:"epoch_start"`
		GroupId               uint64          `db:"group_id"`
		AttestationEfficiency sql.NullFloat64 `db:"attestation_efficiency"`
		ProposerEfficiency    sql.NullFloat64 `db:"proposer_efficiency"`
		SyncEfficiency        sql.NullFloat64 `db:"sync_efficiency"`
	}

	var queryResults []queryResult

	cutOffDate := time.Date(2023, 9, 27, 23, 59, 59, 0, time.UTC).Add(time.Hour*24*30).AddDate(0, 0, -30)

	if dashboardId.Validators != nil {
		query := `select epoch_start, 0 AS group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
			select
				epoch_start,
				COALESCE(SUM(attestations_reward), 0)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
				COALESCE(SUM(blocks_proposed), 0)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
				COALESCE(SUM(sync_executed), 0)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
			from  validator_dashboard_data_daily
			WHERE day > $1 AND validator_index = ANY($2)
			group by 1
		) as a ORDER BY epoch_start;`
		err := d.alloyReader.Select(&queryResults, query, cutOffDate, dashboardId.Validators)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table validator_dashboard_data_daily: %v", err)
		}
	} else {
		queryParams := []interface{}{cutOffDate, dashboardId.Id}

		groupIdQuery := "v.group_id,"
		groupByQuery := "GROUP BY 1, 2"
		orderQuery := "ORDER BY epoch_start, group_id"
		if dashboardId.AggregateGroups {
			queryParams = append(queryParams, t.DefaultGroupId)
			groupIdQuery = "$3::smallint AS group_id,"
			groupByQuery = "GROUP BY 1"
			orderQuery = "ORDER BY epoch_start"
		}
		query := fmt.Sprintf(`SELECT epoch_start, group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
			SELECT
				d.epoch_start, 
				%s
				COALESCE(SUM(d.attestations_reward), 0)::decimal / NULLIF(SUM(d.attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
				COALESCE(SUM(d.blocks_proposed), 0)::decimal / NULLIF(SUM(d.blocks_scheduled)::decimal, 0) AS proposer_efficiency,
				COALESCE(SUM(d.sync_executed), 0)::decimal / NULLIF(SUM(d.sync_scheduled)::decimal, 0) AS sync_efficiency
			FROM users_val_dashboards_validators v
			INNER JOIN validator_dashboard_data_daily d ON d.validator_index = v.validator_index
			WHERE day > $1 AND dashboard_id = $2
			%s
		) as a %s;`, groupIdQuery, groupByQuery, orderQuery)
		err := d.alloyReader.Select(&queryResults, query, queryParams...)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table validator_dashboard_data_daily: %v", err)
		}
	}

	// convert the returned data to the expected return type (not pretty)
	epochsMap := make(map[uint64]bool)
	groups := make(map[uint64]bool)
	data := make(map[uint64]map[uint64]float64)
	for _, row := range queryResults {
		epochsMap[row.StartEpoch] = true
		groups[row.GroupId] = true

		if data[row.StartEpoch] == nil {
			data[row.StartEpoch] = make(map[uint64]float64)
		}
		data[row.StartEpoch][row.GroupId] = d.calculateTotalEfficiency(row.AttestationEfficiency, row.ProposerEfficiency, row.SyncEfficiency)
	}

	epochsArray := make([]uint64, 0, len(epochsMap))
	for epoch := range epochsMap {
		epochsArray = append(epochsArray, epoch)
	}
	sort.Slice(epochsArray, func(i, j int) bool {
		return epochsArray[i] < epochsArray[j]
	})

	groupsArray := make([]uint64, 0, len(groups))
	for group := range groups {
		groupsArray = append(groupsArray, group)
	}
	sort.Slice(groupsArray, func(i, j int) bool {
		return groupsArray[i] < groupsArray[j]
	})

	ret.Categories = epochsArray
	ret.Series = make([]t.ChartSeries[int, float64], 0, len(groupsArray))

	seriesMap := make(map[uint64]*t.ChartSeries[int, float64])
	for group := range groups {
		series := t.ChartSeries[int, float64]{
			Id:   int(group),
			Data: make([]float64, 0, len(epochsMap)),
		}
		seriesMap[group] = &series
	}

	for _, epoch := range epochsArray {
		for _, group := range groupsArray {
			seriesMap[group].Data = append(seriesMap[group].Data, data[epoch][group])
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

	// Get the validator duties to check the last fulfilled attestation
	dutiesInfo, releaseValDutiesLock, err := d.services.GetCurrentDutiesInfo()
	defer releaseValDutiesLock()
	if err != nil {
		return nil, err
	}

	// Set the threshold for "online" => "offline" to 2 epochs without attestation
	attestationThresholdSlot := uint64(0)
	twoEpochs := 2 * utils.Config.Chain.ClConfig.SlotsPerEpoch
	if dutiesInfo.LatestSlot >= twoEpochs {
		attestationThresholdSlot = dutiesInfo.LatestSlot - twoEpochs
	}

	// Fill the data
	for _, validatorIndex := range validatorIndices {
		metadata := validatorMapping.ValidatorMetadata[validatorIndex]

		switch constypes.ValidatorStatus(metadata.Status) {
		case constypes.PendingInitialized:
			result.Deposited = append(result.Deposited, validatorIndex)
		case constypes.PendingQueued:
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
		case constypes.ActiveOngoing, constypes.ActiveExiting, constypes.ActiveSlashed:
			var lastAttestionSlot uint32
			for slot, attested := range dutiesInfo.EpochAttestationDuties[validatorIndex] {
				if attested && slot > lastAttestionSlot {
					lastAttestionSlot = slot
				}
			}
			if lastAttestionSlot < uint32(attestationThresholdSlot) {
				result.Offline = append(result.Offline, validatorIndex)
			} else {
				result.Online = append(result.Online, validatorIndex)
			}

			if constypes.ValidatorStatus(metadata.Status) == constypes.ActiveExiting {
				result.Exiting = append(result.Exiting, t.IndexTimestamp{
					Index:     validatorIndex,
					Timestamp: uint64(utils.EpochToTime(uint64(metadata.ExitEpoch.Int64)).Unix()),
				})
			} else if constypes.ValidatorStatus(metadata.Status) == constypes.ActiveSlashed {
				result.Slashing = append(result.Slashing, validatorIndex)
			}
		case constypes.ExitedUnslashed, constypes.ExitedSlashed, constypes.WithdrawalPossible, constypes.WithdrawalDone:
			if metadata.Slashed {
				result.Slashed = append(result.Slashed, validatorIndex)
			} else {
				result.Exited = append(result.Exited, validatorIndex)
			}

			if constypes.ValidatorStatus(metadata.Status) == constypes.WithdrawalPossible {
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
			} else if constypes.ValidatorStatus(metadata.Status) == constypes.WithdrawalDone {
				result.Withdrawn = append(result.Withdrawn, validatorIndex)
			}
		}
	}

	return result, nil
}

func (d *DataAccessService) GetValidatorDashboardSyncSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSyncSummaryValidators, error) {
	// possible periods are: all_time, last_30d, last_7d, last_24h, last_1h
	result := &t.VDBSyncSummaryValidators{}
	var resultMutex = &sync.RWMutex{}

	type PastStruct struct {
		Index uint64
		Count uint64
	}

	wg := errgroup.Group{}

	// Get the table name based on the period
	table, _, _, err := d.getTablesForPeriod(period)
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
		currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, err := d.getCurrentAndUpcomingSyncCommittees(latestEpoch)
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
			From(table).
			Limit(1)

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		var epochStart uint64
		err = d.alloyReader.Get(&epochStart, query, args...)
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
			LeftJoin(goqu.I(table).As("r"), goqu.On(goqu.L("sc.validatorindex = r.validator_index"))).
			Where(goqu.L("period >= ? AND period < ? AND validatorindex = ANY(?)", pastSyncPeriodCutoff, currentSyncPeriod, pq.Array(validatorIndices)))

		query, args, err = ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		var validatorIndices []uint64
		err = d.alloyReader.Select(&validatorIndices, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data for past sync committees: %w", err)
		}

		validatorCountMap := make(map[uint64]uint64)
		for _, validatorIndex := range validatorIndices {
			validatorCountMap[validatorIndex]++
		}

		resultMutex.Lock()
		for validatorIndex, count := range validatorCountMap {
			result.Past = append(result.Past, PastStruct{
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

	type GotSlashedStruct struct {
		Index     uint64
		SlashedBy uint64
	}
	type HasSlashedStruct struct {
		Index          uint64
		SlashedIndices []uint64
	}

	// Get the table names based on the period
	table, slashedByCountTable, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, err
	}

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	var queryResult []struct {
		EpochStart     uint64        `db:"epoch_start"`
		EpochEnd       uint64        `db:"epoch_end"`
		ValidatorIndex uint64        `db:"validator_index"`
		SlashedBy      sql.NullInt64 `db:"slashed_by"`
		SlashedAmount  uint32        `db:"slashed_amount"`
	}

	// Build the query
	ds := goqu.Dialect("postgres").Select(
		goqu.L("r.epoch_start"),
		goqu.L("r.epoch_end"),
		goqu.L("r.validator_index"),
		goqu.L("r.slashed_by"),
		goqu.L("COALESCE(s.slashed_amount, 0) AS slashed_amount")).
		From(goqu.T(table).As("r")).
		LeftJoin(goqu.T(slashedByCountTable).As("s"), goqu.On(goqu.L("r.validator_index = s.slashed_by"))).
		Where(goqu.L("(r.slashed_by IS NOT NULL OR s.slashed_amount > 0)"))

	// handle the case when we have a list of validators
	if len(dashboardId.Validators) > 0 {
		ds = ds.
			Where(goqu.L("r.validator_index = ANY(?)", pq.Array(dashboardId.Validators)))
	} else {
		ds = ds.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

		if groupId != t.AllGroups {
			ds = ds.Where(goqu.L("v.group_id = ?", groupId))
		}
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = d.alloyReader.Select(&queryResult, query, args...)
	if err != nil {
		log.Error(err, "error while getting validator dashboard slashed validators list", 0)
		return nil, err
	}

	// Process the data and get the slashing validators
	var slashingValidators []uint64
	for _, queryEntry := range queryResult {
		if queryEntry.SlashedBy.Valid {
			result.GotSlashed = append(result.GotSlashed, GotSlashedStruct{
				Index:     queryEntry.ValidatorIndex,
				SlashedBy: uint64(queryEntry.SlashedBy.Int64),
			})
		}

		if queryEntry.SlashedAmount > 0 {
			slashingValidators = append(slashingValidators, queryEntry.ValidatorIndex)
		}
	}

	if len(slashingValidators) == 0 {
		// We don't have any slashing validators so we can return early
		return result, nil
	}

	// If we have slashing validators then get the validators that got slashed
	proposalSlashings := make(map[uint64][]uint64)
	attestationSlashings := make(map[uint64][]uint64)

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
			Where(goqu.L("bps.block_slot >= ? AND bps.block_slot <= ? AND b.proposer = ANY(?)", slotStart, slotEnd, pq.Array(slashingValidators)))

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Select(&queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data from table %s: %v", table, err)
		}

		for _, queryEntry := range queryResult {
			if _, ok := proposalSlashings[queryEntry.ProposerSlashing]; !ok {
				proposalSlashings[queryEntry.ProposerSlashing] = make([]uint64, 0)
			}
			proposalSlashings[queryEntry.ProposerSlashing] = append(proposalSlashings[queryEntry.ProposerSlashing], queryEntry.ProposerSlashed)
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
			Where(goqu.L("bas.block_slot >= ? AND bas.block_slot <= ? AND b.proposer = ANY(?)", slotStart, slotEnd, pq.Array(slashingValidators)))

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Select(&queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data from table %s: %v", table, err)
		}

		for _, queryEntry := range queryResult {
			inter := intersect.Simple(queryEntry.Attestestation1Indices, queryEntry.Attestestation2Indices)
			if len(inter) == 0 {
				log.WarnWithStackTrace(nil, "No intersection found for attestation violation", 0)
			}
			for _, v := range inter {
				if _, ok := attestationSlashings[queryEntry.Proposer]; !ok {
					attestationSlashings[queryEntry.Proposer] = make([]uint64, 0)
				}
				attestationSlashings[queryEntry.Proposer] = append(attestationSlashings[queryEntry.Proposer], uint64(v.(int64)))
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
		result.HasSlashed = append(result.HasSlashed, HasSlashedStruct{
			Index:          slashingIdx,
			SlashedIndices: slashedIdxs,
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
	table, _, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, err
	}

	// Build the query and get the data
	var queryResult []struct {
		Slot           uint64        `db:"slot"`
		Block          sql.NullInt64 `db:"exec_block_number"`
		Status         string        `db:"status"`
		ValidatorIndex uint64        `db:"validator_index"`
	}

	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("b.slot"),
			goqu.L("b.exec_block_number"),
			goqu.L("b.status"),
			goqu.L("r.validator_index")).
		From(goqu.T(table).As("r")).
		InnerJoin(goqu.L("blocks AS b"), goqu.On(goqu.L("b.epoch >= r.epoch_start AND b.epoch <= r.epoch_end AND r.validator_index = b.proposer")))

	if len(dashboardId.Validators) > 0 {
		ds = ds.
			Where(goqu.L("r.validator_index = ANY(?)", pq.Array(dashboardId.Validators)))
	} else {
		ds = ds.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

		if groupId != t.AllGroups {
			ds = ds.Where(goqu.L("v.group_id = ?", groupId))
		}
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %v", err)
	}

	err = d.alloyReader.Select(&queryResult, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving data from table %s: %v", table, err)
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

	type Proposed struct {
		Index          uint64
		ProposedBlocks []uint64
	}
	type Missed struct {
		Index        uint64
		MissedBlocks []uint64
	}

	for validatorIndex, blockNumbers := range proposedValidatorMap {
		result.Proposed = append(result.Proposed, Proposed{
			Index:          validatorIndex,
			ProposedBlocks: blockNumbers,
		})
	}
	for validatorIndex, slotNumbers := range missedValidatorMap {
		result.Missed = append(result.Missed, Missed{
			Index:        validatorIndex,
			MissedBlocks: slotNumbers,
		})
	}

	return result, nil
}

func (d *DataAccessService) getCurrentAndUpcomingSyncCommittees(latestEpoch uint64) (map[uint64]bool, map[uint64]bool, error) {
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

	err = d.readerDb.Select(&queryResult, query, args...)
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

func (d *DataAccessService) getTablesForPeriod(period enums.TimePeriod) (string, string, int, error) {
	table := ""
	slashedByCountTable := ""
	days := 0

	switch period {
	case enums.TimePeriods.Last24h:
		table = "validator_dashboard_data_rolling_daily"
		slashedByCountTable = "validator_dashboard_data_rolling_daily_slashedby_count"
		days = 1
	case enums.TimePeriods.Last7d:
		table = "validator_dashboard_data_rolling_weekly"
		slashedByCountTable = "validator_dashboard_data_rolling_weekly_slashedby_count"
		days = 7
	case enums.TimePeriods.Last30d:
		table = "validator_dashboard_data_rolling_monthly"
		slashedByCountTable = "validator_dashboard_data_rolling_monthly_slashedby_count"
		days = 30
	case enums.TimePeriods.AllTime:
		table = "validator_dashboard_data_rolling_total"
		slashedByCountTable = "validator_dashboard_data_rolling_total_slashedby_count"
		days = -1
	default:
		return "", "", 0, fmt.Errorf("not-implemented time period: %v", period)
	}

	return table, slashedByCountTable, days, nil
}
