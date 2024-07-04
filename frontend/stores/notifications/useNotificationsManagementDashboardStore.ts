import { defineStore } from 'pinia'
import type { TableQueryParams } from '~/types/datatable'
import { API_PATH } from '~/types/customFetch'
import type { NotificationsManagementDashboardResponse } from '~/types/notifications/management'

const notificationsManagementDashboardStore = defineStore('notifications_management_dashboard_store', () => {
  const data = ref < NotificationsManagementDashboardResponse>()
  const query = ref < TableQueryParams>()

  return { data, query }
})

export function useNotificationsManagementDashboardStore () {
  const { fetch } = useCustomFetch()
  const { data, query: storedQuery } = storeToRefs(notificationsManagementDashboardStore())
  const isLoading = ref(false)

  const dashboards = computed(() => data.value)
  const query = computed(() => storedQuery.value)

  async function getDashboards (query?: TableQueryParams) {
    isLoading.value = true
    storedQuery.value = query
    const res = await fetch<NotificationsManagementDashboardResponse>(API_PATH.NOTIFICATIONS_MANAGEMENT_DASHBOARD, undefined, undefined, query)

    isLoading.value = false
    if (JSON.stringify(storedQuery.value) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }

    data.value = res
    return res
  }

  return { dashboards, query, isLoading, getDashboards }
}
