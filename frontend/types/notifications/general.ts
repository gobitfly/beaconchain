import type { ApiDataResponse } from '~/types/api/common'

///
// TODO: replace with real api types, once ready
export interface NotificationsManagementGeneral {
  enabled_notifications: {
    general_do_not_disturb_expire_ts: number
    general_email: boolean
    general_push: boolean
    paired_devices_count: number
  }
}

export type NotificationsManagementDashboardResponse = ApiDataResponse<NotificationsManagementGeneral>;

export interface NotificationsManagementGeneralProvider{
  settings: globalThis.ComputedRef<NotificationsManagementDashboardResponse | undefined>,
  refresh: () => Promise<NotificationsManagementDashboardResponse>,
  isLoading: Ref<boolean>
}
