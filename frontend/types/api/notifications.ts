// Code generated by tygo. DO NOT EDIT.
/* eslint-disable */
import type { ApiDataResponse, ApiPagingResponse, IndexBlocks, Address, Hash } from './common'

//////////
// source: notifications.go

/**
 * ------------------------------------------------------------
 * Overview
 */
export interface NotificationOverviewData {
  is_email_notifications_enabled: boolean;
  is_push_notifications_enabled: boolean;
  /**
   * these will list 3 group names
   */
  vdb_most_notified_groups: string[];
  adb_most_notified_groups: string[];
  last_24h_emails_count: number /* uint64 */; // daily limit should be available in user info
  last_24h_push_count: number /* uint64 */;
  last_24h_webhook_count: number /* uint64 */;
  /**
   * counts are shown in their respective tables
   */
  vdb_subscriptions_count: number /* uint64 */;
  adb_subscriptions_count: number /* uint64 */;
  machines_subscription_count: number /* uint64 */;
  clients_subscription_count: number /* uint64 */;
  networks_subscription_count: number /* uint64 */;
}
export type InternalGetUserNotificationsResponse = ApiDataResponse<NotificationOverviewData>;
/**
 * ------------------------------------------------------------
 * Dashboards Table
 */
export interface NotificationDashboardsTableRow {
  is_account_dashboard: boolean; // if false it's a validator dashboard
  chain_id: number /* uint64 */;
  epoch: number /* uint64 */;
  dashboard_id: number /* uint64 */;
  group_id: number /* uint64 */;
  group_name: string;
  entity_count: number /* uint64 */;
  event_types: ('validator_online' | 'validator_offline' | 'group_online' | 'group_offline' | 'attestation_missed' | 'proposal_success' | 'proposal_missed' | 'proposal_upcoming' | 'max_collateral' | 'min_collateral' | 'sync' | 'withdrawal' | 'validator_got_slashed' | 'validator_has_slashed' | 'incoming_tx' | 'outgoing_tx' | 'transfer_erc20' | 'transfer_erc721' | 'transfer_erc1155')[];
}
export type InternalGetUserNotificationDashboardsResponse = ApiPagingResponse<NotificationDashboardsTableRow>;
export interface NotificationEventGroup {
  group_name: string;
  dashboard_id: number /* uint64 */;
}
export interface NotificationEventGroupBackOnline {
  group_name: string;
  dashboard_id: number /* uint64 */;
  epoch_count: number /* uint64 */;
}
export interface NotificationEventValidatorBackOnline {
  index: number /* uint64 */;
  epoch_count: number /* uint64 */;
}
export interface NotificationValidatorDashboardDetail {
  validator_offline: number /* uint64 */[]; // validator indices
  group_offline: NotificationEventGroup[];
  proposal_missed: IndexBlocks[];
  proposal_done: IndexBlocks[];
  upcoming_proposals: number /* uint64 */[]; // slot numbers
  slashed: number /* uint64 */[]; // validator indices
  sync_committee: number /* uint64 */[]; // validator indices
  attestation_missed: IndexBlocks[];
  withdrawal: IndexBlocks[];
  validator_offline_reminder: number /* uint64 */[]; // validator indices
  group_offline_reminder: NotificationEventGroup[];
  validator_back_online: NotificationEventValidatorBackOnline[];
  group_back_online: NotificationEventGroupBackOnline[];
}
export type InternalGetUserNotificationsValidatorDashboardResponse = ApiDataResponse<NotificationValidatorDashboardDetail>;
export interface NotificationEventExecution {
  address: Address;
  amount: string /* decimal.Decimal */;
  transaction_hash: Hash;
  token_name: string; // this field will prob change depending on how execution stuff is implemented
}
export interface NotificationAccountDashboardDetail {
  incoming_transactions: NotificationEventExecution[];
  outgoing_transactions: NotificationEventExecution[];
  erc20_token_transfers: NotificationEventExecution[];
  erc721_token_transfers: NotificationEventExecution[];
  erc1155_token_transfers: NotificationEventExecution[];
}
export type InternalGetUserNotificationsAccountDashboardResponse = ApiDataResponse<NotificationAccountDashboardDetail>;
/**
 * ------------------------------------------------------------
 * Machines Table
 */
export interface NotificationMachinesTableRow {
  machine_name: string;
  threshold?: number /* float64 */;
  event_type: 'offline' | 'storage' | 'cpu' | 'memory';
  timestamp: number /* int64 */;
}
export type InternalGetUserNotificationMachinesResponse = ApiPagingResponse<NotificationMachinesTableRow>;
/**
 * ------------------------------------------------------------
 * Clients Table
 */
export interface NotificationClientsTableRow {
  client_name: string;
  version: string;
  url: string;
  timestamp: number /* int64 */;
}
export type InternalGetUserNotificationClientsResponse = ApiPagingResponse<NotificationClientsTableRow>;
/**
 * ------------------------------------------------------------
 * Rocket Pool Table
 */
export interface NotificationRocketPoolTableRow {
  timestamp: number /* int64 */;
  event_type: 'reward_round' | 'collateral_max' | 'collateral_min';
  threshold?: number /* float64 */; // only for some notification types, e.g. max collateral
  node: Address;
}
export type InternalGetUserNotificationRocketPoolResponse = ApiPagingResponse<NotificationRocketPoolTableRow>;
/**
 * ------------------------------------------------------------
 * Networks Table
 */
