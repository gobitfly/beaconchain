// Code generated by tygo. DO NOT EDIT.
/* eslint-disable */
import type { ApiDataResponse, ApiPagingResponse, Hash } from './common'

//////////
// source: notifications.go

/**
 * ------------------------------------------------------------
 * Overview
 */
export interface NotificationsOverviewData {
  email_notifications_enabled: boolean;
  push_notifications_enabled: boolean;
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
  rocket_pool_subscription_count: number /* uint64 */;
  networks_subscription_count: number /* uint64 */;
}
export type InternalGetNotificationsResponse = ApiDataResponse<NotificationsOverviewData>;
/**
 * ------------------------------------------------------------
 * Dashboards Table
 */
export interface NotificationsDashboardsTableRow {
  is_account_dashboard: boolean; // if false it's a validator dashboard
  chain_id: number /* uint64 */;
  timestamp: number /* uint64 */;
  dashboard_id: number /* uint64 */;
  group_name: string;
  notification_id: number /* uint64 */; // may be string? db schema is not defined afaik
  entity_count: number /* uint64 */;
  event_types: string[];
}
export type InternalGetNotificationDashboards = ApiPagingResponse<NotificationsDashboardsTableRow>;
/**
 * ------------------------------------------------------------
 * Machines Table
 */
export interface NotificationsMachinesTableRow {
  machine_name: string;
  threshold: number /* float64 */;
  event_name: string;
  timestamp: number /* uint64 */;
}
export type InternalGetNotificationMachines = ApiPagingResponse<NotificationsMachinesTableRow>;
/**
 * ------------------------------------------------------------
 * Clients Table
 */
export interface NotificationsClientsTableRow {
  client_name: string;
  version: string;
  timestamp: number /* uint64 */;
}
export type InternalGetNotificationClients = ApiPagingResponse<NotificationsClientsTableRow>;
/**
 * ------------------------------------------------------------
 * Rocket Pool Table
 */
export interface NotificationRocketPoolTableRow {
  timestamp: number /* uint64 */;
  event_type: string;
  alert_value?: number /* float64 */; // only for some notification types, e.g. max collateral
  node_address: Hash;
}
export type InternalGetNotificationRocketPool = ApiPagingResponse<NotificationRocketPoolTableRow>;
/**
 * ------------------------------------------------------------
 * Networks Table
 */
export interface NotificationNetworksTableRow {
  chain_id: number /* uint64 */;
  timestamp: number /* uint64 */;
  event_type: string;
  alert_value: string /* decimal.Decimal */; // wei string for gas alerts, otherwise percentage (0-1) for participation rate
}
export type InternalGetNotificationNetworks = ApiPagingResponse<NotificationNetworksTableRow>;
/**
 * ------------------------------------------------------------
 * Notification Settings
 */
export interface NotificationEventsNetwork {
  gas_above: string /* decimal.Decimal */;
  gas_below: string /* decimal.Decimal */;
  participation_rate: number /* float64 */;
}
export interface NotificationSettingsNetwork {
  chain_id: number /* uint64 */;
  subscribed_events: NotificationEventsNetwork;
}
export type InternalPutNotificationSettingsNetworksResponse = ApiDataResponse<NotificationSettingsNetwork>;
export interface NotificationSettingsPairedDevice {
  id: string;
  name?: string;
  enable_notifications: boolean;
}
export type InternalPutNotificationSettingsPairedDevicesResponse = ApiDataResponse<NotificationSettingsNetwork>;
export interface NotificationEventsGeneral {
  machines_offline: boolean;
  machines_storage_usage: number /* float64 */;
  machines_cpu_usage: number /* float64 */;
  machines_memory_usage: number /* float64 */;
  clients: string[];
  clients_rocket_pool_smart_node: boolean;
  clients_mev_boost: boolean;
  rocket_pool_new_reward_round: boolean;
  rocket_pool_max_collateral: number /* float64 */;
  rocket_pool_min_collateral: number /* float64 */;
}
export interface NotificationSettings {
  do_not_disturb_timestamp: number /* uint64 */;
  enable_email: boolean;
  enable_push: boolean;
  subscribed_events: NotificationEventsGeneral;
  networks: NotificationSettingsNetwork[];
  paired_devices: NotificationSettingsPairedDevice[];
}
export type InternalGetNotificationSettingsResponse = ApiDataResponse<NotificationSettings>;
export interface NotificationEventsValidatorDashboard {
  validator_offline: boolean;
  group_offline: number /* float64 */;
  attestations_missed: boolean;
  block_proposal: boolean;
  upcoming_block_proposal: boolean;
  sync: boolean;
  withdrawal_processed: boolean;
  slashed: boolean;
  realtime_mode: boolean;
}
export interface NotificationSettingsValidatorDashboard {
  webhook_url: string;
  webhook_discord: boolean;
  real_time_mode: boolean;
  subscribed_events: NotificationEventsValidatorDashboard;
}
export type InternalPutNotificationSettingsValidatorDashboardResponse = ApiDataResponse<NotificationSettingsValidatorDashboard>;
export interface NotificationEventsAccountDashboard {
  incoming_transactions: boolean;
  outgoing_transactions: boolean;
  track_erc20_token_transfers: boolean;
  track_erc721_token_transfers: boolean;
  track_erc1155_token_transfers: boolean;
}
export interface NotificationSettingsAccountDashboard {
  webhook_url: string;
  webhook_discord: boolean;
  networks: number /* uint64 */[];
  ignore_spam_transactions: boolean;
  subscribed_events: NotificationEventsAccountDashboard;
}
export type InternalPutNotificationSettingsAccountDashboardResponse = ApiDataResponse<NotificationSettingsAccountDashboard>;
export interface NotificationSettingsDashboardsTableRow {
  is_account_dashboard: boolean; // if false it's a validator dashboard
  dashboard_id: number /* uint64 */;
  group_name: string;
  /**
   * if it's a validator dashboard, SubscribedEvents is NotificationEventsValidatorDashboard, otherwise NotificationEventsAccountDashboard
   */
  subscribed_events: NotificationEventsValidatorDashboard | NotificationEventsAccountDashboard;
  chain_ids: number /* uint64 */[];
}
export type InternalGetNotificationSettingsDashboardsResponse = ApiDataResponse<NotificationSettingsDashboardsTableRow>;
