import type { FetchError } from 'ofetch'
import type { GetUserNotificationNetworksResponse } from '~/types/api/notifications'

export default defineEventHandler(async (event) => {
  const hasSessionCookie = getCookie(event, 'session_id')
  if (!hasSessionCookie) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized - No session cookie found',
    })
  }

  try {
    return await get<GetUserNotificationNetworksResponse>(event,
      '/users/me/notifications/networks',
    )
  }
  catch (error) {
    throw createError({
      statusCode: (error as FetchError).statusCode,
      statusMessage: (error as FetchError).statusMessage,
    })
  }
})
