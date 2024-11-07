package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/userservice"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

type AppRepository interface {
	GetUserIdByRefreshToken(ctx context.Context, claimUserID, claimAppID, claimDeviceID uint64, hashedRefreshToken string) (uint64, error)
	MigrateMobileSession(ctx context.Context, oldHashedRefreshToken, newHashedRefreshToken, deviceID, deviceName string) error
	AddUserDevice(ctx context.Context, userID uint64, hashedRefreshToken string, deviceID, deviceName string, appID uint64) error
	GetAppDataFromRedirectUri(ctx context.Context, callback string) (*t.OAuthAppData, error)
	AddMobileNotificationToken(ctx context.Context, userID uint64, deviceID, notifyToken string) error
	GetAppSubscriptionCount(ctx context.Context, userID uint64) (uint64, error)
	AddMobilePurchase(ctx context.Context, tx *sql.Tx, userID uint64, paymentDetails t.MobileSubscription, verifyResponse *userservice.VerifyResponse, extSubscriptionId string) error
	GetLatestBundleForNativeVersion(ctx context.Context, nativeVersion uint64) (*t.MobileAppBundleStats, error)
	IncrementBundleDeliveryCount(ctx context.Context, bundleVerison uint64) error
	GetValidatorDashboardMobileValidators(ctx context.Context, dashboardId t.VDBId, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBMobileValidatorsColumn], search string, limit uint64) ([]t.MobileValidatorDashboardValidatorsTableRow, *t.Paging, error)
}

// GetUserIdByRefreshToken basically used to confirm the claimed user id with the refresh token. Returns the userId if successful
func (d *DataAccessService) GetUserIdByRefreshToken(ctx context.Context, claimUserID, claimAppID, claimDeviceID uint64, hashedRefreshToken string) (uint64, error) {
	if hashedRefreshToken == "" { // sanity
		return 0, errors.New("empty refresh token")
	}
	var userID uint64
	err := d.userWriter.GetContext(ctx, &userID,
		`SELECT user_id FROM users_devices WHERE user_id = $1 AND 
			refresh_token = $2 AND app_id = $3 AND id = $4 AND active = true`, claimUserID, hashedRefreshToken, claimAppID, claimDeviceID)
	if errors.Is(err, sql.ErrNoRows) {
		return userID, fmt.Errorf("%w: user not found via refresh token", ErrNotFound)
	}
	return userID, err
}

func (d *DataAccessService) MigrateMobileSession(ctx context.Context, oldHashedRefreshToken, newHashedRefreshToken, deviceID, deviceName string) error {
	result, err := d.userWriter.ExecContext(ctx, "UPDATE users_devices SET refresh_token = $2, device_identifier = $3, device_name = $4 WHERE refresh_token = $1", oldHashedRefreshToken, newHashedRefreshToken, deviceID, deviceName)
	if err != nil {
		return errors.Wrap(err, "Error updating refresh token")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "Error getting rows affected")
	}

	if rowsAffected != 1 {
		return errors.New(fmt.Sprintf("illegal number of rows affected, expected 1 got %d", rowsAffected))
	}

	return err
}

func (d *DataAccessService) GetAppDataFromRedirectUri(ctx context.Context, callback string) (*t.OAuthAppData, error) {
	data := t.OAuthAppData{}
	err := d.userWriter.GetContext(ctx, &data, "SELECT id, app_name, redirect_uri, active, owner_id FROM oauth_apps WHERE active = true AND redirect_uri = $1", callback)
	return &data, err
}

func (d *DataAccessService) AddUserDevice(ctx context.Context, userID uint64, hashedRefreshToken string, deviceID, deviceName string, appID uint64) error {
	_, err := d.userWriter.ExecContext(ctx, "INSERT INTO users_devices (user_id, refresh_token, device_identifier, device_name, app_id, created_ts) VALUES($1, $2, $3, $4, $5, 'NOW()') ON CONFLICT DO NOTHING",
		userID, hashedRefreshToken, deviceID, deviceName, appID,
	)
	return err
}

