import type {
  GetValidatorDashboardExecutionLayerDepositsResponse,
  GetValidatorDashboardTotalExecutionDepositsResponse,
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
    elDeposits,
    totalElDeposits,
  ] = await Promise.all([
    get<GetValidatorDashboardExecutionLayerDepositsResponse>(
      event,
      `/validator-dashboards/${dashboardId}/execution-layer-deposits`,
    ),
    get<GetValidatorDashboardTotalExecutionDepositsResponse>(
      event,
      `/validator-dashboards/${dashboardId}/total-execution-layer-deposits`,
    ),
  ])
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data?.error || 'Internal Server Error',
      })
    })

  // Each response list item should have a unique identifier.
  const elDepositsWithIdentifier = {
    data: elDeposits.data.map(item => ({
      ...item,
      identifier: `${item.block}-${item.block_index}`,
    })),
    paging: elDeposits.paging,
  }

  return {
    ...elDepositsWithIdentifier,
    total_amount: totalElDeposits.total_amount,
  }
})
