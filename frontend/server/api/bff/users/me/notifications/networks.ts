import type { FetchError } from 'ofetch'
import type { GetUserNotificationMachinesResponse } from '~/types/api/notifications'

export default defineEventHandler(async (event) => {
  const hasSessionCookie = getCookie(event, 'session_id')
  if (!hasSessionCookie) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized - No session cookie found',
    })
  }

  try {
    return await get<GetUserNotificationMachinesResponse>(event,
      '/users/me/notifications/machines',
    )
  }
  catch (error) {
    throw createError({
      statusCode: (error as FetchError).statusCode,
      statusMessage: (error as FetchError).statusMessage,
    })
  }
})
