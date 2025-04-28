import type {
  API_PATH,
  API_PATH_RESPONSE,
  PathValues,
} from '~/types/customFetch'

export type Query = {
  /** Return data after this timestamp. */
  after_ts?: string,
  /**
  * Aggregation type to get data for.
  * @default "hourly"
  */
  aggregation?: 'daily' | 'epoch' | 'hourly' | 'weekly',
  /** Return data before this timestamp. */
  before_ts?: string,
  cursor?: string,
  efficiency_type?: 'all' | 'attestation' | 'proposal' | 'sync',
  group_id?: number,
  /** Provide a comma separated list of group IDs to filter the results by. If omitted, all groups will be included. */
  group_ids?: string,
  limit?: number,
  modes?: 'rocket_pool',
  networks?: string,
  search?: string,
  sort?: SortParameter,
}
export type SortParameter = `${string}:${'asc' | 'desc'}`
export const useTable = <T extends API_PATH>(key: string, {
  apiPath,
  pathValues,
}: {
  apiPath: T,
  pathValues?: PathValues,
}) => {
  const { fetch } = useCustomFetch()

  const query = ref<Query>({
    limit: 25,
  })
  const {
    data,
    error,
    refresh,
    status,
  } = useAsyncData(key, () => fetch<API_PATH_RESPONSE[T]>(apiPath, {
    query: query.value,
  },
  pathValues,
  ), {
    watch: [ query.value ],
  })

  const isLoading = computed(() => status.value === 'pending')

  return {
    data,
    error,
    isLoading,
    query,
    refresh,
    // updateQueryParameters,
  }
}
