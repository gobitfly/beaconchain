package dataaccess

import (
	"database/sql"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	utilMath "github.com/protolambda/zrnt/eth2/util/math"
	"github.com/shopspring/decimal"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
)

func (d *DataAccessService) GetValidatorDashboardInfo(dashboardId t.VDBIdPrimary) (*t.DashboardInfo, error) {
	result := &t.DashboardInfo{}

	err := d.alloyReader.Get(result, `
		SELECT 
			id, 
			user_id
		FROM users_val_dashboards
		WHERE id = $1
	`, dashboardId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: dashboard with id %v not found", ErrNotFound, dashboardId)
	}
	return result, err
}

func (d *DataAccessService) GetValidatorDashboardInfoByPublicId(publicDashboardId t.VDBIdPublic) (*t.DashboardInfo, error) {
	result := &t.DashboardInfo{}

	err := d.alloyReader.Get(result, `
		SELECT 
			uvd.id,
			uvd.user_id
		FROM users_val_dashboards_sharing uvds
		LEFT JOIN users_val_dashboards uvd ON uvd.id = uvds.dashboard_id
		WHERE uvds.public_id = $1
	`, publicDashboardId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: public id %v not found", ErrNotFound, publicDashboardId)
	}
	return result, err
}

// param validators: slice of validator public keys or indices
func (d *DataAccessService) GetValidatorsFromSlices(indices []t.VDBValidator, publicKeys []string) ([]t.VDBValidator, error) {
	if len(indices) == 0 && len(publicKeys) == 0 {
		return []t.VDBValidator{}, nil
	}

	mapping, release, err := d.services.GetCurrentValidatorMapping()
	defer release()
	if err != nil {
		return nil, err
	}

	validators := make(map[t.VDBValidator]bool, 0)
	for _, pubkey := range publicKeys {
		if v, ok := mapping.ValidatorIndices[pubkey]; ok {
			validators[v] = true
		}
	}
	for _, index := range indices {
		if index < t.VDBValidator(len(mapping.ValidatorPubkeys)) {
			validators[index] = true
		}
	}

	result := maps.Keys(validators)

	return result, nil
}

func (d *DataAccessService) GetUserDashboards(userId uint64) (*t.UserDashboardsData, error) {
	result := &t.UserDashboardsData{}

	dbReturn := []struct {
		Id           uint64         `db:"id"`
		Name         string         `db:"name"`
		PublicId     sql.NullString `db:"public_id"`
		PublicName   sql.NullString `db:"public_name"`
		SharedGroups sql.NullBool   `db:"shared_groups"`
	}{}

	// Get the validator dashboards including the public ones
	err := d.alloyReader.Select(&dbReturn, `
		SELECT 
			uvd.id,
			uvd.name,
			uvds.public_id,
			uvds.name AS public_name,
			uvds.shared_groups
		FROM users_val_dashboards uvd
		LEFT JOIN users_val_dashboards_sharing uvds ON uvd.id = uvds.dashboard_id
		WHERE uvd.user_id = $1
	`, userId)
	if err != nil {
		return nil, err
	}

	// Fill the result
	validatorDashboardMap := make(map[uint64]*t.ValidatorDashboard, 0)
	for _, row := range dbReturn {
		if _, ok := validatorDashboardMap[row.Id]; !ok {
			validatorDashboardMap[row.Id] = &t.ValidatorDashboard{
				Id:        row.Id,
				Name:      row.Name,
				PublicIds: []t.VDBPublicId{},
			}
		}
		if row.PublicId.Valid {
			result := t.VDBPublicId{}
			result.PublicId = row.PublicId.String
			result.Name = row.PublicName.String
			result.ShareSettings.GroupNames = row.SharedGroups.Bool

			validatorDashboardMap[row.Id].PublicIds = append(validatorDashboardMap[row.Id].PublicIds, result)
		}
	}
	for _, validatorDashboard := range validatorDashboardMap {
		result.ValidatorDashboards = append(result.ValidatorDashboards, *validatorDashboard)
	}

	// Get the account dashboards
	err = d.alloyReader.Select(&result.AccountDashboards, `
		SELECT 
			id,
			name
		FROM users_acc_dashboards
		WHERE user_id = $1
	`, userId)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DataAccessService) CreateValidatorDashboard(userId uint64, name string, network uint64) (*t.VDBPostReturnData, error) {
	result := &t.VDBPostReturnData{}

	tx, err := d.alloyWriter.Beginx()
	if err != nil {
		return nil, fmt.Errorf("error starting db transactions to create a validator dashboard: %w", err)
	}
	defer utils.Rollback(tx)

	// Create validator dashboard for user
	err = tx.Get(result, `
		INSERT INTO users_val_dashboards (user_id, network, name)
			VALUES ($1, $2, $3)
		RETURNING id, user_id, name, network, (EXTRACT(epoch FROM created_at))::BIGINT as created_at
	`, userId, network, name)
	if err != nil {
		return nil, err
	}

	// Create a default group for the new dashboard
	_, err = tx.Exec(`
		INSERT INTO users_val_dashboards_groups (dashboard_id, name)
			VALUES ($1, $2)
	`, result.Id, t.DefaultGroupName)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing tx to create a validator dashboard: %w", err)
	}

	return result, nil
}

