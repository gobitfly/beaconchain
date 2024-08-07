import type { TableQueryParams } from '~/types/datatable'
import { API_PATH } from '~/types/customFetch'
import type { InternalGetUserNotificationSettingsDashboardsResponse } from '~/types/api/notifications'

export function useNotificationsManagementDashboards () {
  const { fetch } = useCustomFetch()

  const data = ref < InternalGetUserNotificationSettingsDashboardsResponse>()
  const { query, pendingQuery, cursor, pageSize, onSort, setCursor, setPageSize, setSearch, setStoredQuery, isStoredQuery } = useTableQuery({ limit: 10, sort: 'dashboard_id:desc' }, 10)
  const isLoading = ref(false)

  const dashboardGroups = computed(() => data.value)

  async function getDashboardGroups (q?: TableQueryParams) {
    isLoading.value = true
    setStoredQuery(q)
    const res = await fetch<InternalGetUserNotificationSettingsDashboardsResponse>(API_PATH.GET_NOTIFICATIONS_SETTINGS_DASHBOARD, undefined, undefined, q)

    isLoading.value = false
    if (!isStoredQuery(q)) {
      return // in case some query params change while loading
    }

    data.value = res
    return res
  }

  watch(query, (q) => {
    getDashboardGroups(q)
  }, { immediate: true })

  return { dashboardGroups, query: pendingQuery, cursor, pageSize, isLoading, onSort, setCursor, setPageSize, setSearch }
}