export interface NotificationNetworksTableRow {
  chain_id: number /* uint64 */;
  timestamp: number /* int64 */;
  event_type: 'new_reward_round' | 'gas_above' | 'gas_below' | 'participation_rate';
  threshold?: string /* decimal.Decimal */; // participation rate threshold should also be passed as decimal string
}
export type InternalGetUserNotificationNetworksResponse = ApiPagingResponse<NotificationNetworksTableRow>;
/**
 * ------------------------------------------------------------
 * Notification Settings
 */
export interface NotificationSettingsNetwork {
  is_gas_above_subscribed: boolean;
  gas_above_threshold: string /* decimal.Decimal */;
  is_gas_below_subscribed: boolean;
  gas_below_threshold: string /* decimal.Decimal */;
  is_participation_rate_subscribed: boolean;
  participation_rate_threshold: number /* float64 */;
  is_new_reward_round_subscribed: boolean;
}
export interface NotificationNetwork {
  chain_id: number /* uint64 */;
  settings: NotificationSettingsNetwork;
}
export type InternalPutUserNotificationSettingsNetworksResponse = ApiDataResponse<NotificationNetwork>;
export interface NotificationPairedDevice {
  id: string;
  paired_timestamp: number /* int64 */;
  name?: string;
  is_notifications_enabled: boolean;
}
export type InternalPutUserNotificationSettingsPairedDevicesResponse = ApiDataResponse<NotificationPairedDevice>;
export interface NotificationSettingsClient {
  id: number /* uint64 */;
  name: string;
  category: 'execution_layer' | 'consensus_layer' | 'other';
  is_subscribed: boolean;
}
export type InternalPutUserNotificationSettingsClientResponse = ApiDataResponse<NotificationSettingsClient>;
export interface NotificationSettingsGeneral {
  do_not_disturb_timestamp: number /* int64 */; // notifications are disabled until this timestamp
  is_email_notifications_enabled: boolean;
  is_push_notifications_enabled: boolean;
  is_machine_offline_subscribed: boolean;
  is_machine_storage_usage_subscribed: boolean;
  machine_storage_usage_threshold: number /* float64 */;
  is_machine_cpu_usage_subscribed: boolean;
  machine_cpu_usage_threshold: number /* float64 */;
  is_machine_memory_usage_subscribed: boolean;
  machine_memory_usage_threshold: number /* float64 */;
}
export type InternalPutUserNotificationSettingsGeneralResponse = ApiDataResponse<NotificationSettingsGeneral>;
export interface NotificationSettings {
  general_settings: NotificationSettingsGeneral;
  has_machines: boolean;
  networks: NotificationNetwork[];
  paired_devices: NotificationPairedDevice[];
  clients: NotificationSettingsClient[];
}
export type InternalGetUserNotificationSettingsResponse = ApiDataResponse<NotificationSettings>;
export interface NotificationSettingsValidatorDashboard {
  webhook_url: string;
  is_webhook_discord_enabled: boolean;
  is_real_time_mode_enabled: boolean;
  is_validator_offline_subscribed: boolean;
  is_group_offline_subscribed: boolean;
  group_offline_threshold: number /* float64 */;
  is_attestations_missed_subscribed: boolean;
  is_block_proposal_subscribed: boolean;
  is_upcoming_block_proposal_subscribed: boolean;
  is_sync_subscribed: boolean;
  is_withdrawal_processed_subscribed: boolean;
  is_slashed_subscribed: boolean;
  is_max_collateral_subscribed: boolean;
  max_collateral_threshold: number /* float64 */;
  is_min_collateral_subscribed: boolean;
  min_collateral_threshold: number /* float64 */;
}
export type InternalPutUserNotificationSettingsValidatorDashboardResponse = ApiDataResponse<NotificationSettingsValidatorDashboard>;
export interface NotificationSettingsAccountDashboard {
  webhook_url: string;
  is_webhook_discord_enabled: boolean;
  is_ignore_spam_transactions_enabled: boolean;
  subscribed_chain_ids: number /* uint64 */[];
  is_incoming_transactions_subscribed: boolean;
  is_outgoing_transactions_subscribed: boolean;
  is_erc20_token_transfers_subscribed: boolean;
  erc20_token_transfers_value_threshold: number /* float64 */;
  is_erc721_token_transfers_subscribed: boolean;
  is_erc1155_token_transfers_subscribed: boolean;
}
export type InternalPutUserNotificationSettingsAccountDashboardResponse = ApiDataResponse<NotificationSettingsAccountDashboard>;
export interface NotificationSettingsDashboardsTableRow {
  is_account_dashboard: boolean; // if false it's a validator dashboard
  dashboard_id: number /* uint64 */;
  group_id: number /* uint64 */;
  group_name: string;
  /**
   * if it's a validator dashboard, SubscribedEvents is NotificationSettingsAccountDashboard, otherwise NotificationSettingsValidatorDashboard
   */
  settings: NotificationSettingsAccountDashboard | NotificationSettingsValidatorDashboard;
  chain_ids: number /* uint64 */[];
}
export type InternalGetUserNotificationSettingsDashboardsResponse = ApiPagingResponse<NotificationSettingsDashboardsTableRow>;
