import Big from 'big.js'

export const multiply = (...numbers: (number | string)[]) => {
  const product = numbers.reduce((product, number) => {
    assertIsNumber(number)
    return product = product.mul(Big(number))
  }, Big(1))
  return `${product}`
}

export const add = (...numbers: (number | string)[]) => {
  const sum = numbers.reduce((sum, number) => {
    assertIsNumber(number)
    return sum = sum.add(Big(number))
  }, Big(0))
  return `${sum}`
}

export const divide = (
  dividend: number | string,
  divisor: number | string,
) => {
  assertIsNumber(dividend, divisor)
  const quotient = new Big(dividend).div(divisor)
  return `${quotient}`
}

export const isGreaterEquals = (
  leftHandValue: number | string,
  rightHandValue: number | string,
) => {
  assertIsNumber(leftHandValue, rightHandValue)
  return Big(leftHandValue).gte(Big(rightHandValue))
}
