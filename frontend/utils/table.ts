import type { DataTableSortEvent } from 'primevue/datatable'
import type {
  Cursor, TableQueryParams,
} from '~/types/datatable'

export const getQueryWithPageSize = (
  limit: number,
  query?: TableQueryParams,
): TableQueryParams => {
  return {
    ...query,
    limit,
  }
}

export const getQueryWithCursor = (
  cursor: Cursor,
  query?: TableQueryParams,
): TableQueryParams => {
  return {
    ...query,
    cursor,
  }
}

export const getQueryWithSearch = (
  search?: string,
  query?: TableQueryParams,
): TableQueryParams => {
  return {
    ...query,
    search,
  }
}

export const getSortOrder = (dir?: null | number) =>
  dir === -1 ? 'asc' : 'desc'

export const getQueryWithSort = (
  sort?: DataTableSortEvent,
  query?: TableQueryParams,
): TableQueryParams => {
  query = query || {}
  if (sort?.multiSortMeta?.length) {
    if (!query) {
      query = {}
    }
    query = {
      ...query,
      sort: sort?.multiSortMeta
        .map((obj) => {
          return `${obj.field}:${getSortOrder(obj.order)}`
        })
        .join(','),
    }
  }
  else if (sort?.sortField && sort?.sortOrder) {
    if (!query) {
      query = {}
    }
    query = {
      ...query,
      sort: `${sort.sortField}:${getSortOrder(sort?.sortOrder)}`,
    }
  }
  else if (query) {
    delete query.sort
    delete query.order
  }
  return query
}
