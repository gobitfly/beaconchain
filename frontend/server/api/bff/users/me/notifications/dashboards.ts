import type { FetchError } from 'ofetch'
import type { GetUserNotificationDashboardsResponse } from '~/types/api/notifications'

export default defineEventHandler(async (event) => {
  const hasSessionCookie = getCookie(event, 'session_id')
  if (!hasSessionCookie) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized - No session cookie found',
    })
  }

  try {
    const notificationDashboards = await get<GetUserNotificationDashboardsResponse>(event,
      '/users/me/notifications/dashboards',
    )

    return {
      data: notificationDashboards.data.map(item => ({
        ...item,
        identifier: `${item.is_account_dashboard}-${item.dashboard_id}-${item.group_id}-${item.epoch}`,
      })),
      paging: notificationDashboards.paging,
    }
  }
  catch (error) {
    throw createError({
      statusCode: (error as FetchError).statusCode,
      statusMessage: (error as FetchError).statusMessage,
    })
  }
})
