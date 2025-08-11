import type {
  GetValidatorDashboardExecutionLayerWithdrawalsResponse,
  GetValidatorDashboardTotalExecutionWithdrawalsResponse,
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
    elWithdrawals,
    totalElWithdrawals,
  ] = await Promise.all([
    get<GetValidatorDashboardExecutionLayerWithdrawalsResponse>(
      event,
      `/validator-dashboards/${dashboardId}/execution-layer-withdrawals`,
    ),
    get<GetValidatorDashboardTotalExecutionWithdrawalsResponse>(
      event,
      `/validator-dashboards/${dashboardId}/total-execution-layer-withdrawals`,
    ),
  ])
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data?.error || 'Internal Server Error',
      })
    })

  // Each response list item should have a unique identifier.
  const elWithdrawalsWithIdentifier = {
    data: elWithdrawals.data.map(item => ({
      ...item,
      identifier: `${item.block_queued}-${item.tx_index_queued}-${item.itx_index_queued}`,
    })),
    paging: elWithdrawals.paging,
  }

  return {
    ...elWithdrawalsWithIdentifier,
    total_amount: totalElWithdrawals.total_amount,
  }
})
