package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (uint64, error)
	CreateUser(ctx context.Context, email, password string) (uint64, error)
	RemoveUser(ctx context.Context, userId uint64) error
	UpdateUserEmail(ctx context.Context, userId uint64) error
	UpdateUserPassword(ctx context.Context, userId uint64, password string) error
	GetEmailConfirmationTime(ctx context.Context, userId uint64) (time.Time, error)
	GetPasswordResetTime(ctx context.Context, userId uint64) (time.Time, error)
	IsPasswordResetAllowed(ctx context.Context, userId uint64) (bool, error)
	UpdateEmailConfirmationTime(ctx context.Context, userId uint64) error
	UpdatePasswordResetTime(ctx context.Context, userId uint64) error
	UpdateEmailConfirmationHash(ctx context.Context, userId uint64, email, confirmationHash string) error
	UpdatePasswordResetHash(ctx context.Context, userId uint64, passwordHash string) error
	GetUserCredentialInfo(ctx context.Context, userId uint64) (*t.UserCredentialInfo, error)
	GetUserIdByApiKey(ctx context.Context, apiKey string) (uint64, error)
	GetUserIdByConfirmationHash(ctx context.Context, hash string) (uint64, error)
	GetUserIdByResetHash(ctx context.Context, hash string) (uint64, error)
	GetUserInfo(ctx context.Context, id uint64) (*t.UserInfo, error)
	GetUserDashboards(ctx context.Context, userId uint64) (*t.UserDashboardsData, error)
	GetUserValidatorDashboardCount(ctx context.Context, userId uint64, active bool) (uint64, error)
}

func (d *DataAccessService) GetUserByEmail(ctx context.Context, email string) (uint64, error) {
	result := uint64(0)
	err := d.userReader.GetContext(ctx, &result, `SELECT id FROM users WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: user not found", ErrNotFound)
	}
	return result, err
}

func (d *DataAccessService) CreateUser(ctx context.Context, email, password string) (uint64, error) {
	// (password is already hashed)
	var result uint64

	apiKey, err := utils.GenerateRandomAPIKey()
	if err != nil {
		return 0, err
	}

	err = d.userWriter.GetContext(ctx, &result, `
    	INSERT INTO users (password, email, register_ts, api_key)
      		VALUES ($1, $2, NOW(), $3)
		RETURNING id`,
		password, email, apiKey,
	)

	return result, err
}

func (d *DataAccessService) RemoveUser(ctx context.Context, userId uint64) error {
	_, err := d.userWriter.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userId)
	return err
}

func (d *DataAccessService) UpdateUserEmail(ctx context.Context, userId uint64) error {
	// Called after user clicked link for email confirmations + changes, so:
	// set email_confirmed true, set email (from email_change_to_value), update stripe email
	// unset email_confirmation_hash

	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			email = email_change_to_value,
			email_change_to_value = NULL,
			email_confirmed = true,
			email_confirmation_hash = NULL,
			stripe_email_pending = true
		WHERE id = $1
	`, userId)

	return err
}

func (d *DataAccessService) UpdateUserPassword(ctx context.Context, userId uint64, password string) error {
	// (password is already hashed)

	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			password = $1,
			password_reset_hash = NULL
		WHERE id = $2
	`, password, userId)

	return err
}

func (d *DataAccessService) GetEmailConfirmationTime(ctx context.Context, userId uint64) (time.Time, error) {
	result := time.Time{}

	var queryResult sql.NullTime
	err := d.userReader.GetContext(ctx, &queryResult, `
    	SELECT
			email_confirmation_ts
		FROM users
		WHERE id = $1`, userId)

	if queryResult.Valid {
		result = queryResult.Time
	}

	return result, err
}

func (d *DataAccessService) GetPasswordResetTime(ctx context.Context, userId uint64) (time.Time, error) {
	result := time.Time{}

	var queryResult sql.NullTime
	err := d.userReader.GetContext(ctx, &queryResult, `
    	SELECT
			password_reset_ts
		FROM users
		WHERE id = $1`, userId)

	if queryResult.Valid {
		result = queryResult.Time
	}

	return result, err
}

func (d *DataAccessService) UpdateEmailConfirmationTime(ctx context.Context, userId uint64) error {
	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			email_confirmation_ts = NOW()
		WHERE id = $1
	`, userId)

	return err
}

