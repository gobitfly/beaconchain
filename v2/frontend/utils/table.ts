import type { DataTableSortEvent } from 'primevue/datatable'
import type { TableQueryParams, Cursor } from '~/types/datatable'

export const setQueryPageSize = (limit: number, query?: TableQueryParams): TableQueryParams => {
  return { ...query, limit }
}

export const setQueryCursor = (cursor: Cursor, query?: TableQueryParams): TableQueryParams => {
  return { ...query, cursor }
}

export const setQuerySearch = (search?: string, query?: TableQueryParams): TableQueryParams => {
  return { ...query, search }
}

export const setQuerySort = (sort?: DataTableSortEvent, query?: TableQueryParams): TableQueryParams => {
  query = query || {}
  if (!sort?.multiSortMeta?.length) {
    if (query) {
      delete query.sort
      delete query.order
    }
  } else {
    if (!query) {
      query = {}
    }
    query.sort = sort?.multiSortMeta.map((obj) => {
      return `${obj.field}:${obj.order === -1 ? 'asc' : 'desc'}`
    }).join(',')
  }
  return query
}
