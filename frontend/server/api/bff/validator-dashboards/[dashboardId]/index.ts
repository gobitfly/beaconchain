import { ERROR_CODE } from '~/shared/utils/helper'
import type { GetValidatorDashboardResponse } from '~/types/api/validator_dashboard'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }
  return await get<GetValidatorDashboardResponse>(
    event,
    `/validator-dashboards/${dashboardId}`,
  )
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
