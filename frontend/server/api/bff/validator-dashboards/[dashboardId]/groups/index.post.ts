import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }

  const body = await readBody(event)

  return await post<VDBOverviewGroup>(
    event,
    `/validator-dashboards/${dashboardId}/groups`,
    {
      body,
    },
  )
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data?.error || 'Internal Server Error',
      })
    })
})
