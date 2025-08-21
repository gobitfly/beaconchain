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
    return await put<GetUserNotificationDashboardsResponse>(event,
      '/users/me/notifications/settings/general',
    )
  }
  catch (error) {
    throw createError({
      statusCode: (error as FetchError).statusCode,
      statusMessage: (error as FetchError).statusMessage,
    })
  }
})
