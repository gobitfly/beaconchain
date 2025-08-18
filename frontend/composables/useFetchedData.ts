import type {
  AvailableRouterMethod,
  InternalApi,
  NitroFetchRequest,
} from 'nitropack'
import type { FetchResult } from '#app'

// Todo think about removing `/api/bff/` from every request
// export type RemovePrefix<T extends string> = T extends `/api/${infer Rest}` ? Rest : T
export type ServerUrl = Extract<NitroFetchRequest, `/api/${string}`>
/**
     * Create a Key in {@link useFetchedData} first and map the ResponseType
     * (from our backend route) as this can not be inferred automatically via nuxt
     * at the moment 😪
     *
     * Might be solved in nuxt 4 (via useFetch factory function) or through newer
     * nitro version
*/
// type GetReturnTypeFromServerUrl<T extends ServerUrl> = FetchResult<T, AvailableRouterMethod<T>>
// only use response types for `GET` requests
// default -> server/api/bff/something.ts
// get -> server/api/bff/something.get.ts
export type GetReturnTypeFromServerUrl<T extends ServerUrl> = InternalApi[T] extends { default: infer R }
  ? R
  : InternalApi[T] extends { get: infer S }
    ? S
    : never

/**
 * This enables `type inference` for useFetchedData for `endpoint aliases`
 */
type ReturnType = {
  dashboardOverview: GetReturnTypeFromServerUrl<'/api/bff/validator-dashboards/:dashboardId'>,
  privateDashboards: GetReturnTypeFromServerUrl<'/api/bff/users/me/dashboards'>,
  productSummary: GetReturnTypeFromServerUrl<'/api/bff/product-summary'>,
  user: GetReturnTypeFromServerUrl<'/api/bff/users/me'>,
  validators: GetReturnTypeFromServerUrl<'/api/bff/validator-dashboards/:dashboardId/validators'>,
}
export type Key = keyof ReturnType
export const useFetchedData = <T extends Key | LooseAutocomplete<ServerUrl>>(key: T) => {
  const { data } = useNuxtData<T extends Key ? ReturnType[T] : FetchResult<T, AvailableRouterMethod<T>>>(key)
  return data
}