func (d *DataAccessService) RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error {
	tx, err := d.alloyWriter.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions to remove a validator dashboard: %w", err)
	}
	defer utils.Rollback(tx)

	// Delete the dashboard
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards WHERE id = $1
	`, dashboardId)
	if err != nil {
		return err
	}

	// Delete all groups for the dashboard
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards_groups WHERE dashboard_id = $1
	`, dashboardId)
	if err != nil {
		return err
	}

	// Delete all validators for the dashboard
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards_validators WHERE dashboard_id = $1
	`, dashboardId)
	if err != nil {
		return err
	}

	// Delete all shared dashboards for the dashboard
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards_sharing WHERE dashboard_id = $1
	`, dashboardId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing tx to remove a validator dashboard: %w", err)
	}
	return nil
}

func (d *DataAccessService) GetValidatorDashboardOverview(dashboardId t.VDBId) (*t.VDBOverviewData, error) {
	data := t.VDBOverviewData{}
	wg := errgroup.Group{}
	var err error
	// Groups
	if dashboardId.Validators == nil {
		// should have valid primary id
		wg.Go(func() error {
			var queryResult []struct {
				Id    uint32 `db:"id"`
				Name  string `db:"name"`
				Count uint64 `db:"count"`
			}
			query := `SELECT id, name, COUNT(validator_index)
			FROM
				users_val_dashboards_groups groups
			LEFT JOIN users_val_dashboards_validators validators
					ON groups.dashboard_id = validators.dashboard_id AND groups.id = validators.group_id
			WHERE
				groups.dashboard_id = $1
			GROUP BY
				groups.id, groups.name`
			if err := d.alloyReader.Select(&queryResult, query, dashboardId.Id); err != nil {
				return err
			}
			for _, res := range queryResult {
				data.Groups = append(data.Groups, t.VDBOverviewGroup{Id: uint64(res.Id), Name: res.Name, Count: res.Count})
			}
			return nil
		})
	}
	validators, err := d.getDashboardValidators(dashboardId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving validators from dashboard id: %v", err)
	}

	// Validator Status
	wg.Go(func() error {
		var query string
		var queryResult []struct {
			Name  string `db:"statename"`
			Count uint64 `db:"statecount"`
		}
		params := []interface{}{}
		if dashboardId.Validators == nil {
			query = `SELECT status AS statename, COUNT(*) AS statecount
			FROM validators v
			INNER JOIN users_val_dashboards_validators uvdv ON uvdv.validator_index = v.validatorindex
			WHERE uvdv.dashboard_id = $1
			GROUP BY status`
			params = append(params, dashboardId.Id)
		} else {
			query = `SELECT status AS statename, COUNT(*) AS statecount
			FROM validators
			WHERE validatorindex = ANY($1)
			GROUP BY status`
			params = append(params, validators)
		}
		err := d.alloyReader.Select(&queryResult, query, params...)
		if err != nil {
			return fmt.Errorf("error retrieving validators data: %v", err)
		}
		for _, state := range queryResult {
			switch state.Name {
			case "exiting_online":
				fallthrough
			case "slashing_online":
				fallthrough
			case "active_online":
				data.Validators.Online += state.Count
			case "exiting_offline":
				fallthrough
			case "slashing_offline":
				fallthrough
			case "active_offline":
				data.Validators.Offline += state.Count
			case "deposited":
				fallthrough
			case "pending":
				data.Validators.Pending += state.Count
			case "slashed":
				data.Validators.Slashed += state.Count
			case "exited":
				data.Validators.Exited += state.Count
			}
		}
		return nil
	})

	query := `SELECT
		SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
		SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
		SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
	FROM %[1]s v
	INNER JOIN users_val_dashboards_validators uvdv ON uvdv.validator_index = v.validator_index
	WHERE uvdv.dashboard_id = $1`

	if dashboardId.Validators != nil {
		query = `SELECT
			SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
			SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
			SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
		FROM %[1]s
		WHERE validator_index = ANY($1)`
	}

	retrieveRewardsAndEfficiency := func(table string, days int, rewards *t.ClElValue[decimal.Decimal], apr *t.ClElValue[float64], efficiency *float64) {
		// Rewards + APR
		wg.Go(func() error {
			(*rewards).Cl, (*apr).Cl, (*rewards).El, (*apr).El, err = d.internal_getElClAPR(validators, days)
			if err != nil {
				return err
			}
			return nil
		})

		// Efficiency
		wg.Go(func() error {
			var params interface{}
			if dashboardId.Validators == nil {
				params = dashboardId.Id
			} else {
				params = validators
			}
			var queryResult struct {
				AttestationEfficiency sql.NullFloat64 `db:"attestation_efficiency"`
				ProposerEfficiency    sql.NullFloat64 `db:"proposer_efficiency"`
				SyncEfficiency        sql.NullFloat64 `db:"sync_efficiency"`
			}

			err := d.alloyReader.Get(&queryResult, fmt.Sprintf(query, table), params)
			if err != nil {
				return err
			}
			*efficiency = d.calculateTotalEfficiency(queryResult.AttestationEfficiency, queryResult.ProposerEfficiency, queryResult.SyncEfficiency)
			return nil
		})
	}

	retrieveRewardsAndEfficiency("validator_dashboard_data_rolling_daily", 1, &data.Rewards.Last24h, &data.Apr.Last24h, &data.Efficiency.Last24h)
	retrieveRewardsAndEfficiency("validator_dashboard_data_rolling_weekly", 7, &data.Rewards.Last7d, &data.Apr.Last7d, &data.Efficiency.Last7d)
	retrieveRewardsAndEfficiency("validator_dashboard_data_rolling_monthly", 30, &data.Rewards.Last30d, &data.Apr.Last30d, &data.Efficiency.Last30d)
	retrieveRewardsAndEfficiency("validator_dashboard_data_rolling_total", -1, &data.Rewards.AllTime, &data.Apr.AllTime, &data.Efficiency.AllTime)

	err = wg.Wait()

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard overview data: %v", err)
	}

	return &data, nil
}

