import type { UseFetchOptions } from 'nuxt/app'

export function useApi<T>(
  url: Parameters<typeof useFetch>[0],
  options?: UseFetchOptions<T>,
) {
  return useFetch(url, {
    ...options,
    $fetch: useNuxtApp().$fetchApi,
  })
}
