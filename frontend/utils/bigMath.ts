import { BigNumber } from '@ethersproject/bignumber'
import Big from 'big.js'
import type { ClElValue } from '~/types/api/common'

// const scaleNumber = (
//   value: bigint | number | string,
//   direction: 'down' | 'up',
// ) => {
//   assertIsNumber(value)
//   // scaling by 10 ** 18 should get enough precision
//   let exponent = ''
//   if (direction === 'up') exponent = 'e+18'
//   if (direction === 'down') exponent = 'e-18'
//   // scaling via `scienfiic notation` (string manipulation) as we also need to scale numbers with fractions
//   // which cannot accurately be represented by a `js number`, e.g. 9.00719925474099199 * 10 ** 18 (not accurate)
//   const scaledInteger = new Intl.NumberFormat('en-US', {
//     // scaling up for BigInt() calculations -> no fractions allowed
//     // afterwards scaling down for display purposes -> we want to show precise values
//     maximumFractionDigits: direction === 'up' ? 0 : 18,
//     useGrouping: false,
//   }).format(`${value}${exponent}` as `${number}`)

//   return scaledInteger
// }

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
export const isGreaterThanBigNumbers = (firstNumber: number | string, secondNumber: number | string) => {
  return Big(firstNumber).gt(Big(secondNumber))
}

const getFactor = (str?: string): number => {
  const decimals = str?.length ?? 0
  return Math.pow(10, decimals)
}

const split = (num: number) => {
  const str = `${num}`
  const split = str.split('.')
  const factor = getFactor(split[1])
  return {
    combined: split.join(''),
    factor,
  }
}

export const divideBigNumbers = (
  dividend: number | string,
  divisor: number | string,
) => {
  assertIsNumber(dividend, divisor)
  const quotient = new Big(dividend).div(divisor)
  return `${quotient}`
}

export const bigMul = (big: BigNumber, num: number): BigNumber => {
  if (!big || !num) {
    return big
  }
  const {
    combined, factor,
  } = split(num)
  return big.mul(combined).div(factor)
}

export const bigDiv = (big: BigNumber, num: number): BigNumber => {
  if (!big || !num) {
    return big
  }
  const {
    combined, factor,
  } = split(num)
  return big.mul(factor).div(combined)
}

export const convertSum = (...values: string[]): BigNumber | undefined => {
  return values?.reduce(
    (sum, newValue) => sum.add(BigNumber.from(newValue)),
    BigNumber.from('0'),
  )
}

export const totalElCl = (value: ClElValue<string>): BigNumber | undefined => {
  if (!value) {
    return
  }
  return convertSum(value.el, value.cl)
}

export const subWei = (total: string, value: string): BigNumber | undefined => {
  if (!total) {
    return
  }

  return BigNumber.from(total).sub(BigNumber.from(value ?? '0'))
}

export const totalElClNumbers = (
  value: ClElValue<number>,
) => {
  return value.el + value.cl
}
