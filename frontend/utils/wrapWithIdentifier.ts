import type { ApiPagingResponse } from '~/types/api/common'
/**
 * This function wraps each element in an `ApiPagingResponse` interface with a unique id field `wrapped_identifier`.
 * @param response The api response to wrap
 * @param getIdFromRow A function that takes a row and builds a unique identifier
 * @returns The wrapped response
 */
export function wrapWithIdentifier<T>(response: ApiPagingResponse<T> | undefined, getIdFromRow: (row: T) => string) {
  if (!response) {
    return
  }
  return {
    data: response.data.map(row => ({
      ...row,
      wrapped_identifier: getIdFromRow(row),
    })),
    paging: response.paging,
  }
}
