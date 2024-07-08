import { API_PATH } from '~/types/customFetch'
import type { NotificationsManagementDashboardResponse, NotificationsManagementGeneralProvider } from '~/types/notifications/general'

export function useUseNotificationsManagementGeneralProvider () {
  const { fetch } = useCustomFetch()

  const data = ref < NotificationsManagementDashboardResponse>()
  const generalSettings = computed(() => data.value)
  const isLoading = ref(false)

  async function refreshGeneralSettings () {
    isLoading.value = true
    const res = await fetch<NotificationsManagementDashboardResponse>(API_PATH.NOTIFICATIONS_MANAGEMENT_GENERAL)

    isLoading.value = false

    data.value = res
    return res
  }

  provide<NotificationsManagementGeneralProvider>('notificationsManagementGeneral', { settings: generalSettings, refresh: refreshGeneralSettings, isLoading })

  return { refreshGeneralSettings }
}
