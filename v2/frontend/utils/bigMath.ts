import { BigNumber } from '@ethersproject/bignumber'
import type { ClElValue } from '~/types/api/common'

const getFactor = (str?: string): number => {
  const decimals = str?.length ?? 0
  return Math.pow(10, decimals)
}

const split = (num: number) => {
  const str = `${num}`
  const split = str.split('.')
  const factor = getFactor(split[1])
  return {
    factor,
    combined: split.join('')
  }
}

export const bigMul = (big: BigNumber, num: number): BigNumber => {
  if (!big || !num) {
    return big
  }
  const { factor, combined } = split(num)
  return big.mul(combined).div(factor)
}

export const bigDiv = (big: BigNumber, num: number): BigNumber => {
  if (!big || !num) {
    return big
  }
  const { factor, combined } = split(num)
  return big.mul(factor).div(combined)
}

export const convertSum = (...values:string[]):BigNumber | undefined => {
  return values?.reduce((sum, newValue) => sum.add(BigNumber.from(newValue)), BigNumber.from('0'))
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

export const totalElClNumbers = (value: ClElValue<number>): number | undefined => {
  if (!value) {
    return
  }
  return value.el + value.cl
}
