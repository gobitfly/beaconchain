import type { InternalGetProductSummaryResponse } from '~/types/api/user'

export default defineEventHandler(async (event) => {
  return await get<InternalGetProductSummaryResponse>(event, '/product-summary')
    // .then(({ data }) => data)
    .catch((error) => {
      throw createError({
        statusCode: error.statusCode || 500,
        statusMessage: error.data.error || 'Internal Server Error',
      })
    })
})
