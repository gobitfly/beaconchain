import type {
  NotificationSettingsGeneral,
  NotificationPairedDevice,
} from '~/types/api/notifications'

export type NotificationsManagementSettingsGeneralTab = Pick<
  NotificationSettingsGeneral,
  | 'do_not_disturb_timestamp'
  | 'is_email_notifications_enabled'
  | 'is_push_notifications_enabled'
>

export interface NotificationsManagementSettingsProvider {
  generalSettings: ComputedRef<
    NotificationsManagementSettingsGeneralTab | undefined
  >
  updateGeneralSettings: (
    newSettings: NotificationsManagementSettingsGeneralTab,
  ) => void
  pairedDevices: ComputedRef<NotificationPairedDevice[] | undefined>
  updatePairedDevices: (newDevices: NotificationPairedDevice[]) => void
  isLoading: Ref<boolean>
}
