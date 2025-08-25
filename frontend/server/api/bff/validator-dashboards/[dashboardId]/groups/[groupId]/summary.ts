import type { GetValidatorDashboardGroupSummaryResponse } from '~/types/api/validator_dashboard'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }
  const groupId = getRouterParam(event, 'groupId')
  if (groupId === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing groupId',
    })
  }
  return await get<GetValidatorDashboardGroupSummaryResponse>(
    event,
    `/validator-dashboards/${dashboardId}/groups/${groupId}/summary`,
    {
      query: getQuery(event),
    },
  )
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data.error || 'Internal Server Error',
      })
    })
})
