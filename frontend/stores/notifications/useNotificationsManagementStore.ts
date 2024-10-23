import type {
  InternalGetUserNotificationSettingsResponse,
  InternalPutUserNotificationSettingsGeneralResponse,
  InternalPutUserNotificationSettingsNetworksResponse,
  InternalPutUserNotificationSettingsPairedDevicesResponse,
  NotificationSettings,
  NotificationSettingsNetwork,
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
        machine_cpu_usage_threshold: 0.0,
        machine_memory_usage_threshold: 0.0,
        machine_storage_usage_threshold: 0.0,
      },
      has_machines: true,
      networks: [],
      paired_devices: [],
    },
  )

  const saveSettings = async () => {
    await fetch<InternalPutUserNotificationSettingsGeneralResponse>(
      API_PATH.NOTIFICATIONS_MANAGEMENT_SAVE, {
        body: settings.value.general_settings,
        method: 'PUT',
      })
  }
  const getSettings = () => {
    return fetch<InternalGetUserNotificationSettingsResponse>(
      API_PATH.NOTIFICATIONS_MANAGEMENT_GENERAL,
    )
  }

  const removeDevice = async (id: number) => {
    return await fetch(
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
    id: number,
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
  const setNotificationForNetwork = async ({
    chain_id,
    settings,
  }: {
    chain_id: string,
    settings: NotificationSettingsNetwork,
  }) => {
    await fetch<InternalPutUserNotificationSettingsNetworksResponse>(
      API_PATH.NOTIFICATIONS_MANAGEMENT_NETWORK_SET_NOTIFICATION,
      {
        body: {
          ...settings,
        },
      },
      {
        network: chain_id,
      },
    )
  }
  const setNotificationForClient = async ({
    client_id,
    is_subscribed,
  }: {
    client_id: number,
    is_subscribed: boolean,
  }) => {
    await fetch<InternalPutUserNotificationSettingsNetworksResponse>(
      API_PATH.NOTIFICATIONS_MANAGEMENT_CLIENTS_SET_NOTIFICATION,
      {
        body: {
          is_subscribed,
        },
      },
      {
        client_id,
      },
    )
  }

  return {
    getSettings,
    removeDevice,
    saveSettings,
    setNotificationForClient,
    setNotificationForNetwork,
    setNotificationForPairedDevice,
    settings,
  }
})
