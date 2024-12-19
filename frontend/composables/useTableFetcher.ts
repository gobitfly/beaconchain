import type {
  ApiPagingResponse, Paging,
} from '~/types/api/common'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'

export const useTableFetcher = <T extends object>(
  fetchFunc: (
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) => Promise<ApiPagingResponse<T> | undefined>,
  defaultQuery: TableQueryParams,
  errFunc: (e: any) => void,
) => {
  const data = ref<T[]>()
  const paging = ref<Paging>()
  const isLoading = ref(false)
  const query = ref<TableQueryParams>(defaultQuery)

  const refresh = (
    dashboardKey: DashboardKey,
    inputQuery: TableQueryParams,
  ) => {
    isLoading.value = true
    query.value = inputQuery
    fetchFunc(dashboardKey, inputQuery)
      .then((fetchedResult) => {
        data.value = fetchedResult?.data
        paging.value = fetchedResult?.paging
      })
      .catch((e) => {
        errFunc(e)
      })
      .finally(() => {
        isLoading.value = false
      })
  }

  return {
    data,
    isLoading,
    paging,
    query,
    refresh,
  }
}
