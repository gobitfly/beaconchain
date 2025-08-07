import type { H3Event } from 'h3'
import type { ApiResponse } from '~/types/api/common'

const config = useRuntimeConfig()
const headers = {
  'x-ssr-secret': config.private.ssrSecret,
}
/**
 * Executes a `get request`, while it forwards headers and cookies to our external API
 * and denormilizes response if it contains no paging information.
 */
export const get = <T extends ApiResponse>(
  event: H3Event,
  endpoint: string,
  options?: Parameters<H3Event['$fetch']>[1],
) => {
  const query = getQuery(event)
  return (event.$fetch<T>(endpoint, {
    query,
    ...options,
    baseURL: config.public.apiClient,
    headers,
    method: 'get',
  }) as Promise<T>)
    .then((response) => {
      if (!response.paging) return response.data
      return response
    }) as (T extends { paging: any } ? Promise<T> : Promise<T['data']>)
}
