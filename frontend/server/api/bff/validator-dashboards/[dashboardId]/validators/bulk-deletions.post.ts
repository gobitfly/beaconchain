import { fetchWithContext } from '~/server/utils/fetchWithContext'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }

  const body = await readBody(event)
  return await fetchWithContext(
    event,
    `/validator-dashboards/${dashboardId}/validators/bulk-deletions`,
    {
      body,
      method: 'post',
    },
  )
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data.error || 'Internal Server Error',
      })
    })
})
