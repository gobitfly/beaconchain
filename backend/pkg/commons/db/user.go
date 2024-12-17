package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("not found")

const hour uint64 = 3600
const day = 24 * hour
const week = 7 * day
const month = 30 * day
const maxJsInt uint64 = 9007199254740991 // 2^53-1 (max safe int in JS)

var freeTierProduct t.PremiumProduct = t.PremiumProduct{
	ProductName: "Free",
	PremiumPerks: t.PremiumPerks{
		AdFree:                      false,
		ValidatorDashboards:         1,
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
		RewardsChartHistorySeconds: t.ChartHistorySeconds{
			Epoch:  0,
			Hourly: 0,
			Daily:  0,
			Weekly: maxJsInt,
		},
		EmailNotificationsPerDay:                       10,
		ConfigureNotificationsViaApi:                   false,
		ValidatorGroupNotifications:                    1,
		WebhookEndpoints:                               1,
		MobileAppCustomThemes:                          false,
		MobileAppWidget:                                false,
		MonitorMachines:                                1,
		MachineMonitoringHistorySeconds:                3600 * 3,
		NotificationsMachineCustomThreshold:            false,
		NotificationsValidatorDashboardGroupEfficiency: false,
	},
	PricePerMonthEur: 0,
	PricePerYearEur:  0,
	ProductIdMonthly: "premium_free",
	ProductIdYearly:  "premium_free.yearly",
}

var adminPerks = t.PremiumPerks{
	AdFree:                      false, // admins want to see ads to check ad configuration
	ValidatorDashboards:         maxJsInt,
	ValidatorsPerDashboard:      maxJsInt,
	ValidatorGroupsPerDashboard: maxJsInt,
	ShareCustomDashboards:       true,
	ManageDashboardViaApi:       true,
	BulkAdding:                  true,
	ChartHistorySeconds: t.ChartHistorySeconds{
		Epoch:  maxJsInt,
		Hourly: maxJsInt,
		Daily:  maxJsInt,
		Weekly: maxJsInt,
	},
	RewardsChartHistorySeconds: t.ChartHistorySeconds{
		Epoch:  maxJsInt,
		Hourly: maxJsInt,
		Daily:  maxJsInt,
		Weekly: maxJsInt,
	},
	EmailNotificationsPerDay:                       maxJsInt,
	ConfigureNotificationsViaApi:                   true,
	ValidatorGroupNotifications:                    maxJsInt,
	WebhookEndpoints:                               maxJsInt,
	MobileAppCustomThemes:                          true,
	MobileAppWidget:                                true,
	MonitorMachines:                                maxJsInt,
	MachineMonitoringHistorySeconds:                maxJsInt,
	NotificationsMachineCustomThreshold:            true,
	NotificationsValidatorDashboardGroupEfficiency: true,
}

