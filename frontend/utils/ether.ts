import { BigNumber } from '@ethersproject/bignumber'

export const OneGwei = BigNumber.from('1000000000')
export const OneEther = BigNumber.from('1000000000000000000')

export function lessThanGwei (value: BigNumber, decimals: number = 0): boolean {
  return value.lt(OneGwei.div(Math.pow(10, decimals)))
}

export function lessThanEth (value: BigNumber, decimals: number = 0): boolean {
  return value.lt(OneEther.div(Math.pow(10, decimals)))
}
