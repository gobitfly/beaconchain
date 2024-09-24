package types

import "github.com/shopspring/decimal"

// ------------------------------------------------------------
// Overview
type NotificationOverviewData struct {
	IsEmailNotificationsEnabled bool `json:"is_email_notifications_enabled"`
	IsPushNotificationsEnabled  bool `json:"is_push_notifications_enabled"`

	// these will list 3 group names
	VDBMostNotifiedGroups [3]string `json:"vdb_most_notified_groups"`
	ADBMostNotifiedGroups [3]string `json:"adb_most_notified_groups"`

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

type InternalGetUserNotificationsResponse ApiDataResponse[NotificationOverviewData]

// ------------------------------------------------------------
// Dashboards Table
type NotificationDashboardsTableRow struct {
	IsAccountDashboard bool     `db:"is_account_dashboard" json:"is_account_dashboard"` // if false it's a validator dashboard
	ChainId            uint64   `db:"chain_id" json:"chain_id"`
	Epoch              uint64   `db:"epoch" json:"epoch"`
	DashboardId        uint64   `db:"dashboard_id" json:"dashboard_id"`
	DashboardName      string   `db:"dashboard_name" json:"dashboard_name"`
	GroupId            uint64   `db:"group_id" json:"group_id"`
	GroupName          string   `db:"group_name" json:"group_name"`
	EntityCount        uint64   `db:"entity_count" json:"entity_count"`
	EventTypes         []string `db:"event_types" json:"event_types" tstype:"('validator_online' | 'validator_offline' | 'group_online' | 'group_offline' | 'attestation_missed' | 'proposal_success' | 'proposal_missed' | 'proposal_upcoming' | 'sync' | 'withdrawal' | 'got_slashed' | 'has_slashed' | 'incoming_tx' | 'outgoing_tx' | 'transfer_erc20' | 'transfer_erc721' | 'transfer_erc1155')[]" faker:"oneof: validator_offline, group_offline, attestation_missed, proposal_success, proposal_missed, proposal_upcoming, sync, withdrawal, slashed_own, incoming_tx, outgoing_tx, transfer_erc20, transfer_erc721, transfer_erc1155"`
}

type InternalGetUserNotificationDashboardsResponse ApiPagingResponse[NotificationDashboardsTableRow]

// ------------------------------------------------------------
// Validator Dashboard Notification Detail

type NotificationEventGroup struct {
	GroupName   string `json:"group_name"`
	DashboardID uint64 `json:"dashboard_id"`
}
type NotificationEventGroupBackOnline struct {
	GroupName   string `json:"group_name"`
	DashboardID uint64 `json:"dashboard_id"`
	EpochCount  uint64 `json:"epoch_count"`
}

type NotificationEventValidatorBackOnline struct {
	Index      uint64 `json:"index"`
	EpochCount uint64 `json:"epoch_count"`
}

type NotificationValidatorDashboardDetail struct {
	ValidatorOffline         []uint64                               `json:"validator_offline"` // validator indices
	GroupOffline             []NotificationEventGroup               `json:"group_offline"`
	ProposalMissed           []IndexBlocks                          `json:"proposal_missed"`
	ProposalDone             []IndexBlocks                          `json:"proposal_done"`
	UpcomingProposals        []uint64                               `json:"upcoming_proposals"` // slot numbers
	Slashed                  []uint64                               `json:"slashed"`            // validator indices
	SyncCommittee            []uint64                               `json:"sync_committee"`     // validator indices
	AttestationMissed        []IndexBlocks                          `json:"attestation_missed"`
	Withdrawal               []IndexBlocks                          `json:"withdrawal"`
	ValidatorOfflineReminder []uint64                               `json:"validator_offline_reminder"` // validator indices
	GroupOfflineReminder     []NotificationEventGroup               `json:"group_offline_reminder"`
	ValidatorBackOnline      []NotificationEventValidatorBackOnline `json:"validator_back_online"`
	GroupBackOnline          []NotificationEventGroupBackOnline     `json:"group_back_online"`
}

type InternalGetUserNotificationsValidatorDashboardResponse ApiDataResponse[NotificationValidatorDashboardDetail]

type NotificationEventExecution struct {
	Address         Address         `json:"address"`
	Amount          decimal.Decimal `json:"amount"`
	TransactionHash Hash            `json:"transaction_hash"`
	TokenName       string          `json:"token_name"` // this field will prob change depending on how execution stuff is implemented
}

type NotificationAccountDashboardDetail struct {
	IncomingTransactions  []NotificationEventExecution `json:"incoming_transactions"`
	OutgoingTransactions  []NotificationEventExecution `json:"outgoing_transactions"`
	ERC20TokenTransfers   []NotificationEventExecution `json:"erc20_token_transfers"`
	ERC721TokenTransfers  []NotificationEventExecution `json:"erc721_token_transfers"`
	ERC1155TokenTransfers []NotificationEventExecution `json:"erc1155_token_transfers"`
}

type InternalGetUserNotificationsAccountDashboardResponse ApiDataResponse[NotificationAccountDashboardDetail]

// ------------------------------------------------------------
// Machines Table
type NotificationMachinesTableRow struct {
	MachineName string  `json:"machine_name"`
	Threshold   float64 `json:"threshold" faker:"boundary_start=0, boundary_end=1"`
	EventType   string  `json:"event_type" tstype:"'offline' | 'storage' | 'cpu' | 'memory'" faker:"oneof: offline, storage, cpu, memory"`
	Timestamp   int64   `json:"timestamp"`
}

type InternalGetUserNotificationMachinesResponse ApiPagingResponse[NotificationMachinesTableRow]

// ------------------------------------------------------------
// Clients Table
type NotificationClientsTableRow struct {
	ClientName string `json:"client_name"`
	Version    string `json:"version"`
	Timestamp  int64  `json:"timestamp"`
}

type InternalGetUserNotificationClientsResponse ApiPagingResponse[NotificationClientsTableRow]

// ------------------------------------------------------------
// Rocket Pool Table
type NotificationRocketPoolTableRow struct {
	Timestamp   int64   `json:"timestamp"`
	EventType   string  `json:"event_type" tstype:"'reward_round' | 'collateral_max' | 'collateral_min'" faker:"oneof: reward_round, collateral_max, collateral_min"`
	AlertValue  float64 `json:"alert_value,omitempty"` // only for some notification types, e.g. max collateral
	NodeAddress Hash    `json:"node_address"`
}

type InternalGetUserNotificationRocketPoolResponse ApiPagingResponse[NotificationRocketPoolTableRow]

// ------------------------------------------------------------
// Networks Table
type NotificationNetworksTableRow struct {
	ChainId    uint64          `json:"chain_id"`
	Timestamp  int64           `json:"timestamp"`
	EventType  string          `json:"event_type" tstype:"'gas_above' | 'gas_below' | 'participation_rate'" faker:"oneof: gas_above, gas_below, participation_rate"`
	AlertValue decimal.Decimal `json:"alert_value"` // wei string for gas alerts, otherwise percentage (0-1) for participation rate
}

type InternalGetUserNotificationNetworksResponse ApiPagingResponse[NotificationNetworksTableRow]

// ------------------------------------------------------------
// Notification Settings
type NotificationSettingsNetwork struct {
	GasAboveThreshold          decimal.Decimal `json:"gas_above_threshold" faker:"boundary_start=0, boundary_end=1"`          // 0 is disabled
	GasBelowThreshold          decimal.Decimal `json:"gas_below_threshold" faker:"boundary_start=0, boundary_end=1"`          // 0 is disabled
	ParticipationRateThreshold float64         `json:"participation_rate_threshold" faker:"boundary_start=0, boundary_end=1"` // 0 is disabled
}
type NotificationNetwork struct {
	ChainId  uint64                      `json:"chain_id"`
	Settings NotificationSettingsNetwork `json:"settings"`
}
type InternalPutUserNotificationSettingsNetworksResponse ApiDataResponse[NotificationNetwork]

type NotificationPairedDevice struct {
	Id                     string `json:"id"`
	PairedTimestamp        int64  `json:"paired_timestamp"`
	Name                   string `json:"name,omitempty"`
	IsNotificationsEnabled bool   `json:"is_notifications_enabled"`
}
type InternalPutUserNotificationSettingsPairedDevicesResponse ApiDataResponse[NotificationPairedDevice]

type NotificationSettingsGeneral struct {
	DoNotDisturbTimestamp       int64 `json:"do_not_disturb_timestamp"` // notifications are disabled until this timestamp
	IsEmailNotificationsEnabled bool  `json:"is_email_notifications_enabled"`
	IsPushNotificationsEnabled  bool  `json:"is_push_notifications_enabled"`

	IsMachineOfflineSubscribed   bool    `json:"is_machine_offline_subscribed"`
	MachineStorageUsageThreshold float64 `json:"machine_storage_usage_threshold" faker:"boundary_start=0, boundary_end=1"` // 0 means disabled
	MachineCpuUsageThreshold     float64 `json:"machine_cpu_usage_threshold" faker:"boundary_start=0, boundary_end=1"`     // 0 means disabled
	MachineMemoryUsageThreshold  float64 `json:"machine_memory_usage_threshold" faker:"boundary_start=0, boundary_end=1"`  // 0 means disabled

	SubscribedClients                    []string `json:"subscribed_clients"`
	IsRocketPoolNewRewardRoundSubscribed bool     `json:"is_rocket_pool_new_reward_round_subscribed"`
	RocketPoolMaxCollateralThreshold     float64  `json:"rocket_pool_max_collateral_threshold" faker:"boundary_start=0, boundary_end=1"` // 0 means disabled
	RocketPoolMinCollateralThreshold     float64  `json:"rocket_pool_min_collateral_threshold" faker:"boundary_start=0, boundary_end=1"` // 0 means disabled
}
type InternalPutUserNotificationSettingsGeneralResponse ApiDataResponse[NotificationSettingsGeneral]
type NotificationSettings struct {
	GeneralSettings NotificationSettingsGeneral `json:"general_settings"`
	Networks        []NotificationNetwork       `json:"networks"`
	PairedDevices   []NotificationPairedDevice  `json:"paired_devices"`
}
type InternalGetUserNotificationSettingsResponse ApiDataResponse[NotificationSettings]

type NotificationSettingsValidatorDashboard struct {
	WebhookUrl              string `json:"webhook_url" faker:"url"`
	IsWebhookDiscordEnabled bool   `json:"is_webhook_discord_enabled"`
	IsRealTimeModeEnabled   bool   `json:"is_real_time_mode_enabled"`

	IsValidatorOfflineSubscribed      bool    `json:"is_validator_offline_subscribed"`
	GroupOfflineThreshold             float64 `json:"group_offline_threshold" faker:"boundary_start=0, boundary_end=1"` // 0 is disabled
	IsAttestationsMissedSubscribed    bool    `json:"is_attestations_missed_subscribed"`
	IsBlockProposalSubscribed         bool    `json:"is_block_proposal_subscribed"`
	IsUpcomingBlockProposalSubscribed bool    `json:"is_upcoming_block_proposal_subscribed"`
	IsSyncSubscribed                  bool    `json:"is_sync_subscribed"`
	IsWithdrawalProcessedSubscribed   bool    `json:"is_withdrawal_processed_subscribed"`
	IsSlashedSubscribed               bool    `json:"is_slashed_subscribed"`
}

type InternalPutUserNotificationSettingsValidatorDashboardResponse ApiDataResponse[NotificationSettingsValidatorDashboard]

type NotificationSettingsAccountDashboard struct {
	WebhookUrl                      string   `json:"webhook_url" faker:"url"`
	IsWebhookDiscordEnabled         bool     `json:"is_webhook_discord_enabled"`
	IsIgnoreSpamTransactionsEnabled bool     `json:"is_ignore_spam_transactions_enabled"`
	SubscribedChainIds              []uint64 `json:"subscribed_chain_ids" faker:"chain_ids"`

	IsIncomingTransactionsSubscribed  bool    `json:"is_incoming_transactions_subscribed"`
	IsOutgoingTransactionsSubscribed  bool    `json:"is_outgoing_transactions_subscribed"`
	IsERC20TokenTransfersSubscribed   bool    `json:"is_erc20_token_transfers_subscribed"`
	ERC20TokenTransfersValueThreshold float64 `json:"erc20_token_transfers_value_threshold" faker:"boundary_start=0, boundary_end=1000000"` // 0 does not disable, is_erc20_token_transfers_subscribed determines if it's enabled
	IsERC721TokenTransfersSubscribed  bool    `json:"is_erc721_token_transfers_subscribed"`
	IsERC1155TokenTransfersSubscribed bool    `json:"is_erc1155_token_transfers_subscribed"`
}
type InternalPutUserNotificationSettingsAccountDashboardResponse ApiDataResponse[NotificationSettingsAccountDashboard]

type NotificationSettingsDashboardsTableRow struct {
	IsAccountDashboard bool   `json:"is_account_dashboard"` // if false it's a validator dashboard
	DashboardId        uint64 `json:"dashboard_id"`
	GroupId            uint64 `json:"group_id"`
	GroupName          string `json:"group_name"`
	// if it's a validator dashboard, SubscribedEvents is NotificationSettingsAccountDashboard, otherwise NotificationSettingsValidatorDashboard
	Settings interface{} `json:"settings" tstype:"NotificationSettingsAccountDashboard | NotificationSettingsValidatorDashboard" faker:"-"`
	ChainIds []uint64    `json:"chain_ids" faker:"chain_ids"`
}

type InternalGetUserNotificationSettingsDashboardsResponse ApiPagingResponse[NotificationSettingsDashboardsTableRow]
