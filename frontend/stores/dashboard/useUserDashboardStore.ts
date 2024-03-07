import { defineStore } from 'pinia'
import { useCustomFetch } from '~/composables/useCustomFetch'
import type { GetUserDashboardsResponse, UserDashboardsData } from '~/types/api/dashboard'

export const useUserDashboardStore = defineStore('user_dashboards', () => {
  const dashboards = ref<UserDashboardsData | undefined | null>()
  async function getDashboards () {
    const res = await useCustomFetch<GetUserDashboardsResponse>(API_PATH.USER_DASHBOARDS)
    dashboards.value = res.data
    return dashboards.value
  }

  return { dashboards, getDashboards }
})
