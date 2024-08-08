import { defineStore } from 'pinia'
import type { InternalGetUserNotificationDashboardsResponse } from '~/types/api/notifications'
import { API_PATH } from '~/types/customFetch'
import type { TableQueryParams } from '~/types/datatable'

const notificationsDashboardStore = defineStore(
  'notifications-dashboard-store',
  () => {
    const data = ref<InternalGetUserNotificationDashboardsResponse | undefined>()
    return { data }
  },
)

export function useNotificationsDashboardStore(networkId: globalThis.Ref<number>) {
  const { isLoggedIn } = useUserStore()

  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(notificationsDashboardStore())
  const {
    query,
    pendingQuery,
    cursor,
    pageSize,
    onSort,
    setCursor,
    setPageSize,
    setSearch,
    setStoredQuery,
    isStoredQuery,
  } = useTableQuery({ limit: 10, sort: 'timestamp:desc' }, 10)
  const isLoading = ref(false)

  async function loadNotificationsDashboards(q: TableQueryParams) {
    isLoading.value = true
    setStoredQuery(q)
    try {
      const result = await fetch<InternalGetUserNotificationDashboardsResponse>(
        API_PATH.NOTIFICATIONS_DASHBOARDS,
        { query: { network: networkId.value } },
        undefined,
        q,
      )

      isLoading.value = false
      if (!isStoredQuery(q)) {
        return // in case some query params change while loading
      }

      data.value = result
    }
    catch (e) {
      data.value = undefined
      isLoading.value = false
    }
    return data.value
  }

  const notificationsDashboards = computed(() => {
    return data.value
  })

  watch([
    query,
    networkId], ([q]) => {
    if (q) {
      isLoggedIn.value && loadNotificationsDashboards(q)
    }
  },
  { immediate: true },
  )

  return {
    cursor,
    notificationsDashboards,
    isLoading,
    onSort,
    pageSize,
    query: pendingQuery,
    setCursor,
    setPageSize,
    setSearch,
  }
}
