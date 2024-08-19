import { warn } from 'vue'
import { API_PATH } from '~/types/customFetch'
import type {
  InternalGetUserNotificationSettingsResponse,
  NotificationPairedDevice,
  NotificationSettings,
} from '~/types/api/notifications'
import type {
  NotificationsManagementSettingsGeneralTab,
  NotificationsManagementSettingsProvider,
} from '~/types/notifications/settings'

export function useUseNotificationsManagementSettingsProvider() {
  const { fetch } = useCustomFetch()

  const {
    bounce,
    instant,
    temp: tempSettings,
    value,
  } = useDebounceValue<NotificationSettings | undefined>(undefined, 1000)
  const isLoading = ref(false)
  let updateRequested = false

  async function refreshSettings() {
    isLoading.value = true
    const res = await fetch<InternalGetUserNotificationSettingsResponse>(
      API_PATH.NOTIFICATIONS_MANAGEMENT_GENERAL,
    )

    isLoading.value = false

    instant(res.data)
    return res.data
  }

  const settings = computed(() => tempSettings.value as NotificationSettings)
  const generalSettings = computed(
    () =>
      settings.value
        .general_settings as NotificationsManagementSettingsGeneralTab,
  )
  const pairedDevices = computed(() => settings.value.paired_devices)

  function updateGeneralSettings(
    newSettings: NotificationsManagementSettingsGeneralTab,
  ) {
    if (tempSettings.value && newSettings) {
      updateRequested = true
      const original: NotificationSettings
        = tempSettings.value as NotificationSettings
      bounce(
        {
          ...original,
          general_settings: {
            ...original.general_settings,
            ...newSettings,
          },
        },
        true,
        true,
      )
    }
  }

  function updatePairedDevices(newDevices: NotificationPairedDevice[]) {
    if (tempSettings.value && newDevices) {
      updateRequested = true
      const original: NotificationSettings
        = tempSettings.value as NotificationSettings
      bounce({
        ...original,
        paired_devices: newDevices,
      }, true, true)
    }
  }

  watch(value, (newSettings) => {
    if (updateRequested) {
      updateRequested = false
      try {
        // adding newSettings here so the parameter is used :)
        warn(`TODO: implement saving of new settings ${newSettings}`)
      }
      catch (e) {
        refreshSettings()
      }
    }
  })

  provide<NotificationsManagementSettingsProvider>(
    'notificationsManagementSettings',
    {
      generalSettings,
      isLoading,
      pairedDevices,
      updateGeneralSettings,
      updatePairedDevices,
    },
  )

  return {
    isLoading,
    refreshSettings,
  }
}
