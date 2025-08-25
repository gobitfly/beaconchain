import type { FetchError } from 'ofetch'
import type { GetUserDashboardsResponse } from '~/types/api/dashboard'

export default defineEventHandler(async (event) => {
  console.log('Fetching user dashboards')
  const hasSessionCookie = getCookie(event, 'session_id')
  if (!hasSessionCookie) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized - No session cookie found',
    })
  }

  try {
    return await get<GetUserDashboardsResponse>(event, '/users/me/dashboards')
  }
  catch (error) {
    throw createError({
      statusCode: (error as FetchError).statusCode,
      statusMessage: (error as FetchError).statusMessage,
    })
  }
})
