import { defineStore } from 'pinia'
import { API_PATH } from '~/types/customFetch'
import type { TableQueryParams } from '~/types/datatable'
import type { NotifcationDashboardResponse } from '~/types/notifications/dashboards'

const notificationsDashboardStore = defineStore(
  'notifications-dashboard-store',
  () => {
    const data = ref<NotifcationDashboardResponse | undefined>()
    return { data }
  },
)

export function useNotificationsDashboardStore() {
  const { isLoggedIn } = useUserStore()

  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(notificationsDashboardStore())
  const {
    cursor,
    isStoredQuery,
    onSort,
    pageSize,
    pendingQuery,
    query,
    setCursor,
    setPageSize,
    setSearch,
    setStoredQuery,
  } = useTableQuery({ limit: 10, sort: 'dashboard:desc' }, 10)
  const isLoading = ref(false)

  async function loadNotificationsDashboards(q: TableQueryParams) {
    isLoading.value = true
    setStoredQuery(q)
    const result = await fetch<NotifcationDashboardResponse>(
      API_PATH.NOTIFICATIONS_DASHBOARDS,
      undefined,
      undefined,
      q,
    )

    isLoading.value = false
    if (!isStoredQuery(q)) {
      return // in case some query params change while loading
    }

    data.value = result
    return data.value
  }

  const notificationsDashboards = computed(() => {
    return data.value
  })

  watch(
    query,
    (q) => {
      if (q) {
        isLoggedIn.value && loadNotificationsDashboards(q)
      }
    },
    { immediate: true },
  )

  return {
    cursor,
    isLoading,
    notificationsDashboards,
    onSort,
    pageSize,
    query: pendingQuery,
    setCursor,
    setPageSize,
    setSearch,
  }
}
