import Big from 'big.js'

export const multiplyBigNumbers = (...numbers: (number | string)[]) => {
  const product = numbers.reduce((product, number) => {
    assertIsNumber(number)
    return product = product.mul(Big(number))
  }, Big(1))
  return `${product}`
}

export const addBigNumbers = (...numbers: (number | string)[]) => {
  const sum = numbers.reduce((sum, number) => {
    assertIsNumber(number)
    return sum = sum.add(Big(number))
  }, Big(0))
  return `${sum}`
}

export const divideBigNumbers = (
  dividend: number | string,
  divisor: number | string,
) => {
  assertIsNumber(dividend, divisor)
  const quotient = new Big(dividend).div(divisor)
  return `${quotient}`
}
