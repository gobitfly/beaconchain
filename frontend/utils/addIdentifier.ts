import type { Paging } from '~/types/api/common'

/**
 * Sometimes the api does not provide a `unique id` for objects in an array.
 * But in `frontend` there are `components` that need  to differentiate between the items.
 * So an identifier for each item is constructed out of a combination of the `object`.
 */
export function addIdentifier<T>(response?: { data?: T[], paging?: Paging }, ...keys: (keyof T)[]) {
  if (!response || !response.data) {
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
