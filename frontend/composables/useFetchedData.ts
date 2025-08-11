import type {
  AvailableRouterMethod,
  // InternalApi,
  NitroFetchRequest,
} from 'nitropack'
import type { FetchResult } from '#app'

// Todo think about removing `/api/bff/` from every request
// export type RemovePrefix<T extends string> = T extends `/api/${infer Rest}` ? Rest : T
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
//   'latestState': InternalApi['/api/bff/latest-state']['default'],
//   'users/:id': InternalApi['/api/bff/users/:id']['default'],
//   'users/me': InternalApi['/api/bff/users/me']['default'],
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
  dashboardOverview: GetReturnTypeFromServerUrl<'/api/bff/validator-dashboards/:dashboardId'>,
  productSummary: GetReturnTypeFromServerUrl<'/api/bff/product-summary'>,
  user: GetReturnTypeFromServerUrl<'/api/bff/users/me'>,
  validators: GetReturnTypeFromServerUrl<'/api/bff/validator-dashboards/:dashboardId/validators'>,
}
export type Key = keyof ReturnType
export const useFetchedData = <T extends Key | LooseAutocomplete<ServerUrl>>(key: T) => {
  // const { data } = useNuxtData <T extends Key ? ReturnType[Key] : FetchResult<T, AvailableRouterMethod<T>>>(key)
  const { data } = useNuxtData<T extends Key ? ReturnType[T] : FetchResult<T, AvailableRouterMethod<T>>>(key)
  return data
}
