import type {
  GetValidatorDashboardConsensusLayerWithdrawalsResponse,
  GetValidatorDashboardTotalConsensusWithdrawalsResponse,
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
    clWithdrawals,
    totalClWithdrawals,
  ] = await Promise.all([
    get<GetValidatorDashboardConsensusLayerWithdrawalsResponse>(
      event,
      `/validator-dashboards/${dashboardId}/consensus-layer-withdrawals`,
    ),
    get<GetValidatorDashboardTotalConsensusWithdrawalsResponse>(
      event,
      `/validator-dashboards/${dashboardId}/total-consensus-layer-withdrawals`,
    ),
  ])
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data?.error || 'Internal Server Error',
      })
    })

  // Each response list item should have a unique identifier.
  const clWithdrawalsWithIdentifier = {
    data: clWithdrawals.data.map(item => ({
      ...item,
      identifier: `${item.slot}-${item.slot_index}`,
    })),
    paging: clWithdrawals.paging,
  }

  return {
    ...clWithdrawalsWithIdentifier,
    total_amount: totalClWithdrawals.total_amount,
  }
})
