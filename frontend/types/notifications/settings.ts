import type {
  NotificationPairedDevice,
  NotificationSettingsGeneral,
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
  >,
  isLoading: Ref<boolean>,
  pairedDevices: ComputedRef<NotificationPairedDevice[] | undefined>,
  updateGeneralSettings: (
    newSettings: NotificationsManagementSettingsGeneralTab,
  ) => void,
  updatePairedDevices: (newDevices: NotificationPairedDevice[]) => void,
}
