import type {
  AvailableRouterMethod,
  // NitroFetchOptions,
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
  watch?: UseFetchOptions<U>['watch'],
  /**
   * Time to live for the cached data in seconds.
   */
  // ttl?: number,
}

// type UseFetch = typeof useFetch

// type InternalApiResponse<T extends LooseAutocomplete<ServerUrl>> = FetchResult<T, AvailableRouterMethod<T>>
// type InternalApiResponse<T extends NitroFetchRequest> = FetchResult<T, AvailableRouterMethod<T>>

export const useApi = function useApi<T extends LooseAutocomplete<ServerUrl>>(
  url: (() => T) | T,
  options?: UseApiOptions<FetchResult<T, AvailableRouterMethod<T>>>,
  // options?: UseApiOptions<GetReturnTypeFromServerUrl<T>>,
) {
  return useFetch(url, {
    key: url,
    // onRequest: ({
    //   options,
    //   request: requestUrl,
    // }) => {
    //   const abortController = new AbortController()
    //   const hasEmptyParameter = `${requestUrl}`.includes('//')
    //   const hasUndefinedParameter = `${requestUrl}`.includes('undefined')
    //   options.signal = abortController.signal
    //   if (hasEmptyParameter || hasUndefinedParameter) {
    //     abortController.abort()
    //     return null
    //   }
    //   if (isDevEnvironment && abortController?.signal.aborted) {
    //     // eslint-disable-next-line no-console
    //     console.log('ℹ️', 'request aborted', {
    //       hasEmptyParameter, hasUndefinedParameter, requestUrl,
    //     })
    //   }
    // },
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
