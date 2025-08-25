import type { GetValidatorDashboardConsensusLayerConsolidationsResponse } from '~/types/api/validator_dashboard'

export default defineEventHandler((event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === undefined) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }

  return get<GetValidatorDashboardConsensusLayerConsolidationsResponse>(
    event,
    `/validator-dashboards/${dashboardId}/consensus-layer-consolidations`,
  )
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data?.error || 'Internal Server Error',
      })
    })
})
