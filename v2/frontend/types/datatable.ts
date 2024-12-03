export type ColumnOrder = 'asc' | 'desc'
export type Cursor = number | string | undefined
export type SortOrder = -1 | 0 | 1 | null | undefined

export type TableQueryParams = {
  cursor?: Cursor,
  is_mocked?: boolean,
  limit?: number,
  order?: ColumnOrder,
  search?: string,
  sort?: string,
}
