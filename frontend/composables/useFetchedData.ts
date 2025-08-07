import type {
  AvailableRouterMethod,
  // InternalApi,
  NitroFetchRequest,
} from 'nitropack'
import type { FetchResult } from '#app'

export type ServerUrl = Extract<Exclude<NitroFetchRequest, object>, `/api/${string}`>

// export type ServerResponse<T extends ServerUrl> = InternalApi[T] extends { default: infer R }
//   ? R
//   :
//     | (InternalApi[T] extends { get: infer Q }
//       ? Q : never)
//     | (InternalApi[T] extends { patch: infer Q }
//       ? Q : never)
//     | (InternalApi[T] extends { post: infer Q }
//       ? Q : never)
//     | (InternalApi[T] extends { put: infer Q }
//       ? Q : never)
// export type ServerResponse<T extends ServerUrl> = InternalApi[T] extends { default: infer R }
//   ? R
//   :
//     | (InternalApi[T] extends { get: infer Q }
//       ? Q : never)
//     | (InternalApi[T] extends { patch: infer Q }
//       ? Q : never)
//     | (InternalApi[T] extends { post: infer Q }
//       ? Q : never)
//     | (InternalApi[T] extends { put: infer Q }
//       ? Q : never)
/**
     * Create a Key in {@link useFetchedData} first and map the ResponseType
     * (from our backend route) as this can not be inferred automatically via nuxt
     * at the moment 😪
     *
     * Might be solved in nuxt 4 (via useFetch factory function)
*/
// export type Key = keyof Response

// type Response = {
//   'latestState': InternalApi['/api/latest-state']['default'],
//   'users/:id': InternalApi['/api/users/me']['default'],
//   'users/me': InternalApi['/api/users/me']['default'],
// }

// type ReplaceColonsWithTemplate<Path extends string> =
//   Path extends `${infer Start}:${infer Param}/${infer Rest}`
//     ? `${Start}\${string}/${ReplaceColonsWithTemplate<Rest>}`
//     : Path extends `${infer Start}:${infer Param}`
//       ? `${Start}\${string}`
//       : Path

// export const useFetchedData = <T extends ServerUrl>(key: T) => {
//   const { data } = useNuxtData <ServerResponse<T>>(key)
//   return data
// }
type GetReturnTypeFromServerUrl<T extends ServerUrl> = FetchResult<T, AvailableRouterMethod<T>>
/**
 * This enables `type inference` for useFetchedData for `endpoint aliases`
 */
type ReturnType = {
  dashboardOverview: GetReturnTypeFromServerUrl<'/api/validator-dashboards/:dashboardId'>,
  productSummary: GetReturnTypeFromServerUrl<'/api/product-summary'>,
  user: GetReturnTypeFromServerUrl<'/api/users/me'>,
  validators: GetReturnTypeFromServerUrl<'/api/validator-dashboards/:dashboardId/validators'>,
}
export type Key = keyof ReturnType
export const useFetchedData = <T extends Key | LooseAutocomplete<ServerUrl>>(key: T) => {
  // const { data } = useNuxtData <T extends Key ? ReturnType[Key] : FetchResult<T, AvailableRouterMethod<T>>>(key)
  const { data } = useNuxtData<T extends Key ? ReturnType[T] : FetchResult<T, AvailableRouterMethod<T>>>(key)
  return data
}
