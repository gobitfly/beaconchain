import { defineStore } from 'pinia'
import type { GetUserDashboardsResponse, UserDashboardsData } from '~/types/api/dashboard'

export const useUserDashboardStore = defineStore('user_dashboards', () => {
  const { fetch } = useCustomFetch()
  const dashboards = ref<UserDashboardsData | undefined | null>()
  async function getDashboards () {
    const res = await fetch<GetUserDashboardsResponse>(API_PATH.USER_DASHBOARDS)
    dashboards.value = res.data
    return dashboards.value
  }

  return { dashboards, getDashboards }
})
