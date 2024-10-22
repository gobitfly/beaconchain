package types

const UserGroupAdmin = "ADMIN"
const UserGroupDev = "DEV"

type UserInfo struct {
	Id            uint64             `json:"id"`
	UserGroup     string             `json:"-"`
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
	ProductStore    ProductStore    `json:"product_store"`
	Start           int64           `json:"start" faker:"unix_time"`
	End             int64           `json:"end" faker:"unix_time"`
}

type InternalGetUserInfoResponse ApiDataResponse[UserInfo]

type EmailUpdate struct {
	Id           uint64 `json:"id"`
	CurrentEmail string `json:"current_email"`
	PendingEmail string `json:"pending_email"`
}

type InternalPostUserEmailResponse ApiDataResponse[EmailUpdate]

type AdConfigurationUpdateData struct {
	JQuerySelector  string `json:"jquery_selector"`
	InsertMode      string `json:"insert_mode"`
	RefreshInterval uint64 `json:"refresh_interval"`
	ForAllUsers     bool   `json:"for_all_users"`
	BannerId        uint64 `json:"banner_id"`
	HtmlContent     string `json:"html_content"`
	Enabled         bool   `json:"enabled"`
}

type AdConfigurationData struct {
	Key string `json:"key"`
	*AdConfigurationUpdateData
}

type ProductCategory string

const ProductCategoryApi ProductCategory = "api"
const ProductCategoryPremium ProductCategory = "premium"
const ProductCategoryPremiumAddon ProductCategory = "premium_addon"

type ProductStore string

const ProductStoreStripe ProductStore = "stripe"
const ProductStoreIosAppstore ProductStore = "ios-appstore"
const ProductStoreAndroidPlaystore ProductStore = "android-playstore"
const ProductStoreEthpool ProductStore = "ethpool"
const ProductStoreCustom ProductStore = "custom"

type ProductSummary struct {
	ValidatorsPerDashboardLimit          uint64                                 `json:"validators_per_dashboard_limit"`
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
	AdFree                                         bool                `json:"ad_free"` // note that this is somhow redunant, since there is already ApiPerks.NoAds
	ValidatorDashboards                            uint64              `json:"validator_dashboards"`
	ValidatorsPerDashboard                         uint64              `json:"validators_per_dashboard"`
	ValidatorGroupsPerDashboard                    uint64              `json:"validator_groups_per_dashboard"`
	ShareCustomDashboards                          bool                `json:"share_custom_dashboards"`
	ManageDashboardViaApi                          bool                `json:"manage_dashboard_via_api"`
	BulkAdding                                     bool                `json:"bulk_adding"`
	ChartHistorySeconds                            ChartHistorySeconds `json:"chart_history_seconds"`
	EmailNotificationsPerDay                       uint64              `json:"email_notifications_per_day"`
	ConfigureNotificationsViaApi                   bool                `json:"configure_notifications_via_api"`
	ValidatorGroupNotifications                    uint64              `json:"validator_group_notifications"`
	WebhookEndpoints                               uint64              `json:"webhook_endpoints"`
	MobileAppCustomThemes                          bool                `json:"mobile_app_custom_themes"`
	MobileAppWidget                                bool                `json:"mobile_app_widget"`
	MonitorMachines                                uint64              `json:"monitor_machines"`
	MachineMonitoringHistorySeconds                uint64              `json:"machine_monitoring_history_seconds"`
	NotificationsMachineCustomThreshold            bool                `json:"notifications_machine_custom_threshold"`
	NotificationsValidatorDashboardRealTimeMode    bool                `json:"notifications_validator_dashboard_real_time_mode"`
	NotificationsValidatorDashboardGroupEfficiency bool                `json:"notifications_validator_dashboard_group_offline"`
}

// TODO @patrick post-beta StripeCreateCheckoutSession and StripeCustomerPortal are currently served from v1 (loadbalanced), Once V1 is not affected by this anymore, consider wrapping this with ApiDataResponse

type StripeCreateCheckoutSession struct {
	SessionId string `json:"sessionId,omitempty"`
	Error     string `json:"error,omitempty"`
}

type StripeCustomerPortal struct {
	Url string `json:"url"`
}

type OAuthAppData struct {
	ID          uint64 `db:"id"`
	Owner       uint64 `db:"owner_id"`
	AppName     string `db:"app_name"`
	RedirectURI string `db:"redirect_uri"`
	Active      bool   `db:"active"`
}