func (d *DataAccessService) CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (*t.VDBPostCreateGroupData, error) {
	result := &t.VDBPostCreateGroupData{}

	// Create a new group that has the smallest unique id possible
	err := d.alloyWriter.Get(result, `
		WITH NextAvailableId AS (
		    SELECT COALESCE(MIN(uvdg1.id) + 1, 0) AS next_id
		    FROM users_val_dashboards_groups uvdg1
		    LEFT JOIN users_val_dashboards_groups uvdg2 ON uvdg1.id + 1 = uvdg2.id AND uvdg1.dashboard_id = uvdg2.dashboard_id
		    WHERE uvdg1.dashboard_id = $1 AND uvdg2.id IS NULL
		)
		INSERT INTO users_val_dashboards_groups (id, dashboard_id, name)
			SELECT next_id, $1, $2
		FROM NextAvailableId
		RETURNING id, name
	`, dashboardId, name)

	return result, err
}

// updates the group name
func (d *DataAccessService) UpdateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64, name string) (*t.VDBPostCreateGroupData, error) {
	tx, err := d.alloyWriter.Beginx()
	if err != nil {
		return nil, fmt.Errorf("error starting db transactions to remove a validator dashboard group: %w", err)
	}
	defer utils.Rollback(tx)

	// Update the group name
	_, err = tx.Exec(`
		UPDATE users_val_dashboards_groups SET name = $1 WHERE dashboard_id = $2 AND id = $3
	`, name, dashboardId, groupId)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing tx to update a validator dashboard group: %w", err)
	}

	ret := &t.VDBPostCreateGroupData{
		Id:   groupId,
		Name: name,
	}
	return ret, nil
}

func (d *DataAccessService) RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error {
	tx, err := d.alloyWriter.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions to remove a validator dashboard group: %w", err)
	}
	defer utils.Rollback(tx)

	// Delete the group
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards_groups WHERE dashboard_id = $1 AND id = $2
	`, dashboardId, groupId)
	if err != nil {
		return err
	}

	// Delete all validators for the group
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards_validators WHERE dashboard_id = $1 AND group_id = $2
	`, dashboardId, groupId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing tx to remove a validator dashboard group: %w", err)
	}
	return nil
}

