import { ERROR_CODE } from '~/shared/utils/helper'
import type { GetValidatorDashboardSlotVizResponse } from '~/types/api/slot_viz'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  // console.log('dashboardId', dashboardId)
  if (dashboardId === '') {
    // throw Error('Missing dashboardId')
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }

  const { group_ids } = getQuery(event)
  return await get<GetValidatorDashboardSlotVizResponse>(
    event,
    `/validator-dashboards/${dashboardId}/slot-viz`,
    { query: { group_ids } },
  )
    // .then(({ data }) => data)
    .catch((error) => {
      if (error?.data?.error === 'bad request: effective balance of validators in list is too high, maximum is 640000000000') {
        throw createError({
          statusCode: error.statusCode || 500,
          statusMessage: ERROR_CODE.EFFECTIVE_BALANCE_EXCEEDS_LIMIT || 'Internal Server Error',
        })
      }
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data.error || 'Internal Server Error',
      })
    })
})
