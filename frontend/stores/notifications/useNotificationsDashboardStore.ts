import { defineStore } from 'pinia'
import type { InternalGetUserNotificationDashboardsResponse } from '~/types/api/notifications'
import { API_PATH } from '~/types/customFetch'
import type { TableQueryParams } from '~/types/datatable'
import type { ChainIDs } from '~/types/network'

const notificationsDashboardStore = defineStore(
  'notifications-dashboard-store',
  () => {
    const data = ref<InternalGetUserNotificationDashboardsResponse | undefined>()
    return { data }
  },
)

export function useNotificationsDashboardStore(networkId: globalThis.Ref<ChainIDs>) {
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
  } = useTableQuery({
    limit: 10,
    sort: 'epoch:desc',
  }, 10)
  const isLoading = ref(false)

  async function loadNotificationsDashboards(query: TableQueryParams) {
    isLoading.value = true
    setStoredQuery(query)
    try {
      const result = await fetch<InternalGetUserNotificationDashboardsResponse>(
        API_PATH.NOTIFICATIONS_DASHBOARDS,
        { query: { networks: networkId.value } },
      )

      isLoading.value = false
      if (!isStoredQuery(query)) {
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
    networkId,
  ], ([ query ]) => {
    if (query) {
      isLoggedIn.value && loadNotificationsDashboards(query)
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
