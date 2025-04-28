export type ColumnOrder = 'asc' | 'desc'
export type Cursor = number | string | undefined

export type TableQueryParams = {
  cursor?: Cursor,
  is_mocked?: boolean,
  limit?: number,
  order?: ColumnOrder,
  search?: string,
  sort?: string,
}
