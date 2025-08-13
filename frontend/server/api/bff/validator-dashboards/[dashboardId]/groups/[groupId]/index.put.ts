import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  const groupId = getRouterParam(event, 'groupId')
  if (dashboardId === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }
  if (groupId === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing groupId',
    })
  }

  const body = await readBody(event)

  return await put<VDBOverviewGroup>(
    event,
    `/validator-dashboards/${dashboardId}/groups/${groupId}`,
    {
      body,
    },
  )
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data?.error || 'Internal Server Error',
      })
    })
})
