import { BigNumber } from '@ethersproject/bignumber'

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
