import type { ApiPagingResponse } from '~/types/api/common'
/**
 * This function wraps each element in an `ApiPagingResponse` interface with a unique id field `wrapped_identifier`.
 * @param response The api response to wrap
 * @param getIdFromRow A function that takes a row and builds a unique identifier
 * @returns The wrapped response
 */
export function wrapWithIdentifier<T>(response: ApiPagingResponse<T> | undefined, ...keys: (keyof T)[]) {
  if (!response) {
    return
  }
  return {
    data: response.data.map(row => ({
      ...row,
      wrapped_identifier: keys.map(key => row[key]).join('-'),
    })),
    paging: response.paging,
  }
}
