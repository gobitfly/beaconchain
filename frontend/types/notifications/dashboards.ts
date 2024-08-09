import type { DashboardType } from '~/types/dashboard'
import type { ApiPagingResponse } from '~/types/api/common'

// TODO: when backend ready replace with generated types
export type NotifcationDashboardRow = {
  dashboardId: number,
  dashboardName: string,
  dashboardNetwork: number,
  entity: {
    count: number,
    type: DashboardType,
  },
  group_id: number,
  nofitication: string[],
  timestamp: number,
}

export type NotifcationDashboardResponse =
  ApiPagingResponse<NotifcationDashboardRow>
