import type {
  AvailableRouterMethod,
  // NitroFetchRequest,
  // NitroFetchRequest,
} from 'nitropack'
import type {
  // AsyncDataOptions,
  FetchResult,
  UseFetchOptions,
} from '#app'

// type useApiDenormalization<T> = T extends { data: any } ? T['data'] : T
type UseApiOptions<U> = {
  body?: UseFetchOptions<U>['body'],
  getCachedData?: UseFetchOptions<U>['getCachedData'],
  immediate?: boolean,
  key?: MaybeRefOrGetter<Key | LooseAutocomplete<ServerUrl>>,
  lazy?: boolean,
  query?: UseFetchOptions<U>['query'],
  transform?: UseFetchOptions<U>['transform'],
  /**
   * Time to live for the cached data in seconds.
   */
  ttl?: number,
  watch?: UseFetchOptions<U>['watch'],
}

// type UseFetch = typeof useFetch

// type InternalApiResponse<T extends LooseAutocomplete<ServerUrl>> = FetchResult<T, AvailableRouterMethod<T>>
// type InternalApiResponse<T extends NitroFetchRequest> = FetchResult<T, AvailableRouterMethod<T>>

export const useApi = function useApi<T extends LooseAutocomplete<ServerUrl>>(
  url: (() => T) | T,
  options?: UseApiOptions<FetchResult<T, AvailableRouterMethod<T>>>,
) {
  // const lastFetchedAt = Date.now()
  return useFetch(url, {
    // getCachedData: (key, nuxtApp) => nuxtApp.payload[key] ?? nuxtApp.payload.data[key],
    // getCachedData: (key, nuxtApp) => {
    //   // lastFetchedAt = Date.now()
    //   const now = Date.now()
    //   console.log('getCachedData', lastFetchedAt, now, lastFetchedAt - now)
    //   // lastFetchedAt = Date.now()
    //   return nuxtApp.payload[key] ?? nuxtApp.payload.data[key]
    // },
    // key: typeof url === 'function' ? () => url() : url,
    key: url,
    ...options,
    $fetch: useNuxtApp().$api as typeof $fetch,
  })
}

// const useTest = (url: LooseAutocomplete<ServerUrl>) => useAsyncData(url, () => $fetch(url))

// export const useApi = function useApi<U extends LooseAutocomplete<ServerUrl>, R extends InternalApiResponse<U>>(
// // export const useApi = function useApi<T extends $Fetch>(
//   url: U,
//   options?: AsyncDataOptions<R>,
//   // options?: UseApiOptions<InternalApiResponse<T>>,
//   // options?: AsyncDataOptions<InternalApiResponse<T>> & { key: string },
// ) {
//   // const { key = url } = options || {}
//   const { $api } = useNuxtApp()
//   const handler = (ctx?: NuxtApp) => $api<R>(url)
//   // return useAsyncData(key, () => $api(url), {
//   const key = url
//   return useAsyncData<R>(key, handler, options)
//   // return useAsyncData(key, () => $api(url), {
//   // ...options,
//   // key: url,
//   // transform:(response: ) => {
//   //   if (hasKey(response, 'paging')) return response
//   //   if (hasKey(response, 'data')) return response.data
//   //   return response
//   // },
//   // ...options,
//   // $fetch: useNuxtApp().$api as typeof $fetch,
//   // })
// }
