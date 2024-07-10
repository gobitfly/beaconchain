package types

import decimal "github.com/jackc/pgx-shopspring-decimal"

// ------------------------------------------------------------
// Overview
type NotificationsOverviewData struct {
	EmailNotificationsEnabled bool `json:"email_notifications_enabled"`
	PushNotificationsEnabled  bool `json:"push_notifications_enabled"`

	// these will list 3 group names
	VDBMostNotifiedGroups []string `json:"vdb_most_notified_groups"`
	ADBMostNotifiedGroups []string `json:"adb_most_notified_groups"`

	Last24hEmailsCount  uint64 `json:"last_24h_emails_count"` // daily limit should be available in user info
	Last24hPushCount    uint64 `json:"last_24h_push_count"`
	Last24hWebhookCount uint64 `json:"last_24h_webhook_count"`

	// counts are shown in their respective tables
	VDBSubscriptionsCount       uint64 `json:"vdb_subscriptions_count"`
	ADBSubscriptionsCount       uint64 `json:"adb_subscriptions_count"`
	MachinesSubscriptionCount   uint64 `json:"machines_subscription_count"`
	ClientsSubscriptionCount    uint64 `json:"clients_subscription_count"`
	RocketPoolSubscriptionCount uint64 `json:"rocket_pool_subscription_count"`
	NetworksSubscriptionCount   uint64 `json:"networks_subscription_count"`
}

type InternalGetNotificationsResponse ApiDataResponse[NotificationsOverviewData]

// ------------------------------------------------------------
// Dashboards Table
type NotificationsDashboardsTableRow struct {
	IsAccountDashboard bool     `json:"is_account_dashboard"` // if false it's a validator dashboard
	ChainId            uint64   `json:"chain_id"`
	Timestamp          uint64   `json:"timestamp"`
	DashboardId        uint64   `json:"dashboard_id"`
	GroupName          string   `json:"group_name"`
	NotificationId     uint64   `json:"notification_id"` // may be string? db schema is not defined afaik
	EntityCount        uint64   `json:"entity_count"`
	EventTypes         []string `json:"event_types"`
}

type InternalGetNotificationDashboards ApiPagingResponse[NotificationsDashboardsTableRow]

// ------------------------------------------------------------
// Machines Table
type NotificationsMachinesTableRow struct {
	MachineName string  `json:"machine_name"`
	Threshold   float64 `json:"threshold"`
	EventType   string  `json:"event_type"`
	Timestamp   uint64  `json:"timestamp"`
}

type InternalGetNotificationMachines ApiPagingResponse[NotificationsMachinesTableRow]

// ------------------------------------------------------------
// Clients Table
type NotificationsClientsTableRow struct {
	ClientName string `json:"client_name"`
	Version    string `json:"version"`
	Timestamp  uint64 `json:"timestamp"`
}

type InternalGetNotificationClients ApiPagingResponse[NotificationsClientsTableRow]

// ------------------------------------------------------------
// Rocket Pool Table
type NotificationRocketPoolTableRow struct {
	Timestamp   uint64  `json:"timestamp"`
	EventType   string  `json:"event_type"`
	AlertValue  float64 `json:"alert_value,omitempty"` // only for some notification types, e.g. max collateral
	NodeAddress Hash    `json:"node_address"`
}

type InternalGetNotificationRocketPool ApiPagingResponse[NotificationRocketPoolTableRow]

// ------------------------------------------------------------
// Networks Table
type NotificationNetworksTableRow struct {
	ChainId    uint64          `json:"chain_id"`
	Timestamp  uint64          `json:"timestamp"`
	EventType  string          `json:"event_type"`
	AlertValue decimal.Decimal `json:"alert_value"` // wei string for gas alerts, otherwise percentage (0-1) for participation rate
}

type InternalGetNotificationNetworks ApiPagingResponse[NotificationNetworksTableRow]

// ------------------------------------------------------------
// Notification Settings
type NotificationEventsNetwork struct {
	GasAbove          decimal.Decimal `json:"gas_above"`
	GasBelow          decimal.Decimal `json:"gas_below"`
	ParticipationRate float64         `json:"participation_rate"`
}
type NotificationSettingsNetwork struct {
	ChainId          uint64                    `json:"chain_id"`
	SubscribedEvents NotificationEventsNetwork `json:"subscribed_events"`
}
type InternalPutNotificationSettingsNetworksResponse ApiDataResponse[NotificationSettingsNetwork]

