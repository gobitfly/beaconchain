import type {
  InternalGetUserNotificationSettingsResponse,
  InternalPutUserNotificationSettingsGeneralResponse,
  InternalPutUserNotificationSettingsPairedDevicesResponse,
  NotificationSettings,
} from '~/types/api/notifications'
import { API_PATH } from '~/types/customFetch'

export const useNotificationsManagementStore = defineStore('notifications-management-store', () => {
  const { fetch } = useCustomFetch()
  const settings = ref<NotificationSettings>(
    {
      clients: [],
      general_settings: {
        do_not_disturb_timestamp: 0,
        is_email_notifications_enabled: false,
        is_machine_cpu_usage_subscribed: false,
        is_machine_memory_usage_subscribed: false,
        is_machine_offline_subscribed: false,
        is_machine_storage_usage_subscribed: false,
        is_push_notifications_enabled: false,
        is_rocket_pool_max_collateral_subscribed: false,
        is_rocket_pool_min_collateral_subscribed: false,
        is_rocket_pool_new_reward_round_subscribed: false,
        machine_cpu_usage_threshold: 0.0,
        machine_memory_usage_threshold: 0.0,
        machine_storage_usage_threshold: 0.0,
        rocket_pool_max_collateral_threshold: 0,
        rocket_pool_min_collateral_threshold: 0,
      },
      networks: [ {
        chain_id: 0,
        settings: {
          gas_above_threshold: '0.0',
          gas_below_threshold: '0.0',
          is_gas_above_subscribed: false,
          is_gas_below_subscribed: false,
          is_participation_rate_subscribed: false,
          participation_rate_threshold: 0,
        },
      } ],
      paired_devices: [ {
        id: '',
        is_notifications_enabled: false,
        name: '',
        paired_timestamp: 0,
      } ],
    },
  )

  const saveSettings = async () => {
    await fetch<InternalPutUserNotificationSettingsGeneralResponse>(
      API_PATH.NOTIFICATIONS_MANAGEMENT_SAVE, {
        body: settings.value.general_settings,
        method: 'PUT',
      })
  }
  const getSettings = async () => {
    const { data } = await fetch<InternalGetUserNotificationSettingsResponse>(
      API_PATH.NOTIFICATIONS_MANAGEMENT_GENERAL,
    )
    settings.value = data
  }

  const removeDevice = async (id: string) => {
    await fetch(
      API_PATH.NOTIFICATIONS_MANAGEMENT_PAIRED_DEVICES_DELETE,
      {},
      {
        paired_device_id: id,
      },
    ).then(() => {
      // using optimistic ui here to avoid calling the api after delete
      settings.value.paired_devices
      = [ ...settings.value.paired_devices.filter(device => device.id !== id) ]
    })
  }
  const setNotificationForPairedDevice = async ({
    id,
    value,
  }: {
    id: string,
    value: boolean,
  }) => {
    await fetch<InternalPutUserNotificationSettingsPairedDevicesResponse>(
      API_PATH.NOTIFICATIONS_MANAGEMENT_PAIRED_DEVICES_SET_NOTIFICATION,
      {
        body: {
          is_notifications_enabled: value,
          name: id,
        },
      },
      {
        paired_device_id: id,
      },
    )
  }

  return {
    getSettings,
    removeDevice,
    saveSettings,
    setNotificationForPairedDevice,
    settings,
  }
})
