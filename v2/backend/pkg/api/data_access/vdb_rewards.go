package dataaccess

import (
	"database/sql"
	"fmt"
	"math/big"
	"slices"
	"sort"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardRewards(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	var paging t.Paging

	// Initialize the cursor
	var currentCursor t.RewardsCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.RewardsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as WithdrawalsCursor: %w", err)
		}
	}

	// Prepare the sorting
	sortSearchDirection := ">="
	if (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse()) {
		sortSearchDirection = "<="
	}

	// Analyze the search term
	// indexSearch := int64(-1)
	// epochSearch := int64(-1)
	// if search != "" {
	// 	if utils.IsHash(search) {
	// 		// Ensure that we have a "0x" prefix for the search term
	// 		if !strings.HasPrefix(search, "0x") {
	// 			search = "0x" + search
	// 		}
	// 		search = strings.ToLower(search)
	// 		if utils.IsHash(search) {
	// 			// Get the current validator state to convert pubkey to index
	// 			validatorMapping, releaseLock, err := d.services.GetCurrentValidatorMapping()
	// 			defer releaseLock()
	// 			if err != nil {
	// 				return nil, nil, err
	// 			}
	// 			if index, ok := validatorMapping.ValidatorIndices[search]; ok {
	// 				indexSearch = int64(*index)
	// 			} else {
	// 				// No validator index for pubkey found, return empty results
	// 				return nil, &paging, nil
	// 			}
	// 		}
	// 	} else if number, err := strconv.ParseUint(search, 10, 64); err == nil {
	// 		indexSearch = int64(number)
	// 		epochSearch = int64(number)
	// 	}
	// }

	type ValidatorGroupInfo struct {
		GroupId   uint64
		GroupName string
	}
	validatorGroupMap := make(map[uint64]ValidatorGroupInfo)
	var validators []uint64
	if dashboardId.Validators == nil {
		// Get the validators and their groups in case a dashboard id is provided
		queryResult := []struct {
			ValidatorIndex uint64 `db:"validator_index"`
			GroupId        uint64 `db:"group_id"`
			GroupName      string `db:"group_name"`
		}{}

		validatorsQuery := `
		SELECT 
			v.validator_index,
			v.group_id,
			g.name AS group_name
		FROM users_val_dashboards_validators v
		INNER JOIN users_val_dashboards_groups g ON v.group_id = g.id AND v.dashboard_id = g.dashboard_id
		WHERE v.dashboard_id = $1
		`

		err := d.alloyReader.Select(&queryResult, validatorsQuery, dashboardId.Id)
		if err != nil {
			return nil, nil, err
		}

		for _, res := range queryResult {
			validatorGroupMap[res.ValidatorIndex] = ValidatorGroupInfo{
				GroupId:   res.GroupId,
				GroupName: res.GroupName,
			}
			validators = append(validators, res.ValidatorIndex)
		}
	} else {
		// In case a list of validators is provided set the group to the default id
		for _, validator := range dashboardId.Validators {
			validatorGroupMap[validator.Index] = ValidatorGroupInfo{
				GroupId:   t.DefaultGroupId,
				GroupName: t.DefaultGroupName,
			}
			validators = append(validators, validator.Index)
		}
	}

	if len(validators) == 0 {
		// Return if there are no validators
		return nil, &paging, nil
	}

	// Get the rewards for the validators
	type ValidatorInfo struct {
		ValidatorIndex        uint64          `db:"validator_index"`
		Epoch                 uint64          `db:"epoch"`
		ClRewards             int64           `db:"cl_rewards"`
		ElRewards             decimal.Decimal `db:"el_rewards"`
		AttestationsScheduled uint64          `db:"attestations_scheduled"`
		AttestationsExecuted  uint64          `db:"attestations_executed"`
		BlocksScheduled       uint64          `db:"blocks_scheduled"`
		BlocksProposed        uint64          `db:"blocks_proposed"`
		SyncScheduled         uint64          `db:"sync_scheduled"`
		SyncExecuted          uint64          `db:"sync_executed"`
		SlashedViolation      uint64          `db:"slashed_violation"`
	}

	queryResult := []ValidatorInfo{}

	queryParams := []interface{}{}
	rewardsQuery := `
		SELECT
			validator_index,
			epoch,
			COALESCE(attestations_reward, 0) + COALESCE(blocks_cl_reward, 0) +
			COALESCE(sync_rewards, 0) + COALESCE(slasher_reward, 0) AS cl_rewards,
			COALESCE(blocks_el_reward, 0) AS el_rewards,
			COALESCE(attestations_scheduled, 0) AS attestations_scheduled,
			COALESCE(attestations_executed, 0) AS attestations_executed,
			COALESCE(blocks_scheduled, 0) AS blocks_scheduled,
			COALESCE(blocks_proposed, 0) AS blocks_proposed,
			COALESCE(sync_scheduled, 0) AS sync_scheduled,
			COALESCE(sync_executed, 0) AS sync_executed,
			COALESCE(slashed_violation, 0) AS slashed_violation
		FROM validator_dashboard_data_epoch
		`

	queryParams = append(queryParams, pq.Array(validators))
	whereQuery := fmt.Sprintf("WHERE validator_index = ANY($%d)", len(queryParams))
	if currentCursor.IsValid() {
		queryParams = append(queryParams, currentCursor.Epoch)
		whereQuery += fmt.Sprintf(" AND epoch %s $%d", sortSearchDirection, len(queryParams))
	}

	rewardsQuery += whereQuery
	err = d.alloyReader.Select(&queryResult, rewardsQuery, queryParams...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &paging, nil
		}
		return nil, nil, fmt.Errorf("error getting rewards for validators: %+v: %w", validators, err)
	}

	epochGroupMap := make(map[uint64]map[uint64]*ValidatorInfo)
	for _, res := range queryResult {
		if _, ok := epochGroupMap[res.Epoch]; !ok {
			epochGroupMap[res.Epoch] = make(map[uint64]*ValidatorInfo)
		}
		if _, ok := epochGroupMap[res.Epoch][validatorGroupMap[res.ValidatorIndex].GroupId]; !ok {
			epochGroupMap[res.Epoch][validatorGroupMap[res.ValidatorIndex].GroupId] = &ValidatorInfo{}
		}
		epochGroup := epochGroupMap[res.Epoch][validatorGroupMap[res.ValidatorIndex].GroupId]

		epochGroup.ClRewards += res.ClRewards
		epochGroup.ElRewards = epochGroup.ElRewards.Add(res.ElRewards)
		epochGroup.AttestationsScheduled += res.AttestationsScheduled
		epochGroup.AttestationsExecuted += res.AttestationsExecuted
		epochGroup.BlocksScheduled += res.BlocksScheduled
		epochGroup.BlocksProposed += res.BlocksProposed
		epochGroup.SyncScheduled += res.SyncScheduled
		epochGroup.SyncExecuted += res.SyncExecuted
		epochGroup.SlashedViolation += res.SlashedViolation
	}

	result := make([]t.VDBRewardsTableRow, 0)
	for epoch, groupMap := range epochGroupMap {
		for groupId, info := range groupMap {
			duty := t.VDBRewardesTableDuty{}
			if info.AttestationsScheduled > 0 {
				attestationPercentage := float64(info.AttestationsExecuted) / float64(info.AttestationsScheduled)
				duty.Attestation = &attestationPercentage
			}
			if info.BlocksScheduled > 0 {
				ProposalPercentage := float64(info.BlocksProposed) / float64(info.BlocksScheduled)
				duty.Proposal = &ProposalPercentage
			}
			if info.SyncScheduled > 0 {
				SyncPercentage := float64(info.SyncExecuted) / float64(info.SyncScheduled)
				duty.Sync = &SyncPercentage
			}
			if info.SlashedViolation > 0 {
				duty.Slashing = &info.SlashedViolation
			}
			reward := t.ClElValue[decimal.Decimal]{
				El: info.ElRewards,
				Cl: utils.GWeiToWei(big.NewInt(info.ClRewards)),
			}
			if duty.Attestation != nil || duty.Proposal != nil || duty.Sync != nil || duty.Slashing != nil {
				result = append(result, t.VDBRewardsTableRow{
					Epoch:   epoch,
					Duty:    duty,
					GroupId: groupId,
					Reward:  reward,
				})
			}
		}
	}

	// Sort the result first by epoch then by group id
	sort.Slice(result, func(i, j int) bool {
		if colSort.Desc {
			if result[i].Epoch == result[j].Epoch {
				return result[i].GroupId > result[j].GroupId
			}
			return result[i].Epoch > result[j].Epoch
		}
		if result[i].Epoch == result[j].Epoch {
			return result[i].GroupId < result[j].GroupId
		}
		return result[i].Epoch < result[j].Epoch
	})

	// Find the cursor and cut away the data that is not needed
	if currentCursor.IsValid() {
		cursorIndex := -1
		for i, row := range result {
			if row.Epoch == currentCursor.Epoch && row.GroupId == currentCursor.GroupId {
				cursorIndex = i
				break
			}
		}
		if cursorIndex == -1 {
			return nil, nil, fmt.Errorf("cursor not found in data: %+v", currentCursor)
		}
		result = result[cursorIndex+1:]
	}

	// Flag if above limit
	moreDataFlag := len(result) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return result, &paging, nil
	}

	// Remove the last entries from data
	if moreDataFlag {
		result = result[:limit]
	}

	// Reverse the data if the cursor is reversed to correct it to the requested direction
	if currentCursor.IsReverse() {
		slices.Reverse(result)
	}

	p, err := utils.GetPagingFromData(result, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return result, p, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupRewards(dashboardId t.VDBId, groupId int64, epoch uint64) (*t.VDBGroupRewardsData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupRewards(dashboardId, groupId, epoch)
}

func (d *DataAccessService) GetValidatorDashboardRewardsChart(dashboardId t.VDBId) (*t.ChartData[int, decimal.Decimal], error) {
	// TODO @recy21
	// bar chart for the CL and EL rewards for each group for each epoch. NO series for all groups combined
	// series id is group id, series property is 'cl' or 'el'
	return d.dummy.GetValidatorDashboardRewardsChart(dashboardId)
}

func (d *DataAccessService) GetValidatorDashboardDuties(dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDuties(dashboardId, epoch, groupId, cursor, colSort, search, limit)
}
