import type { FetchError } from 'ofetch'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  try {
    const {
      headers,
    } = await fetchRaw('/login', {
      body,
      method: 'POST',
    })

    const sessionCookie = headers.getSetCookie()
    appendHeader(event, 'set-cookie', sessionCookie)
    return sendNoContent(event)
  }
  catch (error) {
    throw createError({
      statusCode: (error as FetchError).statusCode,
      statusMessage: (error as FetchError).statusMessage,
    })
  }
})
