import type { FetchError } from 'ofetch'
import type { GetUserNotificationsValidatorDashboardResponse } from '~/types/api/notifications'

export default defineEventHandler(async (event) => {
  const hasSessionCookie = getCookie(event, 'session_id')
  if (!hasSessionCookie) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized - No session cookie found',
    })
  }
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === undefined) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }
  const groupId = getRouterParam(event, 'groupId')
  if (groupId === undefined) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing groupId',
    })
  }
  const epoch = getRouterParam(event, 'epoch')
  if (epoch === undefined) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing epoch',
    })
  }

  try {
    return await get<GetUserNotificationsValidatorDashboardResponse>(event,
      `/users/me/notifications/validator-dashboards/${dashboardId}/groups/${groupId}/epochs/${epoch}`,
    )
  }
  catch (error) {
    throw createError({
      statusCode: (error as FetchError).statusCode,
      statusMessage: (error as FetchError).statusMessage,
    })
  }
})
