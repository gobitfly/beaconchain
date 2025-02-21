import type { DataTableSortEvent } from 'primevue/datatable'
import type {
  Cursor, TableQueryParams,
} from '~/types/datatable'

export function useTableQuery(
  initialQuery?: TableQueryParams,
  initialPageSize = 10,
) {
  const cursor = ref<Cursor>()
  const pageSize = ref<number>(initialPageSize)
  const storedQuery = ref<TableQueryParams | undefined>()

  const {
    bounce: setQuery,
    temp: pendingQuery,
    value: query,
  } = useDebounceValue<TableQueryParams | undefined>(initialQuery, 500)

  const onSort = (sort: DataTableSortEvent) => {
    setQuery(getQueryWithSort(sort, pendingQuery.value))
  }

  const setCursor = (value: Cursor) => {
    cursor.value = value
    setQuery(getQueryWithCursor(value, pendingQuery.value))
  }

  const setPageSize = (value: number) => {
    pageSize.value = value
    setQuery(getQueryWithPageSize(value, pendingQuery.value))
  }

  const setSearch = (value?: string) => {
    setQuery(getQueryWithSearch(value, pendingQuery.value))
  }

  const setStoredQuery = (q?: TableQueryParams) => {
    storedQuery.value = q
  }

  const isStoredQuery = (q?: TableQueryParams) => {
    return JSON.stringify(storedQuery.value) === JSON.stringify(q)
  }

  return {
    cursor,
    isStoredQuery,
    onSort,
    pageSize,
    pendingQuery,
    query,
    setCursor,
    setPageSize,
    setSearch,
    setStoredQuery,
  }
}
