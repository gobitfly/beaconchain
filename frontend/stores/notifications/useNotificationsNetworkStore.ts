import { defineStore } from 'pinia'
import type { InternalGetUserNotificationNetworksResponse } from '~/types/api/notifications'
import { API_PATH } from '~/types/customFetch'
import type { TableQueryParams } from '~/types/datatable'

const notificationsNetworkStore = defineStore('notifications-network-store', () => {
  const data = ref<InternalGetUserNotificationNetworksResponse | undefined>()
  return { data }
})

export function useNotificationsNetworkStore() {
  const { isLoggedIn } = useUserStore()

  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(notificationsNetworkStore())
  const {
    cursor, isStoredQuery, onSort, pageSize, pendingQuery, query, setCursor, setPageSize, setSearch, setStoredQuery,
  } = useTableQuery({
    limit: 10, sort: 'timestamp:desc',
  }, 10)
  const isLoading = ref(false)

  async function loadNetworkNotifications(q: TableQueryParams) {
    isLoading.value = true
    setStoredQuery(q)
    try {
      const result = await fetch<InternalGetUserNotificationNetworksResponse>(
        API_PATH.NOTIFICATIONS_NETWORK,
        undefined,
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

  const networkNotifications = computed(() => {
    return data.value
  })

  watch(query, (q) => {
    if (q) {
      isLoggedIn.value && loadNetworkNotifications(q)
    }
  }, { immediate: true })

  return {
    cursor,
    isLoading,
    networkNotifications,
    onSort,
    pageSize,
    query: pendingQuery,
    setCursor,
    setPageSize,
    setSearch,
  }
}
