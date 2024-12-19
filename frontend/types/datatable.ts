import type { Paging } from './api/common'
import type { TimeFrame } from './value'

export type ColumnOrder = 'asc' | 'desc'
export type Cursor = number | string | undefined
export type SortOrder = -1 | 0 | 1 | null | undefined

export type TablePeriodQueryParams = {
  cursor?: Cursor,
  is_mocked?: boolean,
  limit?: number,
  order?: ColumnOrder,
  period: TimeFrame,
  search?: string,
  sort?: string,
}

export type TableProps<T> = {
  data: T[] | undefined,
  isLoading: boolean,
  paging: Paging | undefined,
  query: TableQueryParams,
}

export type TableQueryParams = {
  cursor?: Cursor,
  is_mocked?: boolean,
  limit?: number,
  order?: ColumnOrder,
  search?: string,
  sort?: string,
}

export type TotalTableProps<T, E> = {
  data: T[] | undefined,
  dataTotal: E | undefined,
  isLoading: boolean,
  isLoadingTotal: boolean,
  paging: Paging | undefined,
  query: TableQueryParams,
}
