import type {
  ApiDataResponse,
  VDBPublicId,
} from '~/types/api/common'

export default defineEventHandler(async (event) => {
  const dashboardId = getRouterParam(event, 'dashboardId')
  if (dashboardId === '') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing dashboardId',
    })
  }
  // const publicId = getRouterParam(event, 'publicId')
  // if (publicId === '') {
  //   throw createError({
  //     statusCode: 400,
  //     statusMessage: 'Missing publicId',
  //   })
  // }
  // const query = getQuery(event)
  const body = await readBody(event)
  return await post<ApiDataResponse<VDBPublicId>>(
    event,
    `/validator-dashboards/${dashboardId}/public-ids`,
    {
      body,
    },
    // },
  )
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data.error || 'Internal Server Error',
      })
    })
})
