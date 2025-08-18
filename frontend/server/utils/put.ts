import type { H3Event } from 'h3'

const config = useRuntimeConfig()
const headers = {
  'x-ssr-secret': config.private.ssrSecret,
}
/**
 * Executes a `put request`, while it forwards headers and cookies to our external API
 */
export const put = <T>(
  event: H3Event,
  endpoint: string,
  options?: Parameters<H3Event['$fetch']>[1],
) => {
  return event.$fetch<T>(endpoint, {
    ...options,
    baseURL: config.public.apiClient,
    headers,
    method: 'put',
  })
}
