import type { UseFetchOptions } from '#app'

type UseApi = typeof useApi
type UseTableUrl = Parameters<UseApi>[0]
type UseTableUrlOptions = Parameters<UseApi>[1]

// export type Query = {
//   // duty?: 'none' | 'proposal' | 'slashed' | 'sync',
//   // group_id?: number,
//   limit: number,
//   search?: string,
//   sort?: `${string}:${'asc' | 'desc'}`,
// }

export const useTable = <T extends LooseAutocomplete<ServerUrl>>(
  url: T,
  options?: UseApiOptions<any>,
) => {
  // const query = ref<Query>({
  //   limit: 25,
  //   // sort: 'asc',
  // })
  return useApi(url, options)
  // const {
  //   data,
  //   error,
  //   refresh,
  //   status,
  // } = useApi(url, {
  //   query,
  //   ...options,
  // })
  // return {
  //   data,
  //   error,
  //   query,
  //   refresh,
  //   status,
  // }
}
