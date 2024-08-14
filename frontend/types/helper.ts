/**
 * Get all possible key paths of an object
 * without arrays
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
 * type Paths = KeyPaths<Person> // "age" | "address.street"
 *
 * @warning
 *
 * Arrays are removed
 */
type KeyPaths<T> = {
  [K in keyof T]: K extends string
    ? T[K] extends object
      ? T[K] extends Array<any>
        ? never // remove Arrays
        : `${K}.${KeyPaths<T[K]>}`
      : K
    : never;
}[keyof T]

export type { KeyPaths }
