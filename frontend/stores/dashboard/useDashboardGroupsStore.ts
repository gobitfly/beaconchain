import { defineStore } from 'pinia'
import { useCustomFetch } from '~/composables/useCustomFetch'
import type { GetUserDashboardsResponse, UserDashboardsData } from '~/types/api/dashboard'

export const useDashboardGroupsStore = defineStore('dashboard_groups', () => {
  const groups = ref<UserDashboardsData | undefined | null>()
  async function getGroups () {
    const res = await useCustomFetch<GetUserDashboardsResponse>(API_PATH.DASHBOARD_GROUPS)
    groups.value = res.data
    return groups.value
  }

  return { groups, getGroups }
})
