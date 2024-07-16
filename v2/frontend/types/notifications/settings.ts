import type { ApiDataResponse } from '~/types/api/common'
///
// TODO: replace with real api types, once ready
export interface NotificationSettingsPairedDevice {
  id: string;
  name?: string;
  enable_notifications: boolean;
  pairedTs: number;
}

export interface NotificationsManagementSettings {
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
  do_not_disturb_timestamp: number /* uint64 */;
  enable_email: boolean;
  enable_push: boolean;
  paired_devices: NotificationSettingsPairedDevice[];
}

export type NotificationsManagementDashboardResponse = ApiDataResponse<NotificationsManagementSettings>;
// end of temp API stuff

export type NotificationsManagementSettingsGeneralTab = Pick<NotificationsManagementSettings, 'do_not_disturb_timestamp' | 'enable_email' | 'enable_push' | 'paired_devices'>;

export interface NotificationsManagementSettingsProvider{
  generalSettings: ComputedRef<NotificationsManagementSettingsGeneralTab | undefined>,
  updateGeneralSettings: (newSettings: NotificationsManagementSettingsGeneralTab) => void,
  isLoading: Ref<boolean>
}