func (d *DataAccessService) GetValidatorDashboardValidators(dashboardId t.VDBId, groupId int64, cursor string, colSort t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
	// Initialize the cursor
	var currentCursor t.ValidatorsCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.ValidatorsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as ValidatorsCursor: %w", err)
		}
	}

	type ValidatorGroupInfo struct {
		GroupId   uint64
		GroupName string
	}
	validatorGroupMap := make(map[t.VDBValidator]ValidatorGroupInfo)
	var validators []t.VDBValidator
	if dashboardId.Validators == nil {
		// Get the validators and their groups in case a dashboard id is provided
		queryResult := []struct {
			ValidatorIndex t.VDBValidator `db:"validator_index"`
			GroupId        uint64         `db:"group_id"`
			GroupName      string         `db:"group_name"`
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
		validatorsParams := []interface{}{dashboardId.Id}

		if groupId != t.AllGroups {
			validatorsQuery += " AND group_id = $2"
			validatorsParams = append(validatorsParams, groupId)
		}
		err := d.alloyReader.Select(&queryResult, validatorsQuery, validatorsParams...)
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
		// In case a list of validators is provided, set the group to the default
		for _, validator := range dashboardId.Validators {
			validatorGroupMap[validator] = ValidatorGroupInfo{
				GroupId:   t.DefaultGroupId,
				GroupName: t.DefaultGroupName,
			}
			validators = append(validators, validator)
		}
	}
	var paging t.Paging

	if len(validators) == 0 {
		// Return if there are no validators
		return nil, &paging, nil
	}

	// Get the current validator state
	validatorMapping, releaseValMapLock, err := d.services.GetCurrentValidatorMapping()
	defer releaseValMapLock()
	if err != nil {
		return nil, nil, err
	}

	// Get the validator duties to check the last fulfilled attestation
	dutiesInfo, releaseValDutiesLock, err := d.services.GetCurrentDutiesInfo()
	defer releaseValDutiesLock()
	if err != nil {
		return nil, nil, err
	}

	// Set the threshold for "online" => "offline" to 2 epochs without attestation
	attestationThresholdSlot := uint64(0)
	twoEpochs := 2 * utils.Config.Chain.ClConfig.SlotsPerEpoch
	if dutiesInfo.LatestSlot >= twoEpochs {
		attestationThresholdSlot = dutiesInfo.LatestSlot - twoEpochs
	}

	// Fill the data
	data := []t.VDBManageValidatorsTableRow{}
	for _, validator := range validators {
		metadata := validatorMapping.ValidatorMetadata[validator]

		row := t.VDBManageValidatorsTableRow{
			Index:                validator,
			PublicKey:            t.PubKey(hexutil.Encode(metadata.PublicKey)),
			GroupId:              validatorGroupMap[validator].GroupId,
			Balance:              utils.GWeiToWei(big.NewInt(int64(metadata.Balance))),
			WithdrawalCredential: t.Hash(hexutil.Encode(metadata.WithdrawalCredentials)),
		}

		status := ""
		switch constypes.ValidatorStatus(metadata.Status) {
		case constypes.PendingInitialized:
			status = "deposited"
		case constypes.PendingQueued:
			status = "pending"
			if metadata.Queues.ActivationIndex.Valid {
				activationIndex := uint64(metadata.Queues.ActivationIndex.Int64)
				row.QueuePosition = &activationIndex
			}
		case constypes.ActiveOngoing, constypes.ActiveExiting, constypes.ActiveSlashed:
			var lastAttestionSlot uint32
			for slot, attested := range dutiesInfo.EpochAttestationDuties[validator] {
				if attested && slot > lastAttestionSlot {
					lastAttestionSlot = slot
				}
			}
			if lastAttestionSlot < uint32(attestationThresholdSlot) {
				status = "offline"
			} else {
				status = "online"
			}
		case constypes.ExitedUnslashed, constypes.ExitedSlashed, constypes.WithdrawalPossible, constypes.WithdrawalDone:
			if metadata.Slashed {
				status = "slashed"
			} else {
				status = "exited"
			}
		}
		row.Status = status

		if search == "" {
			data = append(data, row)
		} else {
			index, err := strconv.ParseUint(search, 10, 64)
			indexSearch := err == nil && index == row.Index

			pubKey := strings.ToLower(strings.TrimPrefix(search, "0x"))
			pubkeySearch := pubKey == strings.TrimPrefix(string(row.PublicKey), "0x")

			groupNameSearch := search == validatorGroupMap[validator].GroupName

			if indexSearch || pubkeySearch || groupNameSearch {
				data = append(data, row)
			}
		}
	}

	// no data found (searched for something that does not exist)
	if len(data) == 0 {
		return nil, &paging, nil
	}

	// Sort the result
	sort.Slice(data, func(i, j int) bool {
		switch colSort.Column {
		case enums.VDBManageValidatorsIndex:
			if data[i].Index != data[j].Index {
				return (data[i].Index < data[j].Index) != colSort.Desc
			}
		case enums.VDBManageValidatorsPublicKey:
			if data[i].PublicKey != data[j].PublicKey {
				return (data[i].PublicKey < data[j].PublicKey) != colSort.Desc
			}
		case enums.VDBManageValidatorsBalance:
			if data[i].Balance.Cmp(data[j].Balance) != 0 {
				return (data[i].Balance.Cmp(data[j].Balance) < 0) != colSort.Desc
			}
		case enums.VDBManageValidatorsStatus:
			if data[i].Status != data[j].Status {
				return (data[i].Status < data[j].Status) != colSort.Desc
			}
		case enums.VDBManageValidatorsWithdrawalCredential:
			if data[i].WithdrawalCredential != data[j].WithdrawalCredential {
				return (data[i].WithdrawalCredential < data[j].WithdrawalCredential) != colSort.Desc
			}
		}
		return false
	})

	// Find the index for the cursor and limit the data
	var cursorIndex uint64
	if currentCursor.IsValid() {
		for idx, row := range data {
			if row.Index == currentCursor.Index {
				cursorIndex = uint64(idx)
				break
			}
		}
	}

	var result []t.VDBManageValidatorsTableRow
	if currentCursor.IsReverse() {
		// opposite direction
		var limitCutoff uint64
		if cursorIndex > limit+1 {
			limitCutoff = cursorIndex - limit - 1
		}
		result = data[limitCutoff:cursorIndex]
	} else {
		if currentCursor.IsValid() {
			cursorIndex++
		}
		limitCutoff := utilMath.MinU64(cursorIndex+limit+1, uint64(len(data)))
		result = data[cursorIndex:limitCutoff]
	}

	// flag if above limit
	moreDataFlag := len(result) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// no paging required
		return result, &paging, nil
	}

	// remove the last entry from data as it is only required for the check
	if moreDataFlag {
		if currentCursor.IsReverse() {
			result = result[1:]
		} else {
			result = result[:len(result)-1]
		}
	}

	p, err := utils.GetPagingFromData(result, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return result, p, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupExists(dashboardId t.VDBIdPrimary, groupId uint64) (bool, error) {
	groupExists := false
	err := d.alloyReader.Get(&groupExists, `
		SELECT EXISTS(
			SELECT
				dashboard_id,
				id
			FROM users_val_dashboards_groups
			WHERE dashboard_id = $1 AND id = $2
		)
	`, dashboardId, groupId)
	return groupExists, err
}

func (d *DataAccessService) AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId int64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error) {
	if len(validators) == 0 {
		// No validators to add
		return nil, nil
	}

	pubkeys := []struct {
		ValidatorIndex t.VDBValidator `db:"validatorindex"`
		Pubkey         []byte         `db:"pubkey"`
	}{}

	addedValidators := []struct {
		ValidatorIndex t.VDBValidator `db:"validator_index"`
		GroupId        uint64         `db:"group_id"`
	}{}

	// Query to find the pubkey for each validator index
	pubkeysQuery := `
		SELECT
			validatorindex,
			pubkey
		FROM validators
		WHERE validatorindex = ANY($1)
	`

	// Query to add the validators to the dashboard and group
	addValidatorsQuery := `
		INSERT INTO users_val_dashboards_validators (dashboard_id, group_id, validator_index)
			VALUES 
	`

	for idx := range validators {
		addValidatorsQuery += fmt.Sprintf("($1, $2, $%d), ", idx+3)
	}
	addValidatorsQuery = addValidatorsQuery[:len(addValidatorsQuery)-2] // remove trailing comma

	// If a validator is already in the dashboard, update the group
	// If the validator is already in that group nothing changes but we will include it in the result anyway
	addValidatorsQuery += `
		ON CONFLICT (dashboard_id, validator_index) DO UPDATE SET 
			dashboard_id = EXCLUDED.dashboard_id,
			group_id = EXCLUDED.group_id,
			validator_index = EXCLUDED.validator_index
		RETURNING validator_index, group_id
	`

	// Find all the pubkeys
	err := d.alloyReader.Select(&pubkeys, pubkeysQuery, pq.Array(validators))
	if err != nil {
		return nil, err
	}

	// Add all the validators to the dashboard and group
	addValidatorsArgsIntf := []interface{}{dashboardId, groupId}
	for _, validatorIndex := range validators {
		addValidatorsArgsIntf = append(addValidatorsArgsIntf, validatorIndex)
	}
	err = d.alloyWriter.Select(&addedValidators, addValidatorsQuery, addValidatorsArgsIntf...)
	if err != nil {
		return nil, err
	}

	// Combine the pubkeys and group ids for the result
	pubkeysMap := make(map[t.VDBValidator]string, len(pubkeys))
	for _, pubKeyInfo := range pubkeys {
		pubkeysMap[pubKeyInfo.ValidatorIndex] = fmt.Sprintf("%#x", pubKeyInfo.Pubkey)
	}

	addedValidatorsMap := make(map[t.VDBValidator]uint64, len(addedValidators))
	for _, addedValidatorInfo := range addedValidators {
		addedValidatorsMap[addedValidatorInfo.ValidatorIndex] = addedValidatorInfo.GroupId
	}

	result := []t.VDBPostValidatorsData{}
	for _, validator := range validators {
		result = append(result, t.VDBPostValidatorsData{
			PublicKey: pubkeysMap[validator],
			GroupId:   addedValidatorsMap[validator],
		})
	}

	return result, nil
}

