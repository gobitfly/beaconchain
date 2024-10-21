import type {
  InternalGetUserNotificationsValidatorDashboardResponse,
  NotificationDashboardsTableRow, NotificationValidatorDashboardDetail,
} from '~/types/api/notifications'
import { API_PATH } from '~/types/customFetch'

export const useNotificationsDashboardDetailsStore = defineStore('notifications-dashboard-details', () => {
  const { fetch } = useCustomFetch()
  const detailsList = ref(new Map())

  const getDetails = async ({
    dashboard_id,
    epoch,
    group_id,
    search,
  }:
    {
      search?: string,
    }
    & Pick<NotificationDashboardsTableRow, 'dashboard_id' | 'epoch' | 'group_id'>,
  ) => {
    return fetch<InternalGetUserNotificationsValidatorDashboardResponse>(
      API_PATH.NOTIFICATIONS_DASHBOARDS_DETAILS_VALIDATOR,
      {
        query: {
          search,
        },
      },
      {
        dashboard_id,
        epoch,
        group_id,
      },
    )
  }

  const addDetails = ({
    details,
    identifier,
  }: {
    details: NotificationValidatorDashboardDetail,
    identifier: string,
  }) => {
    detailsList.value.set(identifier, details)
  }

  return {
    addDetails,
    detailsList,
    getDetails,
  }
})
