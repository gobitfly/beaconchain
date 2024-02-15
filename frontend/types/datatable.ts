export type ColumnOrder = 'asc' | 'desc'

export type TableQUeryParams = {
  limit?: number,
  cursor?: string,
  order?: ColumnOrder
  sort?: string,
  search?: string
}
