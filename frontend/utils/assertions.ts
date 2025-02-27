/**
 *
 * This function will throw when one of the given `(string) values` is not a number.
 *
 */
export const assertIsNumber = (...values: (number | string)[]) => {
  const message = 'Value is not a number:'
  values.forEach((value) => {
    if (typeof value === 'string' && Number.isNaN(Number(value))) {
      logError(`${message} ${value}`)
      throw new Error(`${message} ${value}`)
    }
  })
}
/**
 *
 * This function will throw when one of the given `(string) values` is zero.
 *
 */
export const assertIsNotZero = (...values: (number | string)[]) => {
  const message = 'Value cannot be zero'
  values.forEach((value) => {
    if (Number(value) === 0) {
      logError(`${message}`)
      throw new Error(`${message}`)
    }
  })
}
