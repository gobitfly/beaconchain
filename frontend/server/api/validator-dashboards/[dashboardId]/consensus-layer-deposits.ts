import type { GetValidatorDashboardConsensusLayerDepositsResponse } from '~/types/api/validator_dashboard'

export default defineEventHandler((event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === undefined) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }

  return get<GetValidatorDashboardConsensusLayerDepositsResponse>(
    event,
    `/validator-dashboards/${dashboardId}/consensus-layer-deposits`,
  )
})