func (d *DataAccessService) AddMobileNotificationToken(ctx context.Context, userID uint64, deviceID, notifyToken string) error {
	_, err := d.userWriter.ExecContext(ctx, "UPDATE users_devices SET notification_token = $1 WHERE user_id = $2 AND device_identifier = $3;",
		notifyToken, userID, deviceID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: user mobile device not found", ErrNotFound)
	}
	return err
}

func (d *DataAccessService) GetAppSubscriptionCount(ctx context.Context, userID uint64) (uint64, error) {
	var count uint64
	err := d.userReader.GetContext(ctx, &count, "SELECT COUNT(receipt) FROM users_app_subscriptions WHERE user_id = $1", userID)
	return count, err
}

func (d *DataAccessService) AddMobilePurchase(ctx context.Context, tx *sql.Tx, userID uint64, paymentDetails t.MobileSubscription, verifyResponse *userservice.VerifyResponse, extSubscriptionId string) error {
	now := time.Now()
	nowTs := now.Unix()
	receiptHash := utils.HashAndEncode(verifyResponse.Receipt)

	query := `INSERT INTO users_app_subscriptions 
				(user_id, product_id, price_micros, currency, created_at, updated_at, validate_remotely, active, store, receipt, expires_at, reject_reason, receipt_hash, subscription_id) 
				VALUES($1, $2, $3, $4, TO_TIMESTAMP($5), TO_TIMESTAMP($6), $7, $8, $9, $10, TO_TIMESTAMP($11), $12, $13, $14) 
			  ON CONFLICT(receipt_hash) DO UPDATE SET product_id = $2, active = $7, updated_at = TO_TIMESTAMP($5);`
	var err error
	if tx == nil {
		_, err = d.userWriter.ExecContext(ctx, query,
			userID, verifyResponse.ProductID, paymentDetails.PriceMicros, paymentDetails.Currency, nowTs, nowTs, verifyResponse.Valid, verifyResponse.Valid, paymentDetails.Transaction.Type, verifyResponse.Receipt, verifyResponse.ExpirationDate, verifyResponse.RejectReason, receiptHash, extSubscriptionId,
		)
	} else {
		_, err = tx.ExecContext(ctx, query,
			userID, verifyResponse.ProductID, paymentDetails.PriceMicros, paymentDetails.Currency, nowTs, nowTs, verifyResponse.Valid, verifyResponse.Valid, paymentDetails.Transaction.Type, verifyResponse.Receipt, verifyResponse.ExpirationDate, verifyResponse.RejectReason, receiptHash, extSubscriptionId,
		)
	}

	return err
}