func (d *DataAccessService) RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error {
	if len(validators) == 0 {
		// Remove all validators for the dashboard
		_, err := d.alloyWriter.Exec(`
			DELETE FROM users_val_dashboards_validators 
			WHERE dashboard_id = $1
		`, dashboardId)
		return err
	}

	//Create the query to delete validators
	deleteValidatorsQuery := `
		DELETE FROM users_val_dashboards_validators
		WHERE dashboard_id = $1 AND validator_index = ANY($2)
	`

	// Delete the validators
	_, err := d.alloyWriter.Exec(deleteValidatorsQuery, dashboardId, pq.Array(validators))

	return err
}

func (d *DataAccessService) CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (*t.VDBPublicId, error) {
	dbReturn := struct {
		PublicId     string `db:"public_id"`
		Name         string `db:"name"`
		SharedGroups bool   `db:"shared_groups"`
	}{}

	// Create the public validator dashboard, multiple entries for the same dashboard are possible
	err := d.alloyWriter.Get(&dbReturn, `
		INSERT INTO users_val_dashboards_sharing (dashboard_id, name, shared_groups)
			VALUES ($1, $2, $3)
		RETURNING public_id, name, shared_groups
	`, dashboardId, name, showGroupNames)
	if err != nil {
		return nil, err
	}

	result := &t.VDBPublicId{}
	result.PublicId = dbReturn.PublicId
	result.Name = dbReturn.Name
	result.ShareSettings.GroupNames = dbReturn.SharedGroups

	return result, nil
}

