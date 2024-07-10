import { warn } from 'vue'
import { API_PATH } from '~/types/customFetch'
import type { NotificationsManagementDashboardResponse, NotificationsManagementSettings, NotificationsManagementSettingsGeneralTab, NotificationsManagementSettingsProvider } from '~/types/notifications/settings'

export function useUseNotificationsManagementSettingsProvider () {
  const { fetch } = useCustomFetch()

  const { value, temp: tempSettings, bounce, instant } = useDebounceValue<NotificationsManagementSettings | undefined>(undefined, 1000)
  const isLoading = ref(false)
  let updateRequested = false

  async function refreshSettings () {
    isLoading.value = true
    const res = await fetch<NotificationsManagementDashboardResponse>(API_PATH.NOTIFICATIONS_MANAGEMENT_GENERAL)

    isLoading.value = false

    instant(res.data)
    return res.data
  }

  const settings = computed(() => tempSettings.value as NotificationsManagementSettings)
  const generalSettings = computed(() => settings.value as NotificationsManagementSettingsGeneralTab)

  function updateGeneralSettings (newSettings:NotificationsManagementSettingsGeneralTab) {
    if (tempSettings.value && newSettings) {
      updateRequested = true
      bounce({ ...settings.value, ...newSettings }, true, true)
    }
  }

  watch(value, (newSettings) => {
    if (updateRequested) {
      updateRequested = false
      try {
        warn('TODO: implement saving of new settings', newSettings)
      } catch (e) {
        refreshSettings()
      }
    }
  })

  provide<NotificationsManagementSettingsProvider>('notificationsManagementSettings', { generalSettings, isLoading, updateGeneralSettings })

  return { refreshSettings, isLoading }
}
