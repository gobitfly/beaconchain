import type { GetValidatorDashboardBlocksResponse } from '~/types/api/validator_dashboard'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }
  const query = getQuery(event)
  return await get<GetValidatorDashboardBlocksResponse>(
    event,
    `/validator-dashboards/${dashboardId}/blocks`,
    {
      query,
    },
  )
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data.error || 'Internal Server Error',
      })
    })
})
