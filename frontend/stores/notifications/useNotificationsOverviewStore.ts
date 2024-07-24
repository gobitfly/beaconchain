import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { useCustomFetch } from '~/composables/useCustomFetch'
import type { NotificationsOverview } from '~/types/notifications/overview'
import { API_PATH } from '~/types/customFetch'

export const useNotificationsStore = defineStore('notifications-store', () => {
  const data = ref<NotificationsOverview | null>(null)
  const error = ref<Error | null>(null)
  const isLoading = ref<boolean>(false)

  const fetchNotificationsOverview = async () => {
    isLoading.value = true
    try {
      const { fetch } = useCustomFetch()
      const response = await fetch<NotificationsOverview>(API_PATH.NOTIFICATIONS_OVERVIEW)
      data.value = response
      error.value = null
    } catch (err) {
      error.value = err as Error
      data.value = null
    } finally {
      isLoading.value = false
    }
    console.log(data)
  }

  // Watcher to react to changes in data
  watch(data, (newValue, oldValue) => {
    console.log('Notifications overview data changed from', oldValue, 'to', newValue)
    // Any additional logic to handle the change
  })

  return {
    data,
    error,
    isLoading,
    fetchNotificationsOverview
  }
})

// Function to use the store and automatically fetch data when needed
export function useNotificationsOverview() {
  const notificationsStore = useNotificationsStore()
  const { data, fetchNotificationsOverview } = notificationsStore

  // Automatically fetch data when the store is used
  fetchNotificationsOverview()

  return notificationsStore
}
