export type Cursor = number | string | undefined
export type ColumnOrder = 'asc' | 'desc'
export type SortOrder = 1 | 0 | -1 | undefined | null

export type TableQueryParams = {
  limit?: number;
  cursor?: Cursor;
  order?: ColumnOrder;
  sort?: string;
  search?: string;
}
