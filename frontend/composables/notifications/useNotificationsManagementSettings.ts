import { inject, warn } from 'vue'
import type { NotificationsManagementSettingsGeneralTab, NotificationsManagementSettingsProvider } from '~/types/notifications/settings'

export function useNotificationsManagementSettings () {
  const provider = inject<NotificationsManagementSettingsProvider>('notificationsManagementSettings')

  const updateGeneralSettings = (newSettings: NotificationsManagementSettingsGeneralTab) => {
    if (!provider) {
      warn('notifications management settings provider not injected')
      return
    }

    provider.updateGeneralSettings(newSettings)
  }

  const generalSettings = computed(() => {
    if (!provider) {
      warn('notifications management settings provider not injected')
      return
    }

    return provider?.generalSettings.value
  })

  const isLoading = computed(() => {
    if (!provider) {
      warn('notifications management settings provider not injected')
      return
    }

    return provider?.isLoading.value
  })

  return { generalSettings, isLoading, updateGeneralSettings }
}
