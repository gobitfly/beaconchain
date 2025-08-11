import type { GetValidatorDashboardExecutionLayerConsolidationsResponse } from '~/types/api/validator_dashboard'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === undefined) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }

  const elConsolidations = await get<GetValidatorDashboardExecutionLayerConsolidationsResponse>(
    event,
    `/validator-dashboards/${dashboardId}/execution-layer-consolidations`,
  )
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data?.error || 'Internal Server Error',
      })
    })

  // Each response list item should have a unique identifier.
  const elConsolidationsWithIdentifier = {
    data: elConsolidations.data.map(item => ({
      ...item,
      identifier: `${item.block_queued}-${item.tx_index_queued}-${item.itx_index_queued}`,
    })),
    paging: elConsolidations.paging,
  }

  return elConsolidationsWithIdentifier
})
