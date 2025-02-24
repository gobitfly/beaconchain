import Big from 'big.js'

export const multiplyBigNumbers = (...numbers: (number | string)[]) => {
  return numbers.reduce((product, number) => {
    assertIsNumber(number)
    return product = product.mul(Big(number))
  }, Big(1))
}

export const addBigNumbers = (...numbers: (number | string)[]) => {
  return numbers.reduce((sum, number) => {
    assertIsNumber(number)
    return sum = sum.add(Big(number))
  }, Big(0))
}

export const divideBigNumbers = (
  dividend: number | string,
  divisor: number | string,
) => {
  assertIsNumber(dividend, divisor)
  const quotient = new Big(dividend).div(divisor)
  return `${quotient}`
}
