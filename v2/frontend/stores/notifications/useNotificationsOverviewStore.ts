import { defineStore } from 'pinia'
import { API_PATH } from '~/types/customFetch'
import type {
  InternalGetUserNotificationsResponse, NotificationOverviewData,
} from '~/types/api/notifications'

const notificationsOverviewStore = defineStore('notifications_overview_store', () => {
  const data = ref<NotificationOverviewData | undefined>()
  return { data }
})

export function useNotificationsDashboardOverviewStore() {
  const { fetch } = useCustomFetch()
  const { data: overview } = storeToRefs(notificationsOverviewStore())

  async function refreshOverview() {
    try {
      const res = await fetch<InternalGetUserNotificationsResponse>(
        API_PATH.NOTIFICATIONS_OVERVIEW,
      )
      overview.value = res.data

      return overview.value
    }
    catch (e) {
      overview.value = undefined
      throw e
    }
  }

  return {
    overview,
    refreshOverview,
  }
}