func (d *DataAccessService) GetLatestBundleForNativeVersion(ctx context.Context, nativeVersion uint64) (*t.MobileAppBundleStats, error) {
	var bundle t.MobileAppBundleStats
	err := d.userReader.GetContext(ctx, &bundle, `
		WITH 
			latest_native AS (
				SELECT max(min_native_version) as max_native_version 
				FROM mobile_app_bundles
			),
			latest_bundle AS (
				SELECT 
					bundle_version, 
					bundle_url, 
					delivered_count, 
					COALESCE(target_count, -1) as target_count
				FROM mobile_app_bundles 
				WHERE min_native_version <= $1 
				ORDER BY bundle_version DESC 
				LIMIT 1
			)
		SELECT
			COALESCE(latest_bundle.bundle_version, 0) as bundle_version,
			COALESCE(latest_bundle.bundle_url, '') as bundle_url,
			COALESCE(latest_bundle.target_count, -1) as target_count,
			COALESCE(latest_bundle.delivered_count, 0) as delivered_count,
			latest_native.max_native_version
		FROM latest_native
		LEFT JOIN latest_bundle ON TRUE;`,
		nativeVersion,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	return &bundle, err
}

func (d *DataAccessService) IncrementBundleDeliveryCount(ctx context.Context, bundleVersion uint64) error {
	_, err := d.userWriter.ExecContext(ctx, "UPDATE mobile_app_bundles SET delivered_count = COALESCE(delivered_count, 0) + 1 WHERE bundle_version = $1", bundleVersion)
	return err
}

func (d *DataAccessService) GetValidatorDashboardMobileWidget(ctx context.Context, dashboardId t.VDBIdPrimary) (*t.MobileWidgetData, error) {
	data := t.MobileWidgetData{}
	eg := errgroup.Group{}
	var err error

	wrappedDashboardId := t.VDBId{
		Id:              dashboardId,
		AggregateGroups: true,
	}

	// Get the average network efficiency
	efficiency, err := d.services.GetCurrentEfficiencyInfo()
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard overview data: %w", err)
	}
	data.NetworkEfficiency = utils.CalculateTotalEfficiency(
		efficiency.AttestationEfficiency[enums.AllTime], efficiency.ProposalEfficiency[enums.AllTime], efficiency.SyncEfficiency[enums.AllTime])

	protocolModes := t.VDBProtocolModes{RocketPool: true}
	rpInfos, err := d.getRocketPoolInfos(ctx, wrappedDashboardId, t.AllGroups)
	if err != nil {
		return nil, fmt.Errorf("error retrieving rocketpool infos: %w", err)
	}

	// Validator status
	eg.Go(func() error {
		validatorMapping, err := d.services.GetCurrentValidatorMapping()
		if err != nil {
			return err
		}

		validators, err := d.getDashboardValidators(ctx, wrappedDashboardId, nil)
		if err != nil {
			return fmt.Errorf("error retrieving validators from dashboard id: %w", err)
		}

		// Status
		for _, validator := range validators {
			metadata := validatorMapping.ValidatorMetadata[validator]

			switch constypes.ValidatorDbStatus(metadata.Status) {
			case constypes.DbExitingOnline, constypes.DbSlashingOnline, constypes.DbActiveOnline:
				data.ValidatorStateCounts.Online++
			case constypes.DbExitingOffline, constypes.DbSlashingOffline, constypes.DbActiveOffline:
				data.ValidatorStateCounts.Offline++
			case constypes.DbDeposited, constypes.DbPending:
				data.ValidatorStateCounts.Pending++
			case constypes.DbSlashed:
				data.ValidatorStateCounts.Slashed++
			case constypes.DbExited:
				data.ValidatorStateCounts.Exited++
			}
		}

		return nil
	})

	// RPL
	eg.Go(func() error {
		rpNetworkStats, err := d.getInternalRpNetworkStats(ctx)
		if err != nil {
			return fmt.Errorf("error retrieving rocketpool network stats: %w", err)
		}
		data.RplPrice = rpNetworkStats.RPLPrice

		// Find rocketpool node effective balance
		type RpOperatorInfo struct {
			EffectiveRPLStake decimal.Decimal `db:"effective_rpl_stake"`
			RPLStake          decimal.Decimal `db:"rpl_stake"`
		}
		var queryResult RpOperatorInfo

		ds := goqu.Dialect("postgres").
			Select(
				goqu.COALESCE(goqu.SUM("rpln.effective_rpl_stake"), 0).As("effective_rpl_stake"),
				goqu.COALESCE(goqu.SUM("rpln.rpl_stake"), 0).As("rpl_stake")).
			From(goqu.L("rocketpool_nodes AS rpln")).
			LeftJoin(goqu.L("rocketpool_minipools AS m"), goqu.On(goqu.L("m.node_address = rpln.address"))).
			LeftJoin(goqu.L("validators AS v"), goqu.On(goqu.L("m.pubkey = v.pubkey"))).
			Where(goqu.L("node_deposit_balance IS NOT NULL")).
			Where(goqu.L("user_deposit_balance IS NOT NULL")).
			LeftJoin(goqu.L("users_val_dashboards_validators uvdv"), goqu.On(goqu.L("uvdv.validator_index = v.validatorindex"))).
			Where(goqu.L("uvdv.dashboard_id = ?", dashboardId))

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.alloyReader.GetContext(ctx, &queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving rocketpool validators data: %w", err)
		}

		if !rpNetworkStats.EffectiveRPLStaked.IsZero() && !queryResult.EffectiveRPLStake.IsZero() && !rpNetworkStats.NodeOperatorRewards.IsZero() && rpNetworkStats.ClaimIntervalHours > 0 {
			share := queryResult.EffectiveRPLStake.Div(rpNetworkStats.EffectiveRPLStaked)

			periodsPerYear := decimal.NewFromFloat(365 / (rpNetworkStats.ClaimIntervalHours / 24))
			data.RplApr = rpNetworkStats.NodeOperatorRewards.
				Mul(share).
				Div(queryResult.RPLStake).
				Mul(periodsPerYear).
				Mul(decimal.NewFromInt(100)).InexactFloat64()
		}
		return nil
	})

	retrieveApr := func(hours int, apr *float64) {
		eg.Go(func() error {
			_, elApr, _, clApr, err := d.internal_getElClAPR(ctx, wrappedDashboardId, t.AllGroups, protocolModes, rpInfos, hours)
			if err != nil {
				return err
			}
			*apr = elApr + clApr
			return nil
		})
	}

	retrieveRewards := func(hours int, rewards *decimal.Decimal) {
		eg.Go(func() error {
			clRewards, _, elRewards, _, err := d.internal_getElClAPR(ctx, wrappedDashboardId, t.AllGroups, protocolModes, rpInfos, hours)
			if err != nil {
				return err
			}
			*rewards = clRewards.Add(elRewards)
			return nil
		})
	}

	retrieveEfficiency := func(table string, efficiency *float64) {
		eg.Go(func() error {
			ds := goqu.Dialect("postgres").
				From(goqu.L(fmt.Sprintf(`%s AS r FINAL`, table))).
				With("validators", goqu.L("(SELECT dashboard_id, validator_index FROM users_val_dashboards_validators WHERE dashboard_id = ?)", dashboardId)).
				Select(
					goqu.L("COALESCE(SUM(r.attestations_reward)::decimal, 0) AS attestations_reward"),
					goqu.L("COALESCE(SUM(r.attestations_ideal_reward)::decimal, 0) AS attestations_ideal_reward"),
					goqu.L("COALESCE(SUM(r.blocks_proposed), 0) AS blocks_proposed"),
					goqu.L("COALESCE(SUM(r.blocks_scheduled), 0) AS blocks_scheduled"),
					goqu.L("COALESCE(SUM(r.sync_executed), 0) AS sync_executed"),
					goqu.L("COALESCE(SUM(r.sync_scheduled), 0) AS sync_scheduled")).
				InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
				Where(goqu.L("r.validator_index IN (SELECT validator_index FROM validators)"))

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

	retrieveRewards(24, &data.Last24hIncome)
	retrieveRewards(7*24, &data.Last7dIncome)
	retrieveApr(30*24, &data.Last30dApr)
	retrieveEfficiency("validator_dashboard_data_rolling_30d", &data.Last30dEfficiency)

	err = eg.Wait()

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard overview data: %w", err)
	}

	return &data, nil
}

func (d *DataAccessService) getInternalRpNetworkStats(ctx context.Context) (*t.RPNetworkStats, error) {
	var networkStats t.RPNetworkStats
	err := d.alloyReader.GetContext(ctx, &networkStats, `
			SELECT 
				EXTRACT(EPOCH FROM claim_interval_time) / 3600 AS claim_interval_hours,
				node_operator_rewards,
				effective_rpl_staked,
				rpl_price 
			FROM rocketpool_network_stats 
			ORDER BY ID 
			DESC 
			LIMIT 1
		`)
	return &networkStats, err
}

func (d *DataAccessService) GetValidatorDashboardMobileValidators(ctx context.Context, dashboardId t.VDBId, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBMobileValidatorsColumn], search string, limit uint64) ([]t.MobileValidatorDashboardValidatorsTableRow, *t.Paging, error) {
	return d.dummy.GetValidatorDashboardMobileValidators(ctx, dashboardId, period, cursor, colSort, search, limit)
}
