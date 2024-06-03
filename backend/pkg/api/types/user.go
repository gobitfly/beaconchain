package types

type UserInfo struct {
	Id            uint64             `json:"id"`
	Email         string             `json:"email"`
	ApiKeys       []string           `json:"api_keys"`
	ApiPerks      ApiPerks           `json:"api_perks"`
	PremiumPerks  PremiumPerks       `json:"premium_perks"`
	Subscriptions []UserSubscription `json:"subscriptions"`
}

type UserSubscription struct {
	ProductId       string          `json:"product_id"`
	ProductName     string          `json:"product_name"`
	ProductCategory ProductCategory `json:"product_category"`
	Start           int64           `json:"start" faker:"unix_time"`
	End             int64           `json:"end" faker:"unix_time"`
}

type InternalGetUserInfoResponse ApiDataResponse[UserInfo]

type ProductCategory string

const ProductCategoryApi ProductCategory = "api"
const ProductCategoryPremium ProductCategory = "premium"
const ProductCategoryPremiumAddon ProductCategory = "premium_addon"

type ProductSummary struct {
	StripePublicKey                      string                                 `json:"stripe_public_key"`
	ApiProducts                          []ApiProduct                           `json:"api_products"`
	PremiumProducts                      []PremiumProduct                       `json:"premium_products"`
	ExtraDashboardValidatorsPremiumAddon []ExtraDashboardValidatorsPremiumAddon `json:"extra_dashboard_validators_premium_addons"`
}

type InternalGetProductSummaryResponse ApiDataResponse[ProductSummary]

type ApiProduct struct {
	ProductId            string   `json:"product_id"`
	ProductName          string   `json:"product_name"`
	ApiPerks             ApiPerks `json:"api_perks"`
	PricePerYearEur      float64  `json:"price_per_year_eur"`
	PricePerMonthEur     float64  `json:"price_per_month_eur"`
	IsPopular            bool     `json:"is_popular"`
	StripePriceIdMonthly string   `json:"stripe_price_id_monthly"`
	StripePriceIdYearly  string   `json:"stripe_price_id_yearly"`
}

type ApiPerks struct {
	UnitsPerSecond    uint64 `json:"units_per_second"`
	UnitsPerMonth     uint64 `json:"units_per_month"`
	ApiKeys           uint64 `json:"api_keys"`
	ConsensusLayerAPI bool   `json:"consensus_layer_api"`
	ExecutionLayerAPI bool   `json:"execution_layer_api"`
	Layer2API         bool   `json:"layer2_api"`
	NoAds             bool   `json:"no_ads"` // note that this is somhow redunant, since there is already PremiumPerks.AdFree
	DiscordSupport    bool   `json:"discord_support"`
}

type PremiumProduct struct {
	ProductName          string       `json:"product_name"`
	PremiumPerks         PremiumPerks `json:"premium_perks"`
	PricePerYearEur      float64      `json:"price_per_year_eur"`
	PricePerMonthEur     float64      `json:"price_per_month_eur"`
	IsPopular            bool         `json:"is_popular"`
	ProductIdMonthly     string       `json:"product_id_monthly"`
	ProductIdYearly      string       `json:"product_id_yearly"`
	StripePriceIdMonthly string       `json:"stripe_price_id_monthly"`
	StripePriceIdYearly  string       `json:"stripe_price_id_yearly"`
}

type ExtraDashboardValidatorsPremiumAddon struct {
	ProductName              string  `json:"product_name"`
	ExtraDashboardValidators uint64  `json:"extra_dashboard_validators"`
	PricePerYearEur          float64 `json:"price_per_year_eur"`
	PricePerMonthEur         float64 `json:"price_per_month_eur"`
	ProductIdMonthly         string  `json:"product_id_monthly"`
	ProductIdYearly          string  `json:"product_id_yearly"`
	StripePriceIdMonthly     string  `json:"stripe_price_id_monthly"`
	StripePriceIdYearly      string  `json:"stripe_price_id_yearly"`
}

type PremiumPerks struct {
	AdFree                          bool   `json:"ad_free"` // note that this is somhow redunant, since there is already ApiPerks.NoAds
	ValidatorDasboards              uint64 `json:"validator_dashboards"`
	ValidatorsPerDashboard          uint64 `json:"validators_per_dashboard"`
	ValidatorGroupsPerDashboard     uint64 `json:"validator_groups_per_dashboard"`
	ShareCustomDashboards           bool   `json:"share_custom_dashboards"`
	ManageDashboardViaApi           bool   `json:"manage_dashboard_via_api"`
	HeatmapHistorySeconds           uint64 `json:"heatmap_history_seconds"`
	SummaryChartHistorySeconds      uint64 `json:"summary_chart_history_seconds"`
	EmailNotificationsPerDay        uint64 `json:"email_notifications_per_day"`
	ConfigureNotificationsViaApi    bool   `json:"configure_notifications_via_api"`
	ValidatorGroupNotifications     uint64 `json:"validator_group_notifications"`
	WebhookEndpoints                uint64 `json:"webhook_endpoints"`
	MobileAppCustomThemes           bool   `json:"mobile_app_custom_themes"`
	MobileAppWidget                 bool   `json:"mobile_app_widget"`
	MonitorMachines                 uint64 `json:"monitor_machines"`
	MachineMonitoringHistorySeconds uint64 `json:"machine_monitoring_history_seconds"`
	CustomMachineAlerts             bool   `json:"custom_machine_alerts"`
}
