import { defineStore } from 'pinia'
import type { InternalGetUserNotificationMachinesResponse } from '~/types/api/notifications'
import { API_PATH } from '~/types/customFetch'
import type { TableQueryParams } from '~/types/datatable'

const notificationsMachineStore = defineStore('notifications-network-store', () => {
  const data = ref<InternalGetUserNotificationMachinesResponse | undefined>()
  return { data }
})

export function useNotificationsMachineStore() {
  const { isLoggedIn } = useUserStore()

  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(notificationsMachineStore())
  const {
    cursor, isStoredQuery, onSort, pageSize, pendingQuery, query, setCursor, setPageSize, setSearch, setStoredQuery,
  } = useTableQuery({
    limit: 10, sort: 'timestamp:desc',
  }, 10)
  const isLoading = ref(false)

  async function loadMachineNotifications(q: TableQueryParams) {
    isLoading.value = true
    setStoredQuery(q)
    try {
      const result = await fetch<InternalGetUserNotificationMachinesResponse>(
        API_PATH.NOTIFICATIONS_MACHINE,
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

  const machineNotifications = computed(() => {
    return data.value
  })

  watch(query, (q) => {
    if (q && isLoggedIn.value) {
      loadMachineNotifications(q)
    }
  }, { immediate: true })

  return {
    cursor,
    isLoading,
    machineNotifications,
    onSort,
    pageSize,
    query: pendingQuery,
    setCursor,
    setPageSize,
    setSearch,
  }
}