func (d *DataAccessService) UpdateValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic, name string, showGroupNames bool) (*t.VDBPublicId, error) {
	dbReturn := struct {
		PublicId     string `db:"public_id"`
		Name         string `db:"name"`
		SharedGroups bool   `db:"shared_groups"`
	}{}

	// Update the name and settings of the public validator dashboard
	err := d.alloyWriter.Get(&dbReturn, `
		UPDATE users_val_dashboards_sharing SET
			name = $1,
			shared_groups = $2
		WHERE public_id = $3
		RETURNING public_id, name, shared_groups
	`, name, showGroupNames, publicDashboardId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: public dashboard id %v not found", ErrNotFound, publicDashboardId)
		}
		return nil, err
	}

	result := &t.VDBPublicId{}
	result.PublicId = dbReturn.PublicId
	result.Name = dbReturn.Name
	result.ShareSettings.GroupNames = dbReturn.SharedGroups

	return result, nil
}

func (d *DataAccessService) RemoveValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic) error {
	// Delete the public validator dashboard
	result, err := d.alloyWriter.Exec(`
		DELETE FROM users_val_dashboards_sharing WHERE public_id = $1
	`, publicDashboardId)
	if err != nil {
		return err
	}

	// Check if the public validator dashboard was deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("error public dashboard id %v does not exist, cannot remove it", publicDashboardId)
	}

	return err
}
