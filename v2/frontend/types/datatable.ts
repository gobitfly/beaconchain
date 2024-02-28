export type Cursor = number | string | undefined
export type ColumnOrder = 'asc' | 'desc'

export type TableQueryParams = {
  limit?: number,
  cursor?: Cursor,
  order?: ColumnOrder
  sort?: string,
  search?: string
}
