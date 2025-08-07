import type { GetValidatorDashboardDutiesResponse } from '~/types/api/validator_dashboard'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }
  const epoch = getRouterParam(event, 'epoch')
  if (epoch === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing epoch',
    })
  }
  const query = getQuery(event)
  return await get<GetValidatorDashboardDutiesResponse>(
    event,
    `/validator-dashboards/${dashboardId}/duties/${epoch}`,
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
