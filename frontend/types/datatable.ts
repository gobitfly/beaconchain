export type Cursor = number | string | undefined
export type ColumnOrder = 'asc' | 'desc'
export type SortOrder = -1 | 0 | 1 | null | undefined

export type TableQueryParams = {
  cursor?: Cursor,
  limit?: number,
  order?: ColumnOrder,
  search?: string,
  sort?: string,
}
