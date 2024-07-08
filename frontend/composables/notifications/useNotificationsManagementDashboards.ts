import type { TableQueryParams } from '~/types/datatable'
import { API_PATH } from '~/types/customFetch'
import type { NotificationsManagementDashboardResponse } from '~/types/notifications/management'

export function useNotificationsManagementDashboards () {
  const { fetch } = useCustomFetch()

  const data = ref < NotificationsManagementDashboardResponse>()
  const { query, pendingQuery, cursor, pageSize, onSort, setCursor, setPageSize, setSearch, setStoredQuery, isStoredQuery } = useTableQuery({ limit: 10, sort: 'dashboard_id:desc' }, 10)
  const isLoading = ref(false)

  const dashboards = computed(() => data.value)

  async function getDashboards (q?: TableQueryParams) {
    isLoading.value = true
    setStoredQuery(q)
    const res = await fetch<NotificationsManagementDashboardResponse>(API_PATH.NOTIFICATIONS_MANAGEMENT_DASHBOARD, undefined, undefined, q)

    isLoading.value = false
    if (!isStoredQuery(q)) {
      return // in case some query params change while loading
    }

    data.value = res
    return res
  }

  watch(query, (q) => {
    getDashboards(q)
  }, { immediate: true })

  return { dashboards, query: pendingQuery, cursor, pageSize, isLoading, onSort, setCursor, setPageSize, setSearch }
}
