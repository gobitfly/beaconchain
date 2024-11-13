package dataaccess

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/doug-martin/goqu/v9"
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

func (d *DataAccessService) GetValidatorDashboardUser(ctx context.Context, dashboardId t.VDBIdPrimary) (*t.DashboardUser, error) {
	result := &t.DashboardUser{}

	err := d.alloyReader.GetContext(ctx, result, `
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

func (d *DataAccessService) GetValidatorDashboardIdByPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) (*t.VDBIdPrimary, error) {
	var result t.VDBIdPrimary

	err := d.alloyReader.GetContext(ctx, &result, `
		SELECT
			uvd.id
		FROM users_val_dashboards_sharing uvds
		LEFT JOIN users_val_dashboards uvd ON uvd.id = uvds.dashboard_id
		WHERE uvds.public_id = $1
	`, publicDashboardId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: public id %v not found", ErrNotFound, publicDashboardId)
	}
	return &result, err
}

func (d *DataAccessService) GetValidatorDashboardInfo(ctx context.Context, dashboardId t.VDBIdPrimary) (*t.ValidatorDashboard, error) {
	result := &t.ValidatorDashboard{}

	wg := errgroup.Group{}
	mutex := &sync.RWMutex{}

	wg.Go(func() error {
		dbReturn := []struct {
			Name         string         `db:"name"`
			Network      uint64         `db:"network"`
			IsArchived   sql.NullString `db:"is_archived"`
			PublicId     sql.NullString `db:"public_id"`
			PublicName   sql.NullString `db:"public_name"`
			SharedGroups sql.NullBool   `db:"shared_groups"`
		}{}

		err := d.alloyReader.SelectContext(ctx, &dbReturn, `
		SELECT
			uvd.name,
			uvd.network,
			uvd.is_archived,
			uvds.public_id,
			uvds.name AS public_name,
			uvds.shared_groups
		FROM users_val_dashboards uvd
		LEFT JOIN users_val_dashboards_sharing uvds ON uvd.id = uvds.dashboard_id
		WHERE uvd.id = $1
	`, dashboardId)
		if err != nil {
			return err
		}

		if len(dbReturn) == 0 {
			return fmt.Errorf("error dashboard with id %v not found", dashboardId)
		}

		mutex.Lock()
		result.Id = uint64(dashboardId)
		result.Name = dbReturn[0].Name
		result.Network = dbReturn[0].Network
		result.IsArchived = dbReturn[0].IsArchived.Valid
		result.ArchivedReason = dbReturn[0].IsArchived.String

		for _, row := range dbReturn {
			if row.PublicId.Valid {
				publicId := t.VDBPublicId{}
				publicId.PublicId = row.PublicId.String
				publicId.Name = row.PublicName.String
				publicId.ShareSettings.ShareGroups = row.SharedGroups.Bool

				result.PublicIds = append(result.PublicIds, publicId)
			}
		}
		mutex.Unlock()

		return nil
	})

	wg.Go(func() error {
		dbReturn := struct {
			GroupCount     uint64 `db:"group_count"`
			ValidatorCount uint64 `db:"validator_count"`
		}{}

		err := d.alloyReader.GetContext(ctx, &dbReturn, `
			WITH dashboards_groups AS
				(SELECT COUNT(uvdg.id) AS group_count FROM users_val_dashboards_groups uvdg WHERE uvdg.dashboard_id = $1),
			dashboards_validators AS
				(SELECT COUNT(uvdv.validator_index) AS validator_count FROM users_val_dashboards_validators uvdv WHERE uvdv.dashboard_id = $1)
			SELECT
			    dashboards_groups.group_count,
			    dashboards_validators.validator_count
			FROM 
			    dashboards_groups,
			    dashboards_validators
		`, dashboardId)
		if err != nil {
			return err
		}

		mutex.Lock()
		result.GroupCount = dbReturn.GroupCount
		result.ValidatorCount = dbReturn.ValidatorCount
		mutex.Unlock()

		return nil
	})

	err := wg.Wait()
	if err != nil {
		return nil, fmt.Errorf("error retrieving user dashboards data: %w", err)
	}

	return result, nil
}

func (d *DataAccessService) GetValidatorDashboardName(ctx context.Context, dashboardId t.VDBIdPrimary) (string, error) {
	var name string
	err := d.alloyReader.GetContext(ctx, &name, `
		SELECT name
		FROM users_val_dashboards
		WHERE id = $1
	`, dashboardId)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: dashboard with id %v not found", ErrNotFound, dashboardId)
	}
	return name, err
}

// param validators: slice of validator public keys or indices
func (d *DataAccessService) GetValidatorsFromSlices(ctx context.Context, indices []t.VDBValidator, publicKeys []string) ([]t.VDBValidator, error) {
	if len(indices) == 0 && len(publicKeys) == 0 {
		return []t.VDBValidator{}, nil
	}

	mapping, err := d.services.GetCurrentValidatorMapping()
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

func (d *DataAccessService) CreateValidatorDashboard(ctx context.Context, userId uint64, name string, network uint64) (*t.VDBPostReturnData, error) {
	result := &t.VDBPostReturnData{}

	tx, err := d.alloyWriter.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting db transactions to create a validator dashboard: %w", err)
	}
	defer utils.Rollback(tx)

	// Create validator dashboard for user
	err = tx.GetContext(ctx, result, `
		INSERT INTO users_val_dashboards (user_id, network, name)
			VALUES ($1, $2, $3)
		RETURNING id, user_id, name, network, (EXTRACT(epoch FROM created_at))::BIGINT as created_at
	`, userId, network, name)
	if err != nil {
		return nil, err
	}

	// Create a default group for the new dashboard
	_, err = tx.ExecContext(ctx, `
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

func (d *DataAccessService) RemoveValidatorDashboard(ctx context.Context, dashboardId t.VDBIdPrimary) error {
	_, err := d.alloyWriter.ExecContext(ctx, `
		DELETE FROM users_val_dashboards WHERE id = $1
	`, dashboardId)
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("%s:%d:", ValidatorDashboardEventPrefix, dashboardId)

	// Remove all events related to the dashboard
	_, err = d.userWriter.ExecContext(ctx, `
		DELETE FROM users_subscriptions WHERE event_filter LIKE ($1 || '%')
	`, prefix)
	return err
}

func (d *DataAccessService) UpdateValidatorDashboardArchiving(ctx context.Context, dashboardId t.VDBIdPrimary, archivedReason *enums.VDBArchivedReason) (*t.VDBPostArchivingReturnData, error) {
	result := &t.VDBPostArchivingReturnData{}

	var archivedReasonText *string
	if archivedReason != nil {
		reason := archivedReason.ToString()
		archivedReasonText = &reason
	}

	err := d.alloyWriter.GetContext(ctx, result, `
		UPDATE users_val_dashboards SET is_archived = $1 WHERE id = $2
		RETURNING id, is_archived IS NOT NULL AS is_archived
	`, archivedReasonText, dashboardId)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DataAccessService) UpdateValidatorDashboardName(ctx context.Context, dashboardId t.VDBIdPrimary, name string) (*t.VDBPostReturnData, error) {
	result := &t.VDBPostReturnData{}

	err := d.alloyWriter.GetContext(ctx, result, `
		UPDATE users_val_dashboards SET name = $1 WHERE id = $2
		RETURNING id, user_id, name, network, (EXTRACT(epoch FROM created_at))::BIGINT as created_at
	`, name, dashboardId)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DataAccessService) GetValidatorDashboardOverview(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes) (*t.VDBOverviewData, error) {
	data := t.VDBOverviewData{}
	eg := errgroup.Group{}
	var err error

	// Network
	if dashboardId.Validators == nil {
		eg.Go(func() error {
			query := `SELECT network
			FROM
				users_val_dashboards
			WHERE
				id = $1`
			return d.alloyReader.GetContext(ctx, &data.Network, query, dashboardId.Id)
		})
	} else { // load the chain id from the config in case of public dashboards
		data.Network = utils.Config.Chain.ClConfig.DepositChainID
	}

	// Groups
	if dashboardId.Validators == nil && !dashboardId.AggregateGroups {
		// should have valid primary id
		eg.Go(func() error {
			var queryResult []struct {
				Id    uint32 `db:"id"`
				Name  string `db:"name"`
				Count uint64 `db:"count"`
			}
			query := `SELECT groups.id, groups.name, COUNT(validators.validator_index)
			FROM
				users_val_dashboards_groups groups
			LEFT JOIN users_val_dashboards_validators validators
					ON groups.dashboard_id = validators.dashboard_id AND groups.id = validators.group_id
			WHERE
				groups.dashboard_id = $1
			GROUP BY
				groups.id, groups.name`
			if err := d.alloyReader.SelectContext(ctx, &queryResult, query, dashboardId.Id); err != nil {
				return err
			}

			for _, res := range queryResult {
				data.Groups = append(data.Groups, t.VDBOverviewGroup{Id: uint64(res.Id), Name: res.Name, Count: res.Count})
			}

			return nil
		})
	}

	// Validator status and balance
	eg.Go(func() error {
		validatorMapping, err := d.services.GetCurrentValidatorMapping()
		if err != nil {
			return err
		}

		validators, err := d.getDashboardValidators(ctx, dashboardId, nil)
		if err != nil {
			return fmt.Errorf("error retrieving validators from dashboard id: %w", err)
		}

		if dashboardId.Validators != nil || dashboardId.AggregateGroups {
			data.Groups = append(data.Groups, t.VDBOverviewGroup{Id: t.DefaultGroupId, Name: t.DefaultGroupName, Count: uint64(len(validators))})
		}

		// Status
		for _, validator := range validators {
			metadata := validatorMapping.ValidatorMetadata[validator]

			switch constypes.ValidatorDbStatus(metadata.Status) {
			case constypes.DbExitingOnline, constypes.DbSlashingOnline, constypes.DbActiveOnline:
				data.Validators.Online++
			case constypes.DbExitingOffline, constypes.DbSlashingOffline, constypes.DbActiveOffline:
				data.Validators.Offline++
			case constypes.DbDeposited, constypes.DbPending:
				data.Validators.Pending++
			case constypes.DbSlashed:
				data.Validators.Slashed++
			case constypes.DbExited:
				data.Validators.Exited++
			}
		}

		// Find rocketpool validators
		type RpOperatorInfo struct {
			ValidatorIndex     uint64          `db:"validatorindex"`
			NodeFee            float64         `db:"node_fee"`
			NodeDepositBalance decimal.Decimal `db:"node_deposit_balance"`
			UserDepositBalance decimal.Decimal `db:"user_deposit_balance"`
		}
		var queryResult []RpOperatorInfo

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
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.alloyReader.SelectContext(ctx, &queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving rocketpool validators data: %w", err)
		}

		rpValidators := make(map[uint64]RpOperatorInfo)
		for _, res := range queryResult {
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

					data.Balances.Total = data.Balances.Total.Add(operatorBalance)
				} else {
					data.Balances.Total = data.Balances.Total.Add(validatorBalance)
				}
				data.Balances.StakedEth = data.Balances.StakedEth.Add(rpValidator.NodeDepositBalance)
			} else {
				data.Balances.Total = data.Balances.Total.Add(validatorBalance)

				nonRpDashboardId.Validators = append(nonRpDashboardId.Validators, validator)
			}
			data.Balances.Effective = data.Balances.Effective.Add(effectiveBalance)
		}

		// Get the total cl deposits for non-rocketpool validators
		if len(nonRpDashboardId.Validators) > 0 {
			totalNonRpDeposits, err := d.GetValidatorDashboardTotalClDeposits(ctx, nonRpDashboardId)
			if err != nil {
				return fmt.Errorf("error retrieving total cl deposits for non-rocketpool validators: %w", err)
			}
			data.Balances.StakedEth = data.Balances.StakedEth.Add(totalNonRpDeposits.TotalAmount)
		}

		return nil
	})

	retrieveRewardsAndEfficiency := func(table string, hours int, rewards *t.ClElValue[decimal.Decimal], apr *t.ClElValue[float64], efficiency *float64) {
		// Rewards + APR
		eg.Go(func() error {
			(*rewards).El, (*apr).El, (*rewards).Cl, (*apr).Cl, err = d.internal_getElClAPR(ctx, dashboardId, -1, hours)
			if err != nil {
				return err
			}
			return nil
		})

		// Efficiency
		eg.Go(func() error {
			ds := goqu.Dialect("postgres").
				From(goqu.L(fmt.Sprintf(`%s AS r FINAL`, table))).
				With("validators", goqu.L("(SELECT dashboard_id, validator_index FROM users_val_dashboards_validators WHERE dashboard_id = ?)", dashboardId.Id)).
				Select(
					goqu.L("COALESCE(SUM(r.attestations_reward)::decimal, 0) AS attestations_reward"),
					goqu.L("COALESCE(SUM(r.attestations_ideal_reward)::decimal, 0) AS attestations_ideal_reward"),
					goqu.L("COALESCE(SUM(r.blocks_proposed), 0) AS blocks_proposed"),
					goqu.L("COALESCE(SUM(r.blocks_scheduled), 0) AS blocks_scheduled"),
					goqu.L("COALESCE(SUM(r.sync_executed), 0) AS sync_executed"),
					goqu.L("COALESCE(SUM(r.sync_scheduled), 0) AS sync_scheduled"))

			if len(dashboardId.Validators) == 0 {
				ds = ds.
					InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
					Where(goqu.L("r.validator_index IN (SELECT validator_index FROM validators)"))
			} else {
				ds = ds.
					Where(goqu.L("r.validator_index IN ?", dashboardId.Validators))
			}

			var queryResult struct {
				AttestationReward      decimal.Decimal `db:"attestations_reward"`
				AttestationIdealReward decimal.Decimal `db:"attestations_ideal_reward"`
				BlocksProposed         uint64          `db:"blocks_proposed"`
				BlocksScheduled        uint64          `db:"blocks_scheduled"`
				SyncExecuted           uint64          `db:"sync_executed"`
				SyncScheduled          uint64          `db:"sync_scheduled"`
			}

			query, args, err := ds.Prepared(true).ToSQL()
			if err != nil {
				return fmt.Errorf("error preparing query: %w", err)
			}

			err = d.clickhouseReader.GetContext(ctx, &queryResult, query, args...)
			if err != nil {
				return err
			}

			// Calculate efficiency
			var attestationEfficiency, proposerEfficiency, syncEfficiency sql.NullFloat64
			if !queryResult.AttestationIdealReward.IsZero() {
				attestationEfficiency.Float64 = queryResult.AttestationReward.Div(queryResult.AttestationIdealReward).InexactFloat64()
				attestationEfficiency.Valid = true
			}
			if queryResult.BlocksScheduled > 0 {
				proposerEfficiency.Float64 = float64(queryResult.BlocksProposed) / float64(queryResult.BlocksScheduled)
				proposerEfficiency.Valid = true
			}
			if queryResult.SyncScheduled > 0 {
				syncEfficiency.Float64 = float64(queryResult.SyncExecuted) / float64(queryResult.SyncScheduled)
				syncEfficiency.Valid = true
			}
			*efficiency = utils.CalculateTotalEfficiency(attestationEfficiency, proposerEfficiency, syncEfficiency)

			return nil
		})
	}

	retrieveRewardsAndEfficiency("validator_dashboard_data_rolling_24h", 24, &data.Rewards.Last24h, &data.Apr.Last24h, &data.Efficiency.Last24h)
	retrieveRewardsAndEfficiency("validator_dashboard_data_rolling_7d", 7*24, &data.Rewards.Last7d, &data.Apr.Last7d, &data.Efficiency.Last7d)
	retrieveRewardsAndEfficiency("validator_dashboard_data_rolling_30d", 30*24, &data.Rewards.Last30d, &data.Apr.Last30d, &data.Efficiency.Last30d)
	retrieveRewardsAndEfficiency("validator_dashboard_data_rolling_total", -1, &data.Rewards.AllTime, &data.Apr.AllTime, &data.Efficiency.AllTime)

	err = eg.Wait()

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard overview data: %w", err)
	}

	return &data, nil
}

