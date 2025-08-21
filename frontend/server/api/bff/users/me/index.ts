import type { FetchError } from 'ofetch'
import type { InternalGetUserInfoResponse } from '~/types/api/user'

export default defineEventHandler(async (event) => {
  const hasSessionCookie = getCookie(event, 'session_id')
  if (!hasSessionCookie) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized - No session cookie found',
    })
  }

  try {
    return await get<InternalGetUserInfoResponse>(event, '/users/me')
  }
  catch (error) {
    throw createError({
      statusCode: (error as FetchError).statusCode,
      statusMessage: (error as FetchError).statusMessage,
    })
  }
})
