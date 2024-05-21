package dataaccess

import (
	"database/sql"
	"fmt"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/pkg/errors"
)

func (d *DataAccessService) GetUser(email string) (*t.User, error) {
	// TODO @recy21
	result := &t.User{}
	err := d.userReader.Get(result, `
		WITH
			latest_and_greatest_sub AS (
				SELECT user_id, product_id FROM users_app_subscriptions 
				left join users on users.id = user_id 
				WHERE users.email = $1 AND active = true
				ORDER BY CASE product_id
					WHEN 'whale' THEN 1
					WHEN 'goldfish' THEN 2
					WHEN 'plankton' THEN 3
					ELSE 4  -- For any other product_id values
				END, users_app_subscriptions.created_at DESC LIMIT 1
			)
		SELECT users.id as id, password, COALESCE(product_id, '') as product_id, COALESCE(user_group, '') AS user_group 
		FROM users
		left join latest_and_greatest_sub on latest_and_greatest_sub.user_id = users.id  
		WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: user with email %s not found", ErrNotFound, email)
	}
	return result, err
}

func (d *DataAccessService) GetUserInfo(id uint64) (*t.UserInfo, error) {
	// TODO patrick
	// return d.dummy.GetUserInfo(id)
	return &t.UserInfo{
		Id:      id,
		Email:   "mail@dummy.com",
		ApiKeys: []string{"dummykey1", "dummykey1"},
		ApiPerks: t.ApiPerks{
			UnitsPerSecond:    10,
			UnitsPerMonth:     10,
			ApiKeys:           4,
			ConsensusLayerAPI: true,
			ExecutionLayerAPI: true,
			Layer2API:         true,
			NoAds:             true,
			DiscordSuport:     false,
		},
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
		Subscriptions: []t.UserSubscription{
			{
				ProductId:       "orca",
				ProductName:     "Orca",
				ProductCategory: t.ProductCategoryPremium,
				Start:           1715768109,
				End:             1718446509,
			},
			{
				ProductId:       "1k_extra_valis_per_dasboard",
				ProductName:     "+1000 Validators per Dasboard",
				ProductCategory: t.ProductCategoryPremiumAddon,
				Start:           1715768109,
				End:             1718446509,
			},
			{
				ProductId:       "10k_extra_valis_per_dasboard",
				ProductName:     "+10,000 Validators per Dasboard",
				ProductCategory: t.ProductCategoryPremiumAddon,
				Start:           1715768109,
				End:             1718446509,
			},
		},
	}, nil
}

func (d *DataAccessService) GetProductSummary() (*t.ProductSummary, error) {
	// TODO patrick
	return &t.ProductSummary{
		ApiProducts: []t.ApiProduct{
			{
				ProductId:        "api_free",
				ProductName:      "Free",
				PricePerMonthEur: 0,
				PricePerYearEur:  0 * 12 * 0.9,
				ApiPerks: t.ApiPerks{
					UnitsPerSecond:    10,
					UnitsPerMonth:     10_000_000,
					ApiKeys:           2,
					ConsensusLayerAPI: true,
					ExecutionLayerAPI: true,
					Layer2API:         true,
					NoAds:             true,
					DiscordSuport:     false,
				},
			},
			{
				ProductId:        "api_iron",
				ProductName:      "Iron",
				PricePerMonthEur: 1.99,
				PricePerYearEur:  1.99 * 12 * 0.9,
				ApiPerks: t.ApiPerks{
					UnitsPerSecond:    20,
					UnitsPerMonth:     20_000_000,
					ApiKeys:           10,
					ConsensusLayerAPI: true,
					ExecutionLayerAPI: true,
					Layer2API:         true,
					NoAds:             true,
					DiscordSuport:     false,
				},
			},
			{
				ProductId:        "api_siler",
				ProductName:      "Silver",
				PricePerMonthEur: 2.99,
				PricePerYearEur:  2.99 * 12 * 0.9,
				ApiPerks: t.ApiPerks{
					UnitsPerSecond:    30,
					UnitsPerMonth:     100_000_000,
					ApiKeys:           20,
					ConsensusLayerAPI: true,
					ExecutionLayerAPI: true,
					Layer2API:         true,
					NoAds:             true,
					DiscordSuport:     false,
				},
			},
			{
				ProductId:        "api_gold",
				ProductName:      "Gold",
				PricePerMonthEur: 3.99,
				PricePerYearEur:  3.99 * 12 * 0.9,
				ApiPerks: t.ApiPerks{
					UnitsPerSecond:    40,
					UnitsPerMonth:     200_000_000,
					ApiKeys:           40,
					ConsensusLayerAPI: true,
					ExecutionLayerAPI: true,
					Layer2API:         true,
					NoAds:             true,
					DiscordSuport:     false,
				},
			},
		},
		PremiumProducts: []t.PremiumProduct{
			{
				ProductId:   "premium_free",
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
				PricePerYearEur:  0 * 12 * 0.9,
			},
			{
				ProductId:   "guppy",
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
					MachineMonitoringHistorySeconds: 3600 * 30,
					CustomMachineAlerts:             true,
				},
				PricePerMonthEur: 9.99,
				PricePerYearEur:  9.99 * 12 * 0.9,
			},
			{
				ProductId:   "dolphin",
				ProductName: "Dolphin",
				PremiumPerks: t.PremiumPerks{
					AdFree:                          true,
					ValidatorDasboards:              1,
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
					MobileAppCustomThemes:           false,
					MobileAppWidget:                 false,
					MonitorMachines:                 10,
					MachineMonitoringHistorySeconds: 3600 * 24 * 30,
					CustomMachineAlerts:             true,
				},
				PricePerMonthEur: 29.99,
				PricePerYearEur:  29.99 * 12 * 0.9,
			},
			{
				ProductId:   "orca",
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
				PricePerMonthEur: 49.99,
				PricePerYearEur:  49.99 * 12 * 0.9,
			},
		},
		ExtraDashboardValidatorsPremiumAddon: []t.ExtraDashboardValidatorsPremiumAddon{
			{
				ProductId:                "1k_extra_valis_per_dasboard",
				ProductName:              "+1000 Validators per Dasboard",
				ExtraDashboardValidators: 1000,
				PricePerMonthEur:         9.99,
				PricePerYearEur:          9.99 * 12 * 0.9,
			},
			{
				ProductId:                "10k_extra_valis_per_dasboard",
				ProductName:              "+10,000 Validators per Dasboard",
				ExtraDashboardValidators: 10000,
				PricePerMonthEur:         15.99,
				PricePerYearEur:          15.99 * 12 * 0.9,
			},
		},
	}, nil
}