func (d *DataAccessService) CreateValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, name string) (*t.VDBPostCreateGroupData, error) {
	result := &t.VDBPostCreateGroupData{}

	// Create a new group that has the smallest unique id possible
	err := d.alloyWriter.GetContext(ctx, result, `
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
func (d *DataAccessService) UpdateValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, name string) (*t.VDBPostCreateGroupData, error) {
	tx, err := d.alloyWriter.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting db transactions to remove a validator dashboard group: %w", err)
	}
	defer utils.Rollback(tx)

	// Update the group name
	_, err = tx.ExecContext(ctx, `
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

func (d *DataAccessService) RemoveValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64) error {
	// Delete the group
	_, err := d.alloyWriter.ExecContext(ctx, `
		DELETE FROM users_val_dashboards_groups WHERE dashboard_id = $1 AND id = $2
	`, dashboardId, groupId)
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("%s:%d:%d", ValidatorDashboardEventPrefix, dashboardId, groupId)

	// Remove all events related to the group
	_, err = d.userWriter.ExecContext(ctx, `
		DELETE FROM users_subscriptions WHERE event_filter = $1
	`, prefix)
	return err
}

func (d *DataAccessService) RemoveValidatorDashboardGroupValidators(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64) error {
	//Create the query to delete validators
	deleteValidatorsQuery := `
		DELETE FROM users_val_dashboards_validators
		WHERE dashboard_id = $1 AND group_id = $2
	`

	// Delete the validators
	_, err := d.alloyWriter.ExecContext(ctx, deleteValidatorsQuery, dashboardId, groupId)

	return err
}

func (d *DataAccessService) GetValidatorDashboardGroupCount(ctx context.Context, dashboardId t.VDBIdPrimary) (uint64, error) {
	var count uint64
	err := d.alloyReader.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM users_val_dashboards_groups WHERE dashboard_id = $1
	`, dashboardId)
	return count, err
}

func (d *DataAccessService) GetValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, cursor string, colSort t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
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
		err := d.alloyReader.SelectContext(ctx, &queryResult, validatorsQuery, validatorsParams...)
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
	validatorMapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, nil, err
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
			Status:               metadata.Status,
			WithdrawalCredential: t.Hash(hexutil.Encode(metadata.WithdrawalCredentials)),
		}

		if constypes.ValidatorDbStatus(metadata.Status) == constypes.DbPending && metadata.Queues.ActivationIndex.Valid {
			activationIndex := uint64(metadata.Queues.ActivationIndex.Int64)
			row.QueuePosition = &activationIndex
		}

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

func (d *DataAccessService) GetValidatorDashboardGroupExists(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64) (bool, error) {
	groupExists := false
	err := d.alloyReader.GetContext(ctx, &groupExists, `
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

func (d *DataAccessService) AddValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error) {
	result := []t.VDBPostValidatorsData{}

	if len(validators) == 0 {
		// No validators to add
		return nil, nil
	}

	tx, err := d.userWriter.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting db transactions to insert validators for a dashboard: %w", err)
	}
	defer utils.Rollback(tx)

	numArgs := 3
	batchSize := 65535 / numArgs // max 65535 params per batch, since postgres uses int16 for binding input params
	batchIdx, allIdx := 0, 0
	var validatorsToInsert []goqu.Record
	for _, validatorIdx := range validators {
		validatorsToInsert = append(validatorsToInsert,
			goqu.Record{"dashboard_id": dashboardId, "group_id": groupId, "validator_index": validatorIdx})

		batchIdx++
		allIdx++

		if batchIdx >= batchSize || allIdx >= len(validators) {
			insertDs := goqu.Dialect("postgres").
				Insert("users_val_dashboards_validators").
				Cols("dashboard_id", "group_id", "validator_index").
				Rows(validatorsToInsert).
				OnConflict(goqu.DoUpdate(
					"dashboard_id, validator_index",
					goqu.Record{
						"dashboard_id":    goqu.L("EXCLUDED.dashboard_id"),
						"group_id":        goqu.L("EXCLUDED.group_id"),
						"validator_index": goqu.L("EXCLUDED.validator_index"),
					},
				))

			query, args, err := insertDs.Prepared(true).ToSQL()
			if err != nil {
				return nil, fmt.Errorf("error preparing query: %w", err)
			}

			_, err = tx.ExecContext(ctx, query, args...)
			if err != nil {
				return nil, err
			}

			batchIdx = 0
			validatorsToInsert = validatorsToInsert[:0]
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing tx to insert validators for a dashboard: %w", err)
	}

	for _, validator := range validators {
		result = append(result, t.VDBPostValidatorsData{
			Index:   validator,
			GroupId: groupId,
		})
	}

	return result, nil
}

// Updates the group for validators already in the dashboard linked to the deposit address.
// Adds up to limit new validators associated with the deposit address, if not already in the dashboard.
func (d *DataAccessService) AddValidatorDashboardValidatorsByDepositAddress(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, address string, limit uint64) ([]t.VDBPostValidatorsData, error) {
	result := []t.VDBPostValidatorsData{}

	addressParsed, err := hex.DecodeString(strings.TrimPrefix(address, "0x"))
	if err != nil {
		return nil, err
	}

	uniqueValidatorIndexesQuery := `
		(SELECT 
   		    DISTINCT uvdv.validator_index
   		FROM validators v
   		JOIN eth1_deposits d ON v.pubkey = d.publickey
   		JOIN users_val_dashboards_validators uvdv ON v.validatorindex = uvdv.validator_index
   		WHERE uvdv.dashboard_id = $1 AND d.from_address = $2)

   		UNION

   		(SELECT 
   		    DISTINCT v.validatorindex AS validator_index
   		FROM validators v
   		JOIN eth1_deposits d ON v.pubkey = d.publickey
   		LEFT JOIN users_val_dashboards_validators uvdv
   		    ON v.validatorindex = uvdv.validator_index AND uvdv.dashboard_id = $1
   		WHERE d.from_address = $2 AND uvdv.validator_index IS NULL
   		ORDER BY validator_index
   		LIMIT $3)`

	addValidatorsQuery := d.getAddValidatorsQuery(uniqueValidatorIndexesQuery)

	var validators []uint64
	err = d.alloyWriter.SelectContext(ctx, &validators, addValidatorsQuery, dashboardId, addressParsed, limit, groupId)
	if err != nil {
		return nil, err
	}

	for _, validator := range validators {
		result = append(result, t.VDBPostValidatorsData{
			Index:   validator,
			GroupId: groupId,
		})
	}

	return result, nil
}

// Updates the group for validators already in the dashboard linked to the withdrawal address.
// Adds up to limit new validators associated with the withdrawal address, if not already in the dashboard.
func (d *DataAccessService) AddValidatorDashboardValidatorsByWithdrawalAddress(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, address string, limit uint64) ([]t.VDBPostValidatorsData, error) {
	result := []t.VDBPostValidatorsData{}

	addressParsed, err := hex.DecodeString(strings.TrimPrefix(address, "0x"))
	if err != nil {
		return nil, err
	}

	uniqueValidatorIndexesQuery := `
		(SELECT 
			DISTINCT uvdv.validator_index
		FROM validators v
		JOIN users_val_dashboards_validators uvdv ON v.validatorindex = uvdv.validator_index
		WHERE uvdv.dashboard_id = $1 AND v.withdrawalcredentials = $2)

		UNION

		(SELECT 
			DISTINCT v.validatorindex AS validator_index
		FROM validators v
		LEFT JOIN users_val_dashboards_validators uvdv 
			ON v.validatorindex = uvdv.validator_index AND uvdv.dashboard_id = $1
		WHERE v.withdrawalcredentials = $2 AND uvdv.validator_index IS NULL
		ORDER BY v.validatorindex
		LIMIT $3)`

	addValidatorsQuery := d.getAddValidatorsQuery(uniqueValidatorIndexesQuery)

	var validators []uint64
	err = d.alloyWriter.SelectContext(ctx, &validators, addValidatorsQuery, dashboardId, addressParsed, limit, groupId)
	if err != nil {
		return nil, err
	}

	for _, validator := range validators {
		result = append(result, t.VDBPostValidatorsData{
			Index:   validator,
			GroupId: groupId,
		})
	}

	return result, nil
}

// Update the group for validators already in the dashboard linked to the graffiti (via produced block).
// Add up to limit new validators associated with the graffiti, if not already in the dashboard.
func (d *DataAccessService) AddValidatorDashboardValidatorsByGraffiti(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, graffiti string, limit uint64) ([]t.VDBPostValidatorsData, error) {
	result := []t.VDBPostValidatorsData{}

	uniqueValidatorIndexesQuery := `
		(SELECT 
			DISTINCT uvdv.validator_index
		FROM blocks b
		JOIN users_val_dashboards_validators uvdv ON b.proposer = uvdv.validator_index
		WHERE uvdv.dashboard_id = $1 AND b.graffiti_text = $2)

		UNION
		
		(SELECT DISTINCT b.proposer AS validator_index
		FROM blocks b
		LEFT JOIN users_val_dashboards_validators uvdv 
			ON b.proposer = uvdv.validator_index AND uvdv.dashboard_id = $1
		WHERE b.graffiti_text = $2 AND uvdv.validator_index IS NULL
		ORDER BY b.proposer
		LIMIT $3)`

	addValidatorsQuery := d.getAddValidatorsQuery(uniqueValidatorIndexesQuery)

	var validators []uint64
	err := d.alloyWriter.SelectContext(ctx, &validators, addValidatorsQuery, dashboardId, graffiti, limit, groupId)
	if err != nil {
		return nil, err
	}

	for _, validator := range validators {
		result = append(result, t.VDBPostValidatorsData{
			Index:   validator,
			GroupId: groupId,
		})
	}

	return result, nil
}

func (d *DataAccessService) getAddValidatorsQuery(uniqueValidatorIndexesQuery string) string {
	return fmt.Sprintf(`
		WITH unique_validator_indexes AS (
			%s
		)
		INSERT INTO users_val_dashboards_validators (dashboard_id, group_id, validator_index)
		SELECT $1 AS dashboard_id, $4 AS group_id, validator_index
		FROM unique_validator_indexes
		ON CONFLICT (dashboard_id, validator_index) DO UPDATE 
		SET
		    dashboard_id = EXCLUDED.dashboard_id,
		    group_id = EXCLUDED.group_id,
		    validator_index = EXCLUDED.validator_index
		RETURNING validator_index`, uniqueValidatorIndexesQuery)
}

func (d *DataAccessService) RemoveValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error {
	if len(validators) == 0 {
		// Remove all validators for the dashboard
		// This is usually forbidden by API validation
		_, err := d.alloyWriter.ExecContext(ctx, `
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
	_, err := d.alloyWriter.ExecContext(ctx, deleteValidatorsQuery, dashboardId, pq.Array(validators))

	return err
}

func (d *DataAccessService) GetValidatorDashboardValidatorsCount(ctx context.Context, dashboardId t.VDBIdPrimary) (uint64, error) {
	var count uint64
	err := d.alloyReader.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM users_val_dashboards_validators
		WHERE dashboard_id = $1
	`, dashboardId)
	return count, err
}

func (d *DataAccessService) CreateValidatorDashboardPublicId(ctx context.Context, dashboardId t.VDBIdPrimary, name string, shareGroups bool) (*t.VDBPublicId, error) {
	dbReturn := struct {
		PublicId     string `db:"public_id"`
		Name         string `db:"name"`
		SharedGroups bool   `db:"shared_groups"`
	}{}

	// Create the public validator dashboard, multiple entries for the same dashboard are possible
	err := d.alloyWriter.GetContext(ctx, &dbReturn, `
		INSERT INTO users_val_dashboards_sharing (dashboard_id, name, shared_groups)
			VALUES ($1, $2, $3)
		RETURNING public_id, name, shared_groups
	`, dashboardId, name, shareGroups)
	if err != nil {
		return nil, err
	}

	result := &t.VDBPublicId{}
	result.PublicId = dbReturn.PublicId
	result.Name = dbReturn.Name
	result.ShareSettings.ShareGroups = dbReturn.SharedGroups

	return result, nil
}

func (d *DataAccessService) GetValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) (*t.VDBPublicId, error) {
	dbReturn := struct {
		PublicId     string `db:"public_id"`
		DashboardId  int    `db:"dashboard_id"`
		Name         string `db:"name"`
		SharedGroups bool   `db:"shared_groups"`
	}{}

	// Get the public validator dashboard
	err := d.alloyReader.GetContext(ctx, &dbReturn, `
		SELECT public_id, dashboard_id, name, shared_groups
		FROM users_val_dashboards_sharing
		WHERE public_id = $1
	`, publicDashboardId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: public dashboard id %v not found", ErrNotFound, publicDashboardId)
		}
		return nil, err
	}

	result := &t.VDBPublicId{}
	result.DashboardId = dbReturn.DashboardId
	result.PublicId = dbReturn.PublicId
	result.Name = dbReturn.Name
	result.ShareSettings.ShareGroups = dbReturn.SharedGroups

	return result, nil
}

