import { defineStore } from 'pinia'
import type { InternalGetUserNotificationClientsResponse } from '~/types/api/notifications'
import { API_PATH } from '~/types/customFetch'
import type { TableQueryParams } from '~/types/datatable'

const notificationsClientStore = defineStore('notifications-clients-store', () => {
  const data = ref<InternalGetUserNotificationClientsResponse | undefined>()
  return { data }
})

export function useNotificationsClientStore() {
  const { isLoggedIn } = useUserStore()

  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(notificationsClientStore())
  const {
    cursor, isStoredQuery, onSort, pageSize, pendingQuery, query, setCursor, setPageSize, setSearch, setStoredQuery,
  } = useTableQuery({
    limit: 10, sort: 'timestamp:desc',
  }, 10)
  const isLoading = ref(false)

  async function loadClientsNotifications(q: TableQueryParams) {
    isLoading.value = true
    setStoredQuery(q)
    try {
      const result = await fetch<InternalGetUserNotificationClientsResponse>(
        API_PATH.NOTIFICATIONS_CLIENTS,
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
    catch {
      data.value = undefined
      isLoading.value = false
    }
    return data.value
  }

  const clientsNotifications = computed(() => {
    return data.value
  })

  watch(query, (q) => {
    if (q && isLoggedIn.value) {
      loadClientsNotifications(q)
    }
  }, { immediate: true })

  return {
    clientsNotifications,
    cursor,
    isLoading,
    onSort,
    pageSize,
    query: pendingQuery,
    setCursor,
    setPageSize,
    setSearch,
  }
}