func GetUserInfo(ctx context.Context, userId uint64, userDbReader *sqlx.DB) (*t.UserInfo, error) {
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

	productSummary, err := GetProductSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting productSummary: %w", err)
	}

	result := struct {
		Email     string `db:"email"`
		UserGroup string `db:"user_group"`
	}{}
	err = userDbReader.GetContext(ctx, &result, `SELECT email, COALESCE(user_group, '') as user_group FROM users WHERE id = $1`, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: user not found", ErrNotFound)
		}
		return nil, err
	}
	userInfo.Email = result.Email
	userInfo.UserGroup = result.UserGroup

	userInfo.Email = utils.CensorEmail(userInfo.Email)

	err = userDbReader.SelectContext(ctx, &userInfo.ApiKeys, `SELECT api_key FROM api_keys WHERE user_id = $1`, userId)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting userApiKeys for user %v: %w", userId, err)
	}

	premiumProduct := struct {
		ProductId string    `db:"product_id"`
		Store     string    `db:"store"`
		Start     time.Time `db:"start"`
		End       time.Time `db:"end"`
	}{}
	err = userDbReader.GetContext(ctx, &premiumProduct, `
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
			return nil, fmt.Errorf("error getting premiumProduct for userId %v: %w", userId, err)
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
	err = userDbReader.SelectContext(ctx, &premiumAddons, `
		SELECT
			price_id,
			to_timestamp((uss.payload->>'current_period_start')::bigint) AS start,
			to_timestamp((uss.payload->>'current_period_end')::bigint) AS end,
			COALESCE((uss.payload->>'quantity')::int,1) AS quantity
		FROM users_stripe_subscriptions uss
		INNER JOIN users u ON u.stripe_customer_id = uss.customer_id
		WHERE u.id = $1 AND uss.active = true AND uss.purchase_group = 'addon'`, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting premiumAddons for userId %v: %w", userId, err)
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

	if userInfo.UserGroup == t.UserGroupAdmin {
		userInfo.PremiumPerks = adminPerks
	}

	return userInfo, nil
}

func GetProductSummary(ctx context.Context) (*t.ProductSummary, error) { // TODO @patrick post-beta put into db instead of hardcoding here and make it configurable
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
					ValidatorDashboards:         1,
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
					RewardsChartHistorySeconds: t.ChartHistorySeconds{
						Epoch:  0,
						Hourly: 0,
						Daily:  maxJsInt,
						Weekly: maxJsInt,
					},
					EmailNotificationsPerDay:                       15,
					ConfigureNotificationsViaApi:                   false,
					ValidatorGroupNotifications:                    3,
					WebhookEndpoints:                               3,
					MobileAppCustomThemes:                          true,
					MobileAppWidget:                                true,
					MonitorMachines:                                2,
					MachineMonitoringHistorySeconds:                3600 * 24 * 30,
					NotificationsMachineCustomThreshold:            true,
					NotificationsValidatorDashboardGroupEfficiency: true,
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
					ValidatorDashboards:         2,
					ValidatorsPerDashboard:      300,
					ValidatorGroupsPerDashboard: 10,
					ShareCustomDashboards:       true,
					ManageDashboardViaApi:       false,
					BulkAdding:                  true,
					ChartHistorySeconds: t.ChartHistorySeconds{
						Epoch:  5 * day,
						Hourly: month,
						Daily:  2 * month,
						Weekly: 6 * month,
					},
					RewardsChartHistorySeconds: t.ChartHistorySeconds{
						Epoch:  0,
						Hourly: maxJsInt,
						Daily:  maxJsInt,
						Weekly: maxJsInt,
					},
					EmailNotificationsPerDay:                       20,
					ConfigureNotificationsViaApi:                   false,
					ValidatorGroupNotifications:                    10,
					WebhookEndpoints:                               10,
					MobileAppCustomThemes:                          true,
					MobileAppWidget:                                true,
					MonitorMachines:                                10,
					MachineMonitoringHistorySeconds:                3600 * 24 * 30,
					NotificationsMachineCustomThreshold:            true,
					NotificationsValidatorDashboardGroupEfficiency: true,
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
					ValidatorDashboards:         2,
					ValidatorsPerDashboard:      1000,
					ValidatorGroupsPerDashboard: 30,
					ShareCustomDashboards:       true,
					ManageDashboardViaApi:       true,
					BulkAdding:                  true,
					ChartHistorySeconds: t.ChartHistorySeconds{
						Epoch:  3 * week,
						Hourly: 6 * month,
						Daily:  12 * month,
						Weekly: maxJsInt,
					},
					RewardsChartHistorySeconds: t.ChartHistorySeconds{
						Epoch:  0,
						Hourly: maxJsInt,
						Daily:  maxJsInt,
						Weekly: maxJsInt,
					},
					EmailNotificationsPerDay:                       50,
					ConfigureNotificationsViaApi:                   true,
					ValidatorGroupNotifications:                    60,
					WebhookEndpoints:                               30,
					MobileAppCustomThemes:                          true,
					MobileAppWidget:                                true,
					MonitorMachines:                                10,
					MachineMonitoringHistorySeconds:                3600 * 24 * 30,
					NotificationsMachineCustomThreshold:            true,
					NotificationsValidatorDashboardGroupEfficiency: true,
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

func GetFreeTierPerks(ctx context.Context) (*t.PremiumPerks, error) {
	return &freeTierProduct.PremiumPerks, nil
}
