// TODO: replace with real api types, once ready

import type { ApiPagingResponse } from '~/types/api/common'
import type { DashboardType } from '~/types/dashboard'

export interface NotificationsManagementDashboardRow {
  group_id: number /* int64 */;
  dashboard_id: number /* int64 */;
  dashboard_name: number /* int64 */;
  dashboard_type: DashboardType;
  subscriptions: string[];
  webhook: {
    url: string;
    via_discord: boolean;
  };
  networks: number[]
}
export type NotificationsManagementDashboardResponse = ApiPagingResponse<NotificationsManagementDashboardRow>;
