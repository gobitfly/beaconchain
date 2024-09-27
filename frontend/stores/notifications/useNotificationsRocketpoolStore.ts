import { defineStore } from 'pinia'
import { API_PATH } from '~/types/customFetch'
import type { InternalGetUserNotificationRocketPoolResponse } from '~/types/api/notifications'
import type { TableQueryParams } from '~/types/datatable'

const notificationsRocketpoolStore = defineStore('notifications-rocket-pool-store', () => {
  const data = ref()
  return { data }
})

export function useNotificationsRocketpoolStore() {
  const { isLoggedIn } = useUserStore()
  const { fetch } = useCustomFetch()
  const { data: rocketpoolNotifications } = storeToRefs(notificationsRocketpoolStore())

  const {
    cursor,
    isStoredQuery,
    onSort,
    pageSize,
    query,
    setCursor,
    setPageSize,
    setSearch,
    setStoredQuery,
  } = useTableQuery({
    limit: 10,
    sort: 'timestamp:desc',
  }, 10)

  const isLoading = ref(false)
  async function loadRocketpoolNotifications(q: TableQueryParams) {
    isLoading.value = true
    setStoredQuery(q)
    try {
      const res = await fetch<InternalGetUserNotificationRocketPoolResponse>(
        API_PATH.NOTIFICATIONS_ROCKETPOOL,
        undefined,
        undefined,
        q,
      )
      isLoading.value = false
      if (!isStoredQuery(q)) {
        return // in case some query params change while loading
      }
      rocketpoolNotifications.value = res
      return rocketpoolNotifications.value
    }
    catch (e) {
      rocketpoolNotifications.value = undefined
      throw e
    }
  }
  //

  watch([ query ], ([ q ]) => {
    if (q) {
      isLoggedIn.value && loadRocketpoolNotifications(q)
    }
  },
  { immediate: true },
  )

  return {
    cursor,
    isLoading,
    onSort,
    pageSize,
    query,
    rocketpoolNotifications,
    setCursor,
    setPageSize,
    setSearch,
  }
}
