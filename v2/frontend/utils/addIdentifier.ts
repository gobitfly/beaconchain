import type { ApiPagingResponse } from '~/types/api/common'
/**
 * Sometimes the api does not provide a `unique id` for objects in an array.
 * But in `frontend` there are `components` that need  to differentiate between the items.
 * So an identifier for each item is constructed out of a combination of the `object`.
 */
export function addIdentifier<T>(response: ApiPagingResponse<T> | undefined, ...keys: (keyof T)[]) {
  if (!response) {
    return
  }
  return {
    data: response.data.map(row => ({
      ...row,
      identifier: keys.map(key => row[key]).join('-'),
    })),
    paging: response.paging,
  }
}
