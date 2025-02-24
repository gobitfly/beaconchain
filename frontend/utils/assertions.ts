/**
 *
 * This function will throw when the given `string value` is not a number.
 * Which is necessary when passing number in `string format`
 *
 */
export const assertIsNumber = (...values: (number | string)[]) => {
  const message = 'Element is not a number:'
  values.forEach((value) => {
    if (typeof value === 'string' && Number.isNaN(Number(value))) {
      logError(`${message} ${value}`)
      throw new Error(`${message} ${value}`)
    }
  })
}