type NotificationSettingsPairedDevice struct {
	Id                  string `json:"id"`
	Name                string `json:"name,omitempty"`
	EnableNotifications bool   `json:"enable_notifications"`
}
type InternalPutNotificationSettingsPairedDevicesResponse ApiDataResponse[NotificationSettingsNetwork]

type NotificationEventsGeneral struct {
	MachinesOffline            bool     `json:"machines_offline"`
	MachinesStorageUsage       float64  `json:"machines_storage_usage"`
	MachinesCpuUsage           float64  `json:"machines_cpu_usage"`
	MachinesMemoryUsage        float64  `json:"machines_memory_usage"`
	Clients                    []string `json:"clients"`
	ClientsRocketPoolSmartNode bool     `json:"clients_rocket_pool_smart_node"`
	ClientsMevBoost            bool     `json:"clients_mev_boost"`
	RocketPoolNewRewardRound   bool     `json:"rocket_pool_new_reward_round"`
	RocketPoolMaxCollateral    float64  `json:"rocket_pool_max_collateral"`
	RocketPoolMinCollateral    float64  `json:"rocket_pool_min_collateral"`
}
type NotificationSettings struct {
	DoNotDisturbTimestamp uint64                             `json:"do_not_disturb_timestamp"`
	EnableEmail           bool                               `json:"enable_email"`
	EnablePush            bool                               `json:"enable_push"`
	SubscribedEvents      NotificationEventsGeneral          `json:"subscribed_events"`
	Networks              []NotificationSettingsNetwork      `json:"networks"`
	PairedDevices         []NotificationSettingsPairedDevice `json:"paired_devices"`
}
type InternalGetNotificationSettingsResponse ApiDataResponse[NotificationSettings]

type NotificationEventsValidatorDashboard struct {
	ValidatorOffline      bool    `json:"validator_offline"`
	GroupOffline          float64 `json:"group_offline"`
	AttestationsMissed    bool    `json:"attestations_missed"`
	BlockProposal         bool    `json:"block_proposal"`
	UpcomingBlockProposal bool    `json:"upcoming_block_proposal"`
	Sync                  bool    `json:"sync"`
	WithdrawalProcessed   bool    `json:"withdrawal_processed"`
	Slashed               bool    `json:"slashed"`
	RealtimeMode          bool    `json:"realtime_mode"`
}
type NotificationSettingsValidatorDashboard struct {
	WebhookUrl       string                               `json:"webhook_url"`
	WebhookDiscord   bool                                 `json:"webhook_discord"`
	RealTimeMode     bool                                 `json:"real_time_mode"`
	SubscribedEvents NotificationEventsValidatorDashboard `json:"subscribed_events"`
}

type InternalPutNotificationSettingsValidatorDashboardResponse ApiDataResponse[NotificationSettingsValidatorDashboard]

type NotificationEventsAccountDashboard struct {
	IncomingTransactions       bool `json:"incoming_transactions"`
	OutgoingTransactions       bool `json:"outgoing_transactions"`
	TrackERC20TokenTransfers   bool `json:"track_erc20_token_transfers"`
	TrackERC721TokenTransfers  bool `json:"track_erc721_token_transfers"`
	TrackERC1155TokenTransfers bool `json:"track_erc1155_token_transfers"`
}

type NotificationSettingsAccountDashboard struct {
	WebhookUrl             string                             `json:"webhook_url"`
	WebhookDiscord         bool                               `json:"webhook_discord"`
	Networks               []uint64                           `json:"networks"`
	IgnoreSpamTransactions bool                               `json:"ignore_spam_transactions"`
	SubscribedEvents       NotificationEventsAccountDashboard `json:"subscribed_events"`
}
type InternalPutNotificationSettingsAccountDashboardResponse ApiDataResponse[NotificationSettingsAccountDashboard]

type NotificationSettingsDashboardsTableRow struct {
	IsAccountDashboard bool   `json:"is_account_dashboard"` // if false it's a validator dashboard
	DashboardId        uint64 `json:"dashboard_id"`
	GroupName          string `json:"group_name"`
	// if it's a validator dashboard, SubscribedEvents is NotificationEventsValidatorDashboard, otherwise NotificationEventsAccountDashboard
	SubscribedEvents interface{} `json:"subscribed_events" tstype:"NotificationEventsValidatorDashboard | NotificationEventsAccountDashboard"`
	ChainIds         []uint64    `json:"chain_ids"`
}

type InternalGetNotificationSettingsDashboardsResponse ApiDataResponse[NotificationSettingsDashboardsTableRow]
