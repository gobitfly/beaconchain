package types

import "github.com/shopspring/decimal"

type MobileBundleData struct {
	BundleUrl                string `json:"bundle_url,omitempty"`
	HasNativeUpdateAvailable bool   `json:"has_native_update_available"`
}

type GetMobileLatestBundleResponse ApiDataResponse[MobileBundleData]

type MobileWidgetData struct {
	ValidatorStateCounts ValidatorStateCounts `json:"validator_state_counts"`
	Last24hIncome        decimal.Decimal      `json:"last_24h_income" faker:"eth"`
	Last7dIncome         decimal.Decimal      `json:"last_7d_income" faker:"eth"`
	Last30dApr           float64              `json:"last_30d_apr"`
	Last30dEfficiency    float64              `json:"last_30d_efficiency"`
	NetworkEfficiency    float64              `json:"network_efficiency"`
	RplPrice             decimal.Decimal      `json:"rpl_price" faker:"eth"`
	RplApr               float64              `json:"rpl_apr"`
}

type InternalGetValidatorDashboardMobileWidgetResponse ApiDataResponse[MobileWidgetData]