func (d *DataAccessService) IsPasswordResetAllowed(ctx context.Context, userId uint64) (bool, error) {
	var result bool

	err := d.userReader.GetContext(ctx, &result, `
    	SELECT
			password_reset_not_allowed
		FROM users
		WHERE id = $1`, userId)

	return !result, err
}

func (d *DataAccessService) UpdatePasswordResetTime(ctx context.Context, userId uint64) error {
	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			password_reset_ts = NOW()
		WHERE id = $1
	`, userId)

	return err
}

func (d *DataAccessService) UpdateEmailConfirmationHash(ctx context.Context, userId uint64, email, confirmationHash string) error {
	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			email_confirmation_hash = $1,
			email_change_to_value = $2
		WHERE id = $3
	`, confirmationHash, email, userId)

	return err
}

func (d *DataAccessService) UpdatePasswordResetHash(ctx context.Context, userId uint64, confirmationHash string) error {
	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			password_reset_hash = $1
		WHERE id = $2
	`, confirmationHash, userId)

	return err
}

func (d *DataAccessService) GetUserCredentialInfo(ctx context.Context, userId uint64) (*t.UserCredentialInfo, error) {
	// TODO @patrick post-beta improve product-mgmt
	// TODO @DATA-ACCESS i quickly hacked this together, maybe improve
	result := &t.UserCredentialInfo{}
	err := d.userReader.GetContext(ctx, result, `
		WITH
			latest_and_greatest_sub AS (
				SELECT user_id, product_id FROM users_app_subscriptions
				LEFT JOIN users ON users.id = user_id AND product_id IN ('orca.yearly', 'orca', 'dolphin.yearly', 'dolphin', 'guppy.yearly', 'guppy', 'whale', 'goldfish', 'plankton')
				WHERE users.id = $1 AND active = true
				ORDER BY CASE product_id
					WHEN 'orca.yearly'    THEN  1
					WHEN 'orca'           THEN  2
					WHEN 'dolphin.yearly' THEN  3
					WHEN 'dolphin'        THEN  4
					WHEN 'guppy.yearly'   THEN  5
					WHEN 'guppy'          THEN  6
					WHEN 'whale'          THEN  7
					WHEN 'goldfish'       THEN  8
					WHEN 'plankton'       THEN  9
					ELSE                       10  -- For any other product_id values
				END, users_app_subscriptions.created_at DESC LIMIT 1
			)
		SELECT users.id AS id, users.email, users.email_confirmed, password, COALESCE(product_id, '') AS product_id, COALESCE(user_group, '') AS user_group
		FROM users
		LEFT JOIN latest_and_greatest_sub ON latest_and_greatest_sub.user_id = users.id
		WHERE users.id = $1`, userId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: user not found", ErrNotFound)
	}
	return result, err
}

func (d *DataAccessService) GetUserIdByApiKey(ctx context.Context, apiKey string) (uint64, error) {
	var userId uint64
	err := d.userReader.GetContext(ctx, &userId, `SELECT user_id FROM api_keys WHERE api_key = $1 LIMIT 1`, apiKey)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: user for api_key not found", ErrNotFound)
	}
	return userId, err
}

func (d *DataAccessService) GetUserIdByConfirmationHash(ctx context.Context, hash string) (uint64, error) {
	var result uint64

	err := d.userReader.GetContext(ctx, &result, `
    	SELECT
			id
		FROM users
		WHERE email_confirmation_hash = $1`, hash)

	return result, err
}

func (d *DataAccessService) GetUserIdByResetHash(ctx context.Context, hash string) (uint64, error) {
	var result uint64

	err := d.userReader.GetContext(ctx, &result, `
    	SELECT
			id
		FROM users
		WHERE password_reset_hash = $1`, hash)

	return result, err
}

func (d *DataAccessService) GetUserInfo(ctx context.Context, userId uint64) (*t.UserInfo, error) {
	// TODO @patrick post-beta improve and unmock
	userInfo := &t.UserInfo{
		Id:      userId,
		ApiKeys: []string{},
		ApiPerks: t.ApiPerks{
			UnitsPerSecond:    10,
			UnitsPerMonth:     10,
			ApiKeys:           4,
			ConsensusLayerAPI: true,
			ExecutionLayerAPI: true,
			Layer2API:         true,
			NoAds:             true,
			DiscordSupport:    false,
		},
		Subscriptions: []t.UserSubscription{},
	}

	productSummary, err := d.GetProductSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting productSummary: %w", err)
	}

	result := struct {
		Email     string `db:"email"`
		UserGroup string `db:"user_group"`
	}{}
	err = d.userReader.GetContext(ctx, &result, `SELECT email, COALESCE(user_group, '') as user_group FROM users WHERE id = $1`, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting userEmail: %w", err)
	}
	userInfo.Email = result.Email
	userInfo.UserGroup = result.UserGroup

	userInfo.Email = utils.CensorEmail(userInfo.Email)

	err = d.userReader.SelectContext(ctx, &userInfo.ApiKeys, `SELECT api_key FROM api_keys WHERE user_id = $1`, userId)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting userApiKeys: %w", err)
	}

	premiumProduct := struct {
		ProductId string    `db:"product_id"`
		Store     string    `db:"store"`
		Start     time.Time `db:"start"`
		End       time.Time `db:"end"`
	}{}
	err = d.userReader.GetContext(ctx, &premiumProduct, `
		SELECT
			COALESCE(uas.product_id, '') AS product_id,
			COALESCE(uas.store, '') AS store,
			COALESCE(to_timestamp((uss.payload->>'current_period_start')::bigint),uas.created_at) AS start,
			COALESCE(to_timestamp((uss.payload->>'current_period_end')::bigint),uas.expires_at) AS end
		FROM users_app_subscriptions uas
		LEFT JOIN users_stripe_subscriptions uss ON uss.subscription_id = uas.subscription_id
		WHERE uas.user_id = $1 AND uas.active = true AND product_id IN ('orca.yearly', 'orca', 'dolphin.yearly', 'dolphin', 'guppy.yearly', 'guppy', 'whale', 'goldfish', 'plankton')
		ORDER BY CASE uas.product_id
			WHEN 'orca.yearly'    THEN  1
			WHEN 'orca'           THEN  2
			WHEN 'dolphin.yearly' THEN  3
			WHEN 'dolphin'        THEN  4
			WHEN 'guppy.yearly'   THEN  5
			WHEN 'guppy'          THEN  6
			WHEN 'whale'          THEN  7
			WHEN 'goldfish'       THEN  8
			WHEN 'plankton'       THEN  9
			ELSE                       10  -- For any other product_id values
		END, uas.id DESC
		LIMIT 1`, userId)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("error getting premiumProduct: %w", err)
		}
		premiumProduct.ProductId = "premium_free"
		premiumProduct.Store = ""
	}

	foundProduct := false
	for _, p := range productSummary.PremiumProducts {
		effectiveProductId := premiumProduct.ProductId
		productName := p.ProductName
		switch premiumProduct.ProductId {
		case "whale":
			effectiveProductId = "dolphin"
			productName = "Whale"
		case "goldfish":
			effectiveProductId = "guppy"
			productName = "Goldfish"
		case "plankton":
			effectiveProductId = "guppy"
			productName = "Plankton"
		}
		if p.ProductIdMonthly == effectiveProductId || p.ProductIdYearly == effectiveProductId {
			userInfo.PremiumPerks = p.PremiumPerks
			foundProduct = true

			store := t.ProductStoreStripe
			switch premiumProduct.Store {
			case "ios-appstore":
				store = t.ProductStoreIosAppstore
			case "android-playstore":
				store = t.ProductStoreAndroidPlaystore
			case "ethpool":
				store = t.ProductStoreEthpool
			case "manuall":
				store = t.ProductStoreCustom
			}

			if effectiveProductId != "premium_free" {
				userInfo.Subscriptions = append(userInfo.Subscriptions, t.UserSubscription{
					ProductId:       premiumProduct.ProductId,
					ProductName:     productName,
					ProductCategory: t.ProductCategoryPremium,
					ProductStore:    store,
					Start:           premiumProduct.Start.Unix(),
					End:             premiumProduct.End.Unix(),
				})
			}
			break
		}
	}
	if !foundProduct {
		return nil, fmt.Errorf("product %s not found", premiumProduct.ProductId)
	}

	premiumAddons := []struct {
		PriceId  string    `db:"price_id"`
		Start    time.Time `db:"start"`
		End      time.Time `db:"end"`
		Quantity int       `db:"quantity"`
	}{}
	err = d.userReader.SelectContext(ctx, &premiumAddons, `
		SELECT
			price_id,
			to_timestamp((uss.payload->>'current_period_start')::bigint) AS start,
			to_timestamp((uss.payload->>'current_period_end')::bigint) AS end,
			COALESCE((uss.payload->>'quantity')::int,1) AS quantity
		FROM users_stripe_subscriptions uss
		INNER JOIN users u ON u.stripe_customer_id = uss.customer_id
		WHERE u.id = $1 AND uss.active = true AND uss.purchase_group = 'addon'`, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting premiumAddons: %w", err)
	}
	for _, addon := range premiumAddons {
		foundAddon := false
		for _, p := range productSummary.ExtraDashboardValidatorsPremiumAddon {
			if p.StripePriceIdMonthly == addon.PriceId || p.StripePriceIdYearly == addon.PriceId {
				foundAddon = true
				for i := 0; i < addon.Quantity; i++ {
					userInfo.PremiumPerks.ValidatorsPerDashboard += p.ExtraDashboardValidators
					userInfo.Subscriptions = append(userInfo.Subscriptions, t.UserSubscription{
						ProductId:       utils.PriceIdToProductId(addon.PriceId),
						ProductName:     p.ProductName,
						ProductCategory: t.ProductCategoryPremiumAddon,
						ProductStore:    t.ProductStoreStripe,
						Start:           addon.Start.Unix(),
						End:             addon.End.Unix(),
					})
				}
			}
		}
		if !foundAddon {
			return nil, fmt.Errorf("addon not found: %v", addon.PriceId)
		}
	}

	if productSummary.ValidatorsPerDashboardLimit < userInfo.PremiumPerks.ValidatorsPerDashboard {
		userInfo.PremiumPerks.ValidatorsPerDashboard = productSummary.ValidatorsPerDashboardLimit
	}

	return userInfo, nil
}

const hour uint64 = 3600
const day = 24 * hour
const week = 7 * day
const month = 30 * day
const fullHistory uint64 = 9007199254740991 // 2^53-1 (max int in JS)

var freeTierProduct t.PremiumProduct = t.PremiumProduct{
	ProductName: "Free",
	PremiumPerks: t.PremiumPerks{
		AdFree:                      false,
		ValidatorDasboards:          1,
		ValidatorsPerDashboard:      20,
		ValidatorGroupsPerDashboard: 1,
		ShareCustomDashboards:       false,
		ManageDashboardViaApi:       false,
		BulkAdding:                  false,
		ChartHistorySeconds: t.ChartHistorySeconds{
			Epoch:  0,
			Hourly: 12 * hour,
			Daily:  0,
			Weekly: 0,
		},
		EmailNotificationsPerDay:        5,
		ConfigureNotificationsViaApi:    false,
		ValidatorGroupNotifications:     1,
		WebhookEndpoints:                1,
		MobileAppCustomThemes:           false,
		MobileAppWidget:                 false,
		MonitorMachines:                 1,
		MachineMonitoringHistorySeconds: 3600 * 3,
		CustomMachineAlerts:             false,
	},
	PricePerMonthEur: 0,
	PricePerYearEur:  0,
	ProductIdMonthly: "premium_free",
	ProductIdYearly:  "premium_free.yearly",
}

func (d *DataAccessService) GetProductSummary(ctx context.Context) (*t.ProductSummary, error) {
	// TODO @patrick post-beta put into db instead of hardcoding here and make it configurable
	return &t.ProductSummary{
		ValidatorsPerDashboardLimit: 102_000,
		StripePublicKey:             utils.Config.Frontend.Stripe.PublicKey,
		ApiProducts: []t.ApiProduct{ // TODO @patrick post-beta this data is not final yet
			{
				ProductId:        "api_free",
				ProductName:      "Free",
				PricePerMonthEur: 0,
				PricePerYearEur:  0 * 12,
				ApiPerks: t.ApiPerks{
					UnitsPerSecond:    10,
					UnitsPerMonth:     10_000_000,
					ApiKeys:           2,
					ConsensusLayerAPI: true,
					ExecutionLayerAPI: true,
					Layer2API:         true,
					NoAds:             true,
					DiscordSupport:    false,
				},
			},
			{
				ProductId:        "iron",
				ProductName:      "Iron",
				PricePerMonthEur: 1.99,
				PricePerYearEur:  math.Floor(1.99*12*0.9*100) / 100,
				ApiPerks: t.ApiPerks{
					UnitsPerSecond:    20,
					UnitsPerMonth:     20_000_000,
					ApiKeys:           10,
					ConsensusLayerAPI: true,
					ExecutionLayerAPI: true,
					Layer2API:         true,
					NoAds:             true,
					DiscordSupport:    false,
				},
			},
			{
				ProductId:        "silver",
				ProductName:      "Silver",
				PricePerMonthEur: 2.99,
				PricePerYearEur:  math.Floor(2.99*12*0.9*100) / 100,
				ApiPerks: t.ApiPerks{
					UnitsPerSecond:    30,
					UnitsPerMonth:     100_000_000,
					ApiKeys:           20,
					ConsensusLayerAPI: true,
					ExecutionLayerAPI: true,
					Layer2API:         true,
					NoAds:             true,
					DiscordSupport:    false,
				},
			},
			{
				ProductId:        "gold",
				ProductName:      "Gold",
				PricePerMonthEur: 3.99,
				PricePerYearEur:  math.Floor(3.99*12*0.9*100) / 100,
				ApiPerks: t.ApiPerks{
					UnitsPerSecond:    40,
					UnitsPerMonth:     200_000_000,
					ApiKeys:           40,
					ConsensusLayerAPI: true,
					ExecutionLayerAPI: true,
					Layer2API:         true,
					NoAds:             true,
					DiscordSupport:    false,
				},
			},
		},
		PremiumProducts: []t.PremiumProduct{
			freeTierProduct,
			{
				ProductName: "Guppy",
				PremiumPerks: t.PremiumPerks{
					AdFree:                      true,
					ValidatorDasboards:          1,
					ValidatorsPerDashboard:      100,
					ValidatorGroupsPerDashboard: 3,
					ShareCustomDashboards:       true,
					ManageDashboardViaApi:       false,
					BulkAdding:                  true,
					ChartHistorySeconds: t.ChartHistorySeconds{
						Epoch:  day,
						Hourly: 7 * day,
						Daily:  month,
						Weekly: 0,
					},
					EmailNotificationsPerDay:        15,
					ConfigureNotificationsViaApi:    false,
					ValidatorGroupNotifications:     3,
					WebhookEndpoints:                3,
					MobileAppCustomThemes:           true,
					MobileAppWidget:                 true,
					MonitorMachines:                 2,
					MachineMonitoringHistorySeconds: 3600 * 24 * 30,
					CustomMachineAlerts:             true,
				},
				PricePerMonthEur:     9.99,
				PricePerYearEur:      107.88,
				ProductIdMonthly:     "guppy",
				ProductIdYearly:      "guppy.yearly",
				StripePriceIdMonthly: utils.Config.Frontend.Stripe.Guppy,
				StripePriceIdYearly:  utils.Config.Frontend.Stripe.GuppyYearly,
			},
			{
				ProductName: "Dolphin",
				PremiumPerks: t.PremiumPerks{
					AdFree:                      true,
					ValidatorDasboards:          2,
					ValidatorsPerDashboard:      300,
					ValidatorGroupsPerDashboard: 10,
					ShareCustomDashboards:       true,
					ManageDashboardViaApi:       false,
					BulkAdding:                  true,
					ChartHistorySeconds: t.ChartHistorySeconds{
						Epoch:  5 * day,
						Hourly: month,
						Daily:  2 * month,
						Weekly: 8 * week,
					},
					EmailNotificationsPerDay:        20,
					ConfigureNotificationsViaApi:    false,
					ValidatorGroupNotifications:     10,
					WebhookEndpoints:                10,
					MobileAppCustomThemes:           true,
					MobileAppWidget:                 true,
					MonitorMachines:                 10,
					MachineMonitoringHistorySeconds: 3600 * 24 * 30,
					CustomMachineAlerts:             true,
				},
				PricePerMonthEur:     29.99,
				PricePerYearEur:      311.88,
				ProductIdMonthly:     "dolphin",
				ProductIdYearly:      "dolphin.yearly",
				StripePriceIdMonthly: utils.Config.Frontend.Stripe.Dolphin,
				StripePriceIdYearly:  utils.Config.Frontend.Stripe.DolphinYearly,
			},
			{
				ProductName: "Orca",
				PremiumPerks: t.PremiumPerks{
					AdFree:                      true,
					ValidatorDasboards:          2,
					ValidatorsPerDashboard:      1000,
					ValidatorGroupsPerDashboard: 30,
					ShareCustomDashboards:       true,
					ManageDashboardViaApi:       true,
					BulkAdding:                  true,
					ChartHistorySeconds: t.ChartHistorySeconds{
						Epoch:  3 * week,
						Hourly: 6 * month,
						Daily:  12 * month,
						Weekly: fullHistory,
					},
					EmailNotificationsPerDay:        50,
					ConfigureNotificationsViaApi:    true,
					ValidatorGroupNotifications:     60,
					WebhookEndpoints:                30,
					MobileAppCustomThemes:           true,
					MobileAppWidget:                 true,
					MonitorMachines:                 10,
					MachineMonitoringHistorySeconds: 3600 * 24 * 30,
					CustomMachineAlerts:             true,
				},
				PricePerMonthEur:     49.99,
				PricePerYearEur:      479.88,
				ProductIdMonthly:     "orca",
				ProductIdYearly:      "orca.yearly",
				StripePriceIdMonthly: utils.Config.Frontend.Stripe.Orca,
				StripePriceIdYearly:  utils.Config.Frontend.Stripe.OrcaYearly,
				IsPopular:            true,
			},
		},
		ExtraDashboardValidatorsPremiumAddon: []t.ExtraDashboardValidatorsPremiumAddon{
			{
				ProductName:              "1k extra valis per dashboard",
				ExtraDashboardValidators: 1000,
				PricePerMonthEur:         74.99,
				PricePerYearEur:          719.88,
				ProductIdMonthly:         "vdb_addon_1k",
				ProductIdYearly:          "vdb_addon_1k.yearly",
				StripePriceIdMonthly:     utils.Config.Frontend.Stripe.VdbAddon1k,
				StripePriceIdYearly:      utils.Config.Frontend.Stripe.VdbAddon1kYearly,
			},
			{
				ProductName:              "10k extra valis per dashboard",
				ExtraDashboardValidators: 10000,
				PricePerMonthEur:         449.99,
				PricePerYearEur:          4319.88,
				ProductIdMonthly:         "vdb_addon_10k",
				ProductIdYearly:          "vdb_addon_10k.yearly",
				StripePriceIdMonthly:     utils.Config.Frontend.Stripe.VdbAddon10k,
				StripePriceIdYearly:      utils.Config.Frontend.Stripe.VdbAddon10kYearly,
			},
		},
	}, nil
}

func (d *DataAccessService) GetFreeTierPerks(ctx context.Context) (*t.PremiumPerks, error) {
	return &freeTierProduct.PremiumPerks, nil
}

func (d *DataAccessService) GetUserDashboards(ctx context.Context, userId uint64) (*t.UserDashboardsData, error) {
	result := &t.UserDashboardsData{}

	wg := errgroup.Group{}

	validatorDashboardMap := make(map[uint64]*t.ValidatorDashboard, 0)
	wg.Go(func() error {
		dbReturn := []struct {
			Id           uint64         `db:"id"`
			Name         string         `db:"name"`
			IsArchived   sql.NullString `db:"is_archived"`
			PublicId     sql.NullString `db:"public_id"`
			PublicName   sql.NullString `db:"public_name"`
			SharedGroups sql.NullBool   `db:"shared_groups"`
		}{}

		err := d.alloyReader.SelectContext(ctx, &dbReturn, `
		SELECT
			uvd.id,
			uvd.name,
			uvd.is_archived,
			uvds.public_id,
			uvds.name AS public_name,
			uvds.shared_groups
		FROM users_val_dashboards uvd
		LEFT JOIN users_val_dashboards_sharing uvds ON uvd.id = uvds.dashboard_id
		WHERE uvd.user_id = $1
	`, userId)
		if err != nil {
			return err
		}

		for _, row := range dbReturn {
			if _, ok := validatorDashboardMap[row.Id]; !ok {
				validatorDashboardMap[row.Id] = &t.ValidatorDashboard{
					Id:             row.Id,
					Name:           row.Name,
					PublicIds:      []t.VDBPublicId{},
					IsArchived:     row.IsArchived.Valid,
					ArchivedReason: row.IsArchived.String,
				}
			}
			if row.PublicId.Valid {
				publicId := t.VDBPublicId{}
				publicId.PublicId = row.PublicId.String
				publicId.Name = row.PublicName.String
				publicId.ShareSettings.ShareGroups = row.SharedGroups.Bool

				validatorDashboardMap[row.Id].PublicIds = append(validatorDashboardMap[row.Id].PublicIds, publicId)
			}
		}

		return nil
	})

	type DashboardCount struct {
		Id             uint64 `db:"id"`
		GroupCount     uint64 `db:"group_count"`
		ValidatorCount uint64 `db:"validator_count"`
	}

	validatorDashboardCountMap := make(map[uint64]DashboardCount, 0)
	wg.Go(func() error {
		dbReturn := []DashboardCount{}

		err := d.alloyReader.SelectContext(ctx, &dbReturn, `
		SELECT
			uvd.id,
			COUNT(DISTINCT(uvdg.id)) AS group_count,
			COUNT(DISTINCT(uvdv.validator_index)) AS validator_count
		FROM users_val_dashboards uvd
		LEFT JOIN users_val_dashboards_groups uvdg ON uvd.id = uvdg.dashboard_id
		LEFT JOIN users_val_dashboards_validators uvdv ON uvd.id = uvdv.dashboard_id
		WHERE uvd.user_id = $1
		GROUP BY uvd.id
	`, userId)
		if err != nil {
			return err
		}

		for _, row := range dbReturn {
			validatorDashboardCountMap[row.Id] = row
		}

		return nil
	})

	err := wg.Wait()
	if err != nil {
		return nil, fmt.Errorf("error retrieving user dashboards data: %v", err)
	}

	// Fill the result
	for _, validatorDashboard := range validatorDashboardMap {
		validatorDashboard.GroupCount = validatorDashboardCountMap[validatorDashboard.Id].GroupCount
		validatorDashboard.ValidatorCount = validatorDashboardCountMap[validatorDashboard.Id].ValidatorCount

		result.ValidatorDashboards = append(result.ValidatorDashboards, *validatorDashboard)
	}

	// Get the account dashboards
	err = d.alloyReader.SelectContext(ctx, &result.AccountDashboards, `
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

// return number of active / archived dashboards
func (d *DataAccessService) GetUserValidatorDashboardCount(ctx context.Context, userId uint64, active bool) (uint64, error) {
	var count uint64
	err := d.alloyReader.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM users_val_dashboards
		WHERE user_id = $1 AND (($2 AND is_archived IS NULL) OR (NOT $2 AND is_archived IS NOT NULL))
	`, userId, active)

	return count, err
}
