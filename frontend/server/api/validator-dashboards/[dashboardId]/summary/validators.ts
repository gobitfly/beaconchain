import type { GetValidatorDashboardSummaryValidatorsResponse } from '~/types/api/validator_dashboard'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }

  return await get<GetValidatorDashboardSummaryValidatorsResponse>(
    event,
    `/validator-dashboards/${dashboardId}/summary/validators`,
  )
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data.error || 'Internal Server Error',
      })
    })
})
