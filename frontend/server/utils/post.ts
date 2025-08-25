import type { H3Event } from 'h3'
import type { ApiDataResponse } from '~/types/api/common'

const config = useRuntimeConfig()
const headers = {
  'x-ssr-secret': config.private.ssrSecret,
}
/**
 * Executes a `post request`, while it forwards headers and cookies to our external API
 */
export const post = <T>(
  event: H3Event,
  endpoint: string,
  options?: Parameters<H3Event['$fetch']>[1],
) => {
  return event.$fetch<ApiDataResponse<T>>(endpoint, {
    ...options,
    baseURL: config.public.apiClient,
    headers,
    method: 'post',
  })
    .then((response) => {
      if (response && response.data) {
        return response.data
      }
    })
}
