import type { H3Event } from 'h3'
import type { ApiResponse } from '~/types/api/common'

const config = useRuntimeConfig()
const headers = {
  'x-ssr-secret': config.private.ssrSecret,
}
/**
 * Executes a `get request`, while it forwards headers and cookies (context)
 * (as our external api needs, for example needs the `session_id` cookie)
 *
 * Returns denormilized response if there is no paging information
 *
 * @example Denormalized Response
 *
 * ✅:
 * {
 *   data: {...}
 *   paging: {...}
 * }
 *
 * ✅:
 * {
 *   ...
 * }
 *
 * ❌:
 * {
 *   data: {...}
 * }
 */
export const fetchWithContext = <T extends ApiResponse>(
  event: H3Event,
  endpoint: string,
  options?: Parameters<H3Event['$fetch']>[1],
) => {
  const query = getQuery(event)
  // if (options?.method === 'delete') {
  //   return event.$fetch<T>(endpoint, {
  //     ...options,
  //     baseURL: config.public.apiClient,
  //   })
  // }
  return (event.$fetch<T>(endpoint, {
    query,
    ...options,
    baseURL: config.public.apiClient,
    headers,
    // method: 'delete',
  }))
    .then((response) => {
      if (!response) return
      if (!response.paging) return response.data
      return response
    }) as (T extends { paging: any } ? Promise<T> : Promise<T['data']>)
}
