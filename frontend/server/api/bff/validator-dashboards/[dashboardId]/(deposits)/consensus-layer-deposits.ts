import type {
  GetValidatorDashboardConsensusLayerDepositsResponse,
  GetValidatorDashboardTotalConsensusDepositsResponse,
} from '~/types/api/validator_dashboard'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === undefined) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }

  const [
    clDeposits,
    totalClDeposits,
  ] = await Promise.all([
    get<GetValidatorDashboardConsensusLayerDepositsResponse>(
      event,
      `/validator-dashboards/${dashboardId}/consensus-layer-deposits`,
    ),
    get<GetValidatorDashboardTotalConsensusDepositsResponse>(
      event,
      `/validator-dashboards/${dashboardId}/total-consensus-layer-deposits`,
    ),
  ])
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data?.error || 'Internal Server Error',
      })
    })

  // Each response list item should have a unique identifier.
  const depositsWithIdentifier = {
    data: clDeposits.data.map(item => ({
      ...item,
      identifier: `${item.slot}-${item.slot_index}`,
    })),
    paging: clDeposits.paging,
  }

  return {
    ...depositsWithIdentifier,
    total_amount: totalClDeposits.total_amount,
  }
})
