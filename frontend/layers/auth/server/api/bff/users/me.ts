import type { FetchError } from 'ofetch'
import type { InternalGetUserInfoResponse } from '~/types/api/user'

const config = useRuntimeConfig()
const headers = {
  'x-ssr-secret': config.private.ssrSecret,
}

export default defineEventHandler(async (event) => {
  const hasSessionCookie = getCookie(event, 'session_id')
  if (!hasSessionCookie) return

  try {
    return await event.$fetch<InternalGetUserInfoResponse>('/users/me', {
      baseURL: config.private.apiServer,
      headers,
    }).then(({ data }) => data)
  }
  catch (error) {
    throw createError({
      statusCode: (error as FetchError).statusCode,
      statusMessage: (error as FetchError).statusMessage,
    })
  }
})
