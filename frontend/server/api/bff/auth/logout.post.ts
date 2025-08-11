import type { FetchError } from 'ofetch'

export default defineEventHandler(async (event) => {
  try {
    await post(event, '/logout')
    deleteCookie(event, 'session_id', {
      domain: 'beaconcha.in',
    })
  }
  catch (error) {
    throw createError({
      statusCode: (error as FetchError).statusCode,
      statusMessage: (error as FetchError).statusMessage,
    })
  }
})