func (d *DataAccessService) UpdateValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic, name string, shareGroups bool) (*t.VDBPublicId, error) {
	dbReturn := struct {
		PublicId     string `db:"public_id"`
		Name         string `db:"name"`
		SharedGroups bool   `db:"shared_groups"`
	}{}

	// Update the name and settings of the public validator dashboard
	err := d.alloyWriter.GetContext(ctx, &dbReturn, `
		UPDATE users_val_dashboards_sharing SET
			name = $1,
			shared_groups = $2
		WHERE public_id = $3
		RETURNING public_id, name, shared_groups
	`, name, shareGroups, publicDashboardId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: public dashboard id %v not found", ErrNotFound, publicDashboardId)
		}
		return nil, err
	}

	result := &t.VDBPublicId{}
	result.PublicId = dbReturn.PublicId
	result.Name = dbReturn.Name
	result.ShareSettings.ShareGroups = dbReturn.SharedGroups

	return result, nil
}

func (d *DataAccessService) RemoveValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) error {
	// Delete the public validator dashboard
	result, err := d.alloyWriter.ExecContext(ctx, `
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

func (d *DataAccessService) GetValidatorDashboardPublicIdCount(ctx context.Context, dashboardId t.VDBIdPrimary) (uint64, error) {
	var count uint64
	err := d.alloyReader.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM users_val_dashboards_sharing
		WHERE dashboard_id = $1
	`, dashboardId)
	return count, err
}
