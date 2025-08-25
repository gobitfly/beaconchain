const config = useRuntimeConfig()
type FetchParams = Parameters<typeof $fetch>

const baseURL = config.public.apiClient
const headers = {
  'x-ssr-secret': config.private.ssrSecret,
}

// export const fetchData = <T>(
//   endpoint: string,
//   options?: FetchParams[1],
// ) => $fetch<T>(endpoint, {
//   baseURL,
//   headers,
//   ...options,
// })
/**
 * Fetch data from our external api while giving access to the raw response (headers, etc.)
 * see: https://github.com/unjs/ofetch?tab=readme-ov-file#-access-to-raw-response
 */
export const fetchRaw = <T>(
  endpoint: string,
  options?: FetchParams[1],
) => $fetch.raw<T>(endpoint, {
  baseURL,
  headers,
  ...options,
})
