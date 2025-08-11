import type { GetValidatorDashboardGroupRewardsResponse } from '~/types/api/validator_dashboard'

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
  const epoch = getRouterParam(event, 'epoch')
  if (epoch === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing epoch',
    })
  }
  const query = getQuery(event)
  return await get<GetValidatorDashboardGroupRewardsResponse>(
    event,
    `/validator-dashboards/${dashboardId}/groups/${groupId}/rewards/${epoch}`,
    {
      query,
    },
  )
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data.error || 'Internal Server Error',
      })
    })
})
