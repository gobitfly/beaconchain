import type { InternalPostSearchResponse } from '~/types/api/search'
import { FetchError } from 'ofetch'
import type { ApiErrorResponse } from '~/types/api/common'

// Replace `chain_id`'s `string` type with our more specific `ChainId` type
export type InternalPostSearchResponseWithChainId = Omit<InternalPostSearchResponse, 'data'> & {
  data: Array<ReplaceChainId<InternalPostSearchResponse['data'][number]>>,
}
type ReplaceChainId<T> = T extends { chain_id: any }
  ? Omit<T, 'chain_id'> & { chain_id: ChainId }
  : T

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const body = await readBody(event)
  const headers = {
    'x-ssr-secret': config.private.ssrSecret,
  }
  const { networks } = body

  try {
    const networkCalls: Promise<InternalPostSearchResponseWithChainId>[] = []

    if (networks.includes(1)) {
      networkCalls.push(
        $fetch<InternalPostSearchResponseWithChainId>('/search', {
          baseURL: config.private.apiServerMainnet,
          body: {
            ...body, networks: [ 1 ],
          },
          headers,
          method: 'POST',
        }),
      )
    }
    if (networks.includes(560048)) {
      networkCalls.push(
        $fetch<InternalPostSearchResponseWithChainId>('/search', {
          baseURL: config.private.apiServerHoodi,
          body: {
            ...body, networks: [ 560048 ],
          },
          headers,
          method: 'POST',
        }),
      )
    }

    const results = await Promise.all(networkCalls)
    return results.flatMap(result => result.data)
  }
  catch (error) {
    if (!(error instanceof FetchError)) {
      throw createError({
        message: 'Error is of unexpected type',
        statusCode: 500,
      })
    }

    throw createError({
      message: (error.data as ApiErrorResponse).error,
      statusCode: error.status,
    })
  }
})
