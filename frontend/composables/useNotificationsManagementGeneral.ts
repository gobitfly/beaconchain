import { inject, warn } from 'vue'
import type { NotificationsManagementGeneralProvider } from '~/types/notifications/general'

export function useNotificationsManagementGeneral () {
  const general = inject<NotificationsManagementGeneralProvider>('notificationsManagementGeneral')

  const refreshGeneralSettings = async () => {
    if (!general) {
      warn('notifications management general provider not injected')
      return
    }

    await general.refresh()
  }

  const generalSettings = computed(() => {
    if (!general) {
      warn('notifications management general provider not injected')
      return
    }

    return general?.settings.value
  })

  const isLoading = computed(() => {
    if (!general) {
      warn('notifications management general provider not injected')
      return
    }

    return general?.isLoading.value
  })

  return { generalSettings, refreshGeneralSettings, isLoading }
}
