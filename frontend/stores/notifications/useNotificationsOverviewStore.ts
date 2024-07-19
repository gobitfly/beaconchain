import { defineStore } from 'pinia'
import { ref } from 'vue'
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
      // TODO: Add correct API path
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

  return {
    data,
    error,
    isLoading,
    fetchNotificationsOverview
  }
})
