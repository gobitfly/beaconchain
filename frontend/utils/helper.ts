/**
 * Get all possible key paths of an object
 * without arrays
 *
 *  * @warning
 *
 * Arrays are removed
 *
 * @example
 *
 * type Person = {
 *  age: number
 *  address: {
 *    street: string
 *  },
 *  hobbies: {
 *    outdoor: string[]
 *  }
 * }
 *
 * type Paths = GetObjectPaths<Person> // "age" | "address.street"
 *
 */
export type GetObjectPaths<T extends object> = {
  [K in keyof T]: K extends string
    ? T[K] extends object
      ? T[K] extends Array<any>
        ? never // remove Arrays
        : `${K}.${GetObjectPaths<T[K]>}`
      : K
    : never;
}[keyof T]

/**
 * Get autocompletion for T while beeing able to also add a string
 * */
export type LooseAutocomplete<T> = (string & {}) | T

/**
 * Check if an object has a specific key in a type-safe way.
 */
export function hasKey<T, K extends PropertyKey>(
  obj: T,
  key: K,
): obj is Record<K, unknown> & T {
  return typeof obj === 'object' && obj !== null && key in obj
}
