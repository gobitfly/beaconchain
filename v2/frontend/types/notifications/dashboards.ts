import type { DashboardType } from '~/types/dashboard'
import type { ApiPagingResponse } from '~/types/api/common'

// TODO: when backend ready replace with generated types
export type NotifcationDashboardRow = {
  timestamp: number
  dashboardId: number
  dashboardName: string
  dashboardNetwork: number
  group_id: number
  entity: {
    type: DashboardType
    count: number
  }
  nofitication: string[]
}

export type NotifcationDashboardResponse = ApiPagingResponse<NotifcationDashboardRow>
