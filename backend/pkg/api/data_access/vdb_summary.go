package dataaccess

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardSummary(dashboardId t.VDBId, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	/* // TODO: implement sorting & paging
	ret := make(map[int64]*t.VDBSummaryTableRow) // map of group id to result row
	retMux := &sync.Mutex{}

	var paging t.Paging

	// retrieve efficiency data for each time period, we cannot do sorting & filtering here as we need access to the whole set
	wg := errgroup.Group{}

	showTotalRow := false

	type queryResult struct {
		GroupId                int64            `db:"group_id"`
		AttestationReward      *decimal.Decimal `db:"attestations_reward"`
		AttestationIdealReward *decimal.Decimal `db:"attestations_ideal_reward"`
		BlocksProposed         *decimal.Decimal `db:"blocks_proposed"`
		BlocksScheduled        *decimal.Decimal `db:"blocks_scheduled"`
		SyncExecuted           *decimal.Decimal `db:"sync_executed"`
		SyncScheduled          *decimal.Decimal `db:"sync_scheduled"`
	}

	searchValidator := -1
	searchGroup := make(map[int]bool)
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
				return []t.VDBSummaryTableRow{}, &paging, nil
			}
			releaseLock()
		} else if number, err := strconv.ParseUint(search, 10, 64); err == nil {
			searchValidator = int(number)
		} else if dashboardId.AggregateGroups {
			// Searching for a group name is not supported when aggregating groups
			return []t.VDBSummaryTableRow{}, &paging, nil
		}
	}

	validators := make([]t.VDBValidator, 0)
	if dashboardId.Validators != nil {
		for _, validator := range dashboardId.Validators {
			if searchValidator != -1 && int(validator) == searchValidator {
				searchGroup[t.DefaultGroupId] = true
			}
			validators = append(validators, validator)
		}

		ret[t.DefaultGroupId] = &t.VDBSummaryTableRow{
			Validators: append([]t.VDBValidator{}, validators...),
		}
	}

	retrieveAndProcessData := func(dashboardId t.VDBIdPrimary, validatorList []t.VDBValidator, aggregateGroups bool, tableName string) (map[int64]float64, error) {
		var queryResult []queryResult

		if len(validatorList) > 0 {
			query := `select $1::smallint AS group_id, attestations_reward, attestations_ideal_reward, blocks_proposed, blocks_scheduled, sync_executed, sync_scheduled FROM (
				select
					SUM(attestations_reward)::decimal AS attestations_reward,
					SUM(attestations_ideal_reward)::decimal AS attestations_ideal_reward,
					SUM(blocks_proposed)::decimal AS blocks_proposed,
					SUM(blocks_scheduled)::decimal AS blocks_scheduled,
					SUM(sync_executed)::decimal AS sync_executed,
					SUM(sync_scheduled)::decimal AS sync_scheduled
				from  %[1]s
				where validator_index = ANY($2)
			) as a;`
			err := d.alloyReader.Select(&queryResult, fmt.Sprintf(query, tableName), t.DefaultGroupId, validatorList)
			if err != nil {
				return nil, fmt.Errorf("error retrieving data from table %s: %v", tableName, err)
			}
		} else {
			queryParams := []interface{}{}
			groupIdQuery := "group_id,"
			if aggregateGroups {
				queryParams = append(queryParams, t.DefaultGroupId)
				groupIdQuery = fmt.Sprintf("$%d::smallint AS group_id,", len(queryParams))
			}

			queryParams = append(queryParams, dashboardId)
			query := fmt.Sprintf(`select group_id, attestations_reward, attestations_ideal_reward, blocks_proposed, blocks_scheduled, sync_executed, sync_scheduled FROM (
				select
					%[1]s
					SUM(attestations_reward)::decimal AS attestations_reward,
					SUM(attestations_ideal_reward)::decimal AS attestations_ideal_reward,
					SUM(blocks_proposed)::decimal AS blocks_proposed,
					SUM(blocks_scheduled)::decimal AS blocks_scheduled,
					SUM(sync_executed)::decimal AS sync_executed,
					SUM(sync_scheduled)::decimal AS sync_scheduled
				from users_val_dashboards_validators
				join %[2]s on %[2]s.validator_index = users_val_dashboards_validators.validator_index
				where dashboard_id = $%[3]d
				group by 1
			) as a;`, groupIdQuery, tableName, len(queryParams))

			err := d.alloyReader.Select(&queryResult, query, queryParams...)
			if err != nil {
				return nil, fmt.Errorf("error retrieving data from table %s: %v", tableName, err)
			}
		}

		data := make(map[int64]float64)
		var totalAttestationReward, totalAttestationIdealReward, totalBlocksProposed, totalBlocksScheduled, totalSyncExecuted, totalSyncScheduled decimal.Decimal
		for _, result := range queryResult {
			if result.AttestationReward != nil {
				totalAttestationReward = totalAttestationReward.Add(*result.AttestationReward)
			}
			if result.AttestationIdealReward != nil {
				totalAttestationIdealReward = totalAttestationIdealReward.Add(*result.AttestationIdealReward)
			}
			if result.BlocksProposed != nil {
				totalBlocksProposed = totalBlocksProposed.Add(*result.BlocksProposed)
			}
			if result.BlocksScheduled != nil {
				totalBlocksScheduled = totalBlocksScheduled.Add(*result.BlocksScheduled)
			}
			if result.SyncExecuted != nil {
				totalSyncExecuted = totalSyncExecuted.Add(*result.SyncExecuted)
			}
			if result.SyncScheduled != nil {
				totalSyncScheduled = totalSyncScheduled.Add(*result.SyncScheduled)
			}

			var attestationEfficiency, proposerEfficiency, syncEfficiency sql.NullFloat64
			if result.AttestationReward != nil && result.AttestationIdealReward != nil {
				attestationEfficiency.Float64 = result.AttestationReward.Div(*result.AttestationIdealReward).InexactFloat64()
				attestationEfficiency.Valid = true
			}
			if result.BlocksProposed != nil && result.BlocksScheduled != nil {
				proposerEfficiency.Float64 = result.BlocksProposed.Div(*result.BlocksScheduled).InexactFloat64()
				proposerEfficiency.Valid = true
			}
			if result.SyncExecuted != nil && result.SyncScheduled != nil {
				syncEfficiency.Float64 = result.SyncExecuted.Div(*result.SyncScheduled).InexactFloat64()
				syncEfficiency.Valid = true
			}
			data[result.GroupId] = d.calculateTotalEfficiency(attestationEfficiency, proposerEfficiency, syncEfficiency)
		}

		var totalAttestationEfficiency, totalProposerEfficiency, totalSyncEfficiency sql.NullFloat64
		if !totalAttestationIdealReward.IsZero() {
			totalAttestationEfficiency.Float64 = totalAttestationReward.Div(totalAttestationIdealReward).InexactFloat64()
			totalAttestationEfficiency.Valid = true
		}
		if !totalBlocksScheduled.IsZero() {
			totalProposerEfficiency.Float64 = totalBlocksProposed.Div(totalBlocksScheduled).InexactFloat64()
			totalProposerEfficiency.Valid = true
		}
		if !totalSyncScheduled.IsZero() {
			totalSyncEfficiency.Float64 = totalSyncExecuted.Div(totalSyncScheduled).InexactFloat64()
			totalSyncEfficiency.Valid = true
		}
		data[t.AllGroups] = d.calculateTotalEfficiency(totalAttestationEfficiency, totalProposerEfficiency, totalSyncEfficiency)

		return data, nil
	}

	if len(validators) == 0 { // retrieve the validators & groups from the dashboard table
		wg.Go(func() error {
			type validatorsPerGroup struct {
				GroupId        int64          `db:"group_id"`
				GroupName      string         `db:"group_name"`
				ValidatorIndex t.VDBValidator `db:"validator_index"`
			}

			var queryResult []validatorsPerGroup

			queryParams := []interface{}{}
			groupIdQuery := "v.group_id,"
			if dashboardId.AggregateGroups {
				queryParams = append(queryParams, t.DefaultGroupId)
				groupIdQuery = fmt.Sprintf("$%d::smallint AS group_id,", len(queryParams))
			}

			if search != "" {
				queryParams = append(queryParams, dashboardId.Id)
				query := fmt.Sprintf(`
				SELECT
					%s
					g.name AS group_name,
					v.validator_index
				FROM users_val_dashboards_validators v
				INNER JOIN users_val_dashboards_groups g ON v.group_id = g.id AND v.dashboard_id = g.dashboard_id
				WHERE v.dashboard_id = $%d
				ORDER BY v.group_id, v.validator_index`, groupIdQuery, len(queryParams))

				err := d.alloyReader.Select(&queryResult, query, queryParams...)
				if err != nil {
					return fmt.Errorf("error retrieving validator groups for dashboard: %v", err)
				}

				prefixSearch := strings.ToLower(search)
				for _, result := range queryResult {
					if searchValidator != -1 && result.ValidatorIndex == uint64(searchValidator) ||
						strings.HasPrefix(strings.ToLower(result.GroupName), prefixSearch) {
						searchGroup[int(result.GroupId)] = true
					}
				}
			} else {
				queryParams = append(queryParams, dashboardId.Id)
				query := fmt.Sprintf(`
				SELECT
					%s
					v.validator_index
				FROM users_val_dashboards_validators v
				WHERE dashboard_id = $%d
				ORDER BY group_id, validator_index`, groupIdQuery, len(queryParams))

				err := d.alloyReader.Select(&queryResult, query, queryParams...)
				if err != nil {
					return fmt.Errorf("error retrieving validator groups for dashboard: %v", err)
				}
			}

			groupCountMap := make(map[int64]bool)
			for _, result := range queryResult {
				groupCountMap[result.GroupId] = true
				if len(groupCountMap) > 1 {
					showTotalRow = true
				}
			}

			retMux.Lock()
			if showTotalRow {
				if ret[t.AllGroups] == nil {
					ret[t.AllGroups] = &t.VDBSummaryTableRow{
						GroupId: t.AllGroups,
					}
				}

				if ret[t.AllGroups].Validators == nil {
					ret[t.AllGroups].Validators = make([]t.VDBValidator, 0, 10)
				}
			}

			for _, result := range queryResult {
				if showTotalRow && len(ret[t.AllGroups].Validators) < 10 {
					ret[t.AllGroups].Validators = append(ret[t.AllGroups].Validators, result.ValidatorIndex)
				}

				if _, ok := searchGroup[int(result.GroupId)]; len(searchGroup) > 0 && !ok {
					// We are searching for groups and this group is not in the search list
					continue
				}

				if ret[result.GroupId] == nil {
					ret[result.GroupId] = &t.VDBSummaryTableRow{
						GroupId: result.GroupId,
					}
				}

				if ret[result.GroupId].Validators == nil {
					ret[result.GroupId].Validators = make([]t.VDBValidator, 0, 10)
				}

				if len(ret[result.GroupId].Validators) < 10 {
					ret[result.GroupId].Validators = append(ret[result.GroupId].Validators, result.ValidatorIndex)
				}
			}
			retMux.Unlock()
			return nil
		})
	}

	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, dashboardId.AggregateGroups, "validator_dashboard_data_rolling_daily")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.Last24h = efficiency
		}
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, dashboardId.AggregateGroups, "validator_dashboard_data_rolling_weekly")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.Last7d = efficiency
		}
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, dashboardId.AggregateGroups, "validator_dashboard_data_rolling_monthly")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.Last30d = efficiency
		}
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, dashboardId.AggregateGroups, "validator_dashboard_data_rolling_total")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.AllTime = efficiency
		}
		return nil
	})
	err := wg.Wait()
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving validator dashboard summary data: %v", err)
	}

	if search != "" && len(searchGroup) == 0 {
		return []t.VDBSummaryTableRow{}, &paging, nil
	}

	for groupId := range ret {
		if _, ok := searchGroup[int(groupId)]; len(searchGroup) > 0 && !ok && groupId != t.AllGroups {
			delete(ret, groupId)
		}
	}
	if !showTotalRow {
		delete(ret, t.AllGroups)
	}

	retArr := make([]t.VDBSummaryTableRow, 0, len(ret))

	for _, row := range ret {
		retArr = append(retArr, *row)
	}

	sort.Slice(retArr, func(i, j int) bool {
		return retArr[i].GroupId < retArr[j].GroupId
	})

	paging.TotalCount = uint64(len(retArr))

	return retArr, &paging, nil */
	return d.dummy.GetValidatorDashboardSummary(dashboardId, period, cursor, colSort, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardGroupSummary(dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBGroupSummaryData, error) {
	// TODO: implement data retrieval for the following new field
	// SyncCommitteeCount
	// TotalMissedRewards

	ret := &t.VDBGroupSummaryData{}

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	table := ""
	slashedByCountTable := ""
	days := 0

	switch period {
	case enums.Last24h:
		table = "validator_dashboard_data_rolling_daily"
		slashedByCountTable = "validator_dashboard_data_rolling_daily_slashedby_count"
		days = 1
	case enums.Last7d:
		table = "validator_dashboard_data_rolling_weekly"
		slashedByCountTable = "validator_dashboard_data_rolling_weekly_slashedby_count"
		days = 7
	case enums.Last30d:
		table = "validator_dashboard_data_rolling_monthly"
		slashedByCountTable = "validator_dashboard_data_rolling_monthly_slashedby_count"
		days = 30
	case enums.AllTime:
		table = "validator_dashboard_data_rolling_total"
		slashedByCountTable = "validator_dashboard_data_rolling_total_slashedby_count"
		days = -1
	default:
		return nil, fmt.Errorf("not-implemented time period: %v", period)
	}

	query := `select
			users_val_dashboards_validators.validator_index,
			epoch_start,
			COALESCE(attestations_source_reward, 0) as attestations_source_reward,
			COALESCE(attestations_target_reward, 0) as attestations_target_reward,
			COALESCE(attestations_head_reward, 0) as attestations_head_reward,
			COALESCE(attestations_inactivity_reward, 0) as attestations_inactivity_reward,
			COALESCE(attestations_inclusion_reward, 0) as attestations_inclusion_reward,
			COALESCE(attestations_reward, 0) as attestations_reward,
			COALESCE(attestations_ideal_source_reward, 0) as attestations_ideal_source_reward,
			COALESCE(attestations_ideal_target_reward, 0) as attestations_ideal_target_reward,
			COALESCE(attestations_ideal_head_reward, 0) as attestations_ideal_head_reward,
			COALESCE(attestations_ideal_inactivity_reward, 0) as attestations_ideal_inactivity_reward,
			COALESCE(attestations_ideal_inclusion_reward, 0) as attestations_ideal_inclusion_reward,
			COALESCE(attestations_ideal_reward, 0) as attestations_ideal_reward,
			COALESCE(attestations_scheduled, 0) as attestations_scheduled,
			COALESCE(attestations_executed, 0) as attestations_executed,
			COALESCE(attestation_head_executed, 0) as attestation_head_executed,
			COALESCE(attestation_source_executed, 0) as attestation_source_executed,
			COALESCE(attestation_target_executed, 0) as attestation_target_executed,
			COALESCE(blocks_scheduled, 0) as blocks_scheduled,
			COALESCE(blocks_proposed, 0) as blocks_proposed,
			COALESCE(blocks_cl_reward, 0) as blocks_cl_reward,
			COALESCE(blocks_el_reward, 0) as blocks_el_reward,
			COALESCE(sync_scheduled, 0) as sync_scheduled,
			COALESCE(sync_executed, 0) as sync_executed,
			COALESCE(sync_rewards, 0) as sync_rewards,
			%[1]s.slashed_by IS NOT NULL AS slashed_in_period,
			COALESCE(%[2]s.slashed_amount, 0) AS slashed_amount,
			COALESCE(deposits_count, 0) as deposits_count,
			COALESCE(withdrawals_count, 0) as withdrawals_count,
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
			COALESCE(attestations_source_reward, 0) as attestations_source_reward,
			COALESCE(attestations_target_reward, 0) as attestations_target_reward,
			COALESCE(attestations_head_reward, 0) as attestations_head_reward,
			COALESCE(attestations_inactivity_reward, 0) as attestations_inactivity_reward,
			COALESCE(attestations_inclusion_reward, 0) as attestations_inclusion_reward,
			COALESCE(attestations_reward, 0) as attestations_reward,
			COALESCE(attestations_ideal_source_reward, 0) as attestations_ideal_source_reward,
			COALESCE(attestations_ideal_target_reward, 0) as attestations_ideal_target_reward,
			COALESCE(attestations_ideal_head_reward, 0) as attestations_ideal_head_reward,
			COALESCE(attestations_ideal_inactivity_reward, 0) as attestations_ideal_inactivity_reward,
			COALESCE(attestations_ideal_inclusion_reward, 0) as attestations_ideal_inclusion_reward,
			COALESCE(attestations_ideal_reward, 0) as attestations_ideal_reward,
			COALESCE(attestations_scheduled, 0) as attestations_scheduled,
			COALESCE(attestations_executed, 0) as attestations_executed,
			COALESCE(attestation_head_executed, 0) as attestation_head_executed,
			COALESCE(attestation_source_executed, 0) as attestation_source_executed,
			COALESCE(attestation_target_executed, 0) as attestation_target_executed,
			COALESCE(blocks_scheduled, 0) as blocks_scheduled,
			COALESCE(blocks_proposed, 0) as blocks_proposed,
			COALESCE(blocks_cl_reward, 0) as blocks_cl_reward,
			COALESCE(blocks_el_reward, 0) as blocks_el_reward,
			COALESCE(sync_scheduled, 0) as sync_scheduled,
			COALESCE(sync_executed, 0) as sync_executed,
			COALESCE(sync_rewards, 0) as sync_rewards,
			%[1]s.slashed_by IS NOT NULL AS slashed_in_period,
			COALESCE(%[2]s.slashed_amount, 0) AS slashed_amount,
			COALESCE(deposits_count, 0) as deposits_count,
			COALESCE(withdrawals_count, 0) as withdrawals_count,
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
		ValidatorIndex                    uint32 `db:"validator_index"`
		EpochStart                        int    `db:"epoch_start"`
		AttestationSourceReward           int64  `db:"attestations_source_reward"`
		AttestationTargetReward           int64  `db:"attestations_target_reward"`
		AttestationHeadReward             int64  `db:"attestations_head_reward"`
		AttestationInactivitytReward      int64  `db:"attestations_inactivity_reward"`
		AttestationInclusionReward        int64  `db:"attestations_inclusion_reward"`
		AttestationReward                 int64  `db:"attestations_reward"`
		AttestationIdealSourceReward      int64  `db:"attestations_ideal_source_reward"`
		AttestationIdealTargetReward      int64  `db:"attestations_ideal_target_reward"`
		AttestationIdealHeadReward        int64  `db:"attestations_ideal_head_reward"`
		AttestationIdealInactivitytReward int64  `db:"attestations_ideal_inactivity_reward"`
		AttestationIdealInclusionReward   int64  `db:"attestations_ideal_inclusion_reward"`
		AttestationIdealReward            int64  `db:"attestations_ideal_reward"`

		AttestationsScheduled     int64 `db:"attestations_scheduled"`
		AttestationsExecuted      int64 `db:"attestations_executed"`
		AttestationHeadExecuted   int64 `db:"attestation_head_executed"`
		AttestationSourceExecuted int64 `db:"attestation_source_executed"`
		AttestationTargetExecuted int64 `db:"attestation_target_executed"`

		BlocksScheduled uint32          `db:"blocks_scheduled"`
		BlocksProposed  uint32          `db:"blocks_proposed"`
		BlocksClReward  uint64          `db:"blocks_cl_reward"`
		BlocksElReward  decimal.Decimal `db:"blocks_el_reward"`

		SyncScheduled uint32 `db:"sync_scheduled"`
		SyncExecuted  uint32 `db:"sync_executed"`
		SyncRewards   int64  `db:"sync_rewards"`

		SlashedInPeriod bool   `db:"slashed_in_period"`
		SlashedAmount   uint32 `db:"slashed_amount"`

		DepositsCount uint32 `db:"deposits_count"`

		WithdrawalsCount uint32 `db:"withdrawals_count"`

		BlockChance            float64 `db:"blocks_expected"`
		SyncCommitteesExpected float64 `db:"sync_committees_expected"`

		InclusionDelaySum int64 `db:"inclusion_delay_sum"`
	}

	var rows []*queryResult
	var err error

	if len(validators) > 0 {
		err = d.alloyReader.Select(&rows, fmt.Sprintf(query, table, slashedByCountTable), validators)
	} else {
		err = d.alloyReader.Select(&rows, fmt.Sprintf(query, table, slashedByCountTable), dashboardId, groupId)
	}

	if err != nil {
		return nil, err
	}

	totalAttestationRewards := int64(0)
	totalIdealAttestationRewards := int64(0)
	totalBlockChance := float64(0)
	totalInclusionDelaySum := int64(0)
	totalInclusionDelayDivisor := int64(0)
	totalSyncExpected := float64(0)
	totalProposals := uint32(0)

	validatorArr := make([]t.VDBValidator, 0)
	startEpoch := math.MaxUint32
	for _, row := range rows {
		if row.EpochStart < startEpoch { // set the start epoch for querying the EL APR
			startEpoch = row.EpochStart
		}
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
			ret.ProposalValidators = append(ret.ProposalValidators, t.VDBValidator(row.ValidatorIndex))
		}

		ret.SyncCommittee.StatusCount.Success += uint64(row.SyncExecuted)
		ret.SyncCommittee.StatusCount.Failed += uint64(row.SyncScheduled) - uint64(row.SyncExecuted)

		if row.SyncScheduled > 0 {
			if ret.SyncCommittee.Validators == nil {
				ret.SyncCommittee.Validators = make([]t.VDBValidator, 0, 10)
			}
			ret.SyncCommittee.Validators = append(ret.SyncCommittee.Validators, t.VDBValidator(row.ValidatorIndex))
		}

		if row.SlashedInPeriod {
			ret.Slashings.StatusCount.Failed++
			ret.Slashings.Validators = append(ret.Slashings.Validators, t.VDBValidator(row.ValidatorIndex))
		}
		ret.Slashings.StatusCount.Success += uint64(row.SlashedAmount)

		totalBlockChance += row.BlockChance
		totalInclusionDelaySum += row.InclusionDelaySum
		totalSyncExpected += row.SyncCommitteesExpected

		if row.InclusionDelaySum > 0 {
			totalInclusionDelayDivisor += row.AttestationsScheduled
		}
	}

	ret.Income.El, ret.Apr.El, ret.Income.Cl, ret.Apr.Cl, err = d.internal_getElClAPR(validatorArr, days)
	if err != nil {
		return nil, err
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
		ret.Luck.Proposal.Average = time.Duration((luckDays / totalBlockChance) * 24 * float64(time.Hour))
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
		ret.Luck.Sync.Average = time.Duration((luckDays / totalSyncExpected) * 24 * float64(time.Hour))
	}

	if totalInclusionDelayDivisor > 0 {
		ret.AttestationAvgInclDist = 1.0 + float64(totalInclusionDelaySum)/float64(totalInclusionDelayDivisor)
	} else {
		ret.AttestationAvgInclDist = 0
	}

	return ret, nil
}

func (d *DataAccessService) internal_getElClAPR(validators []t.VDBValidator, days int) (elIncome decimal.Decimal, elAPR float64, clIncome decimal.Decimal, clAPR float64, err error) {
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

func (d *DataAccessService) GetValidatorDashboardSummaryChart(dashboardId t.VDBId) (*t.ChartData[int, float64], error) {
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
				SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
				SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
				SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
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
				SUM(d.attestations_reward)::decimal / NULLIF(SUM(d.attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
				SUM(d.blocks_proposed)::decimal / NULLIF(SUM(d.blocks_scheduled)::decimal, 0) AS proposer_efficiency,
				SUM(d.sync_executed)::decimal / NULLIF(SUM(d.sync_scheduled)::decimal, 0) AS sync_efficiency
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

// allowed periods are: all_time, last_24h, last_7d, last_30d
func (d *DataAccessService) GetValidatorDashboardValidatorIndices(dashboardId t.VDBId, groupId int64, duty enums.ValidatorDuty, period enums.TimePeriod) ([]t.VDBValidator, error) {
	var validators []t.VDBValidator

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	if dashboardId.Validators == nil {
		// Get the validators in case a dashboard id is provided
		validatorsQuery := `
		SELECT 
			validator_index
		FROM users_val_dashboards_validators
		WHERE dashboard_id = $1
		`
		validatorsParams := []interface{}{dashboardId.Id}

		if groupId != t.AllGroups {
			validatorsQuery += " AND group_id = $2"
			validatorsParams = append(validatorsParams, groupId)
		}
		err := d.alloyReader.Select(&validators, validatorsQuery, validatorsParams...)
		if err != nil {
			return nil, err
		}
	} else {
		// In case a list of validators is provided use them
		validators = dashboardId.Validators
	}

	if len(validators) == 0 {
		// Return if there are no validators
		return []t.VDBValidator{}, nil
	}

	if duty == enums.ValidatorDuties.None {
		// If we don't need to filter by duty return all validators in the dashboard and group
		return validators, nil
	}

	// Get the table name based on the period
	tableName := ""
	switch period {
	case enums.TimePeriods.AllTime:
		tableName = "validator_dashboard_data_rolling_total"
	case enums.TimePeriods.Last24h:
		tableName = "validator_dashboard_data_rolling_daily"
	case enums.TimePeriods.Last7d:
		tableName = "validator_dashboard_data_rolling_weekly"
	case enums.TimePeriods.Last30d:
		tableName = "validator_dashboard_data_rolling_monthly"
	}

	// Get the column condition based on the duty
	columnCond := ""
	switch duty {
	case enums.ValidatorDuties.Sync:
		columnCond = "sync_scheduled > 0"
	case enums.ValidatorDuties.Proposal:
		columnCond = "blocks_scheduled > 0"
	case enums.ValidatorDuties.Slashed:
		columnCond = "slashed_by IS NOT NULL"
	}

	// Get ALL validator indices for the given filters
	query := fmt.Sprintf(`
		SELECT
			validator_index
		FROM %s
		WHERE validator_index = ANY($1) AND %s`, tableName, columnCond)

	var result []t.VDBValidator
	err := d.alloyReader.Select(&result, query, pq.Array(validators))
	return result, err
}
