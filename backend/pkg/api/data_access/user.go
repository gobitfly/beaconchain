package dataaccess

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/pkg/errors"
)

type UserRepository interface {
	GetUserCredentialInfo(email string) (*t.UserCredentialInfo, error)
	GetUserIdByApiKey(apiKey string) (uint64, error)
	GetUserInfo(id uint64) (*t.UserInfo, error)
	GetUserDashboards(userId uint64) (*t.UserDashboardsData, error)
	GetUserValidatorDashboardCount(userId uint64) (uint64, error)
}

func (d *DataAccessService) GetUserCredentialInfo(email string) (*t.UserCredentialInfo, error) {
	// TODO @patrick post-beta improve product-mgmt
	result := &t.UserCredentialInfo{}
	err := d.userReader.Get(result, `
		WITH
			latest_and_greatest_sub AS (
				SELECT user_id, product_id FROM users_app_subscriptions 
				LEFT JOIN users ON users.id = user_id AND product_id IN ('orca.yearly', 'orca', 'dolphin.yearly', 'dolphin', 'guppy.yearly', 'guppy', 'whale', 'goldfish', 'plankton')
				WHERE users.email = $1 AND active = true
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
		SELECT users.id AS id, password, COALESCE(product_id, '') AS product_id, COALESCE(user_group, '') AS user_group 
		FROM users
		LEFT JOIN latest_and_greatest_sub ON latest_and_greatest_sub.user_id = users.id  
		WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: user with email %s not found", ErrNotFound, email)
	}
	return result, err
}

func (d *DataAccessService) GetUserIdByApiKey(apiKey string) (uint64, error) {
	var userId uint64
	err := d.userReader.Get(&userId, `SELECT user_id FROM api_keys WHERE api_key = $1 LIMIT 1`, apiKey)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: user for api_key not found", ErrNotFound)
	}
	return userId, err
}

func (d *DataAccessService) GetUserInfo(userId uint64) (*t.UserInfo, error) {
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

	productSummary, err := d.GetProductSummary()
	if err != nil {
		return nil, fmt.Errorf("error getting productSummary: %w", err)
	}

	err = d.userReader.Get(&userInfo.Email, `SELECT email FROM users WHERE id = $1`, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting userEmail: %w", err)
	}

	err = d.userReader.Select(&userInfo.ApiKeys, `SELECT api_key FROM api_keys WHERE user_id = $1`, userId)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting userApiKeys: %w", err)
	}

	premiumProduct := struct {
		ProductId string    `db:"product_id"`
		Store     string    `db:"store"`
		Start     time.Time `db:"start"`
		End       time.Time `db:"end"`
	}{}
	err = d.userReader.Get(&premiumProduct, `
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
	err = d.userReader.Select(&premiumAddons, `
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

func (d *DataAccessService) GetProductSummary() (*t.ProductSummary, error) {
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
			{
				ProductName: "Free",
				PremiumPerks: t.PremiumPerks{
					AdFree:                          false,
					ValidatorDasboards:              1,
					ValidatorsPerDashboard:          20,
					ValidatorGroupsPerDashboard:     1,
					ShareCustomDashboards:           false,
					ManageDashboardViaApi:           false,
					HeatmapHistorySeconds:           0,
					SummaryChartHistorySeconds:      3600 * 12,
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
			},
			{
				ProductName: "Guppy",
				PremiumPerks: t.PremiumPerks{
					AdFree:                          true,
					ValidatorDasboards:              1,
					ValidatorsPerDashboard:          100,
					ValidatorGroupsPerDashboard:     3,
					ShareCustomDashboards:           true,
					ManageDashboardViaApi:           false,
					HeatmapHistorySeconds:           3600 * 24 * 7,
					SummaryChartHistorySeconds:      3600 * 24 * 7,
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
					AdFree:                          true,
					ValidatorDasboards:              2,
					ValidatorsPerDashboard:          300,
					ValidatorGroupsPerDashboard:     10,
					ShareCustomDashboards:           true,
					ManageDashboardViaApi:           false,
					HeatmapHistorySeconds:           3600 * 24 * 30,
					SummaryChartHistorySeconds:      3600 * 24 * 14,
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
					AdFree:                          true,
					ValidatorDasboards:              2,
					ValidatorsPerDashboard:          1000,
					ValidatorGroupsPerDashboard:     30,
					ShareCustomDashboards:           true,
					ManageDashboardViaApi:           true,
					HeatmapHistorySeconds:           3600 * 24 * 365,
					SummaryChartHistorySeconds:      3600 * 24 * 365,
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
			result.ShareSettings.ShareGroups = row.SharedGroups.Bool

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

func (d *DataAccessService) GetUserValidatorDashboardCount(userId uint64) (uint64, error) {
	var count uint64
	err := d.alloyReader.Get(&count, `
		SELECT COUNT(*) FROM users_val_dashboards
		WHERE user_id = $1
	`, userId)
	return count, err
}
