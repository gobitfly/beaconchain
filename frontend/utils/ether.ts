import { formatEther, commify } from '@ethersproject/units'
import { type BigNumberish, BigNumber } from '@ethersproject/bignumber'
import { type NumberFormatConfig, addPlusSign } from './format'

export const OneGwei = BigNumber.from('1000000000')
export const OneEther = BigNumber.from('1000000000000000000')

export function lessThenGwei (value: BigNumber, decimals: number = 0): boolean {
  return value.lt(OneGwei.div(Math.pow(10, decimals)))
}

export function lessThenEth (value: BigNumber, decimals: number = 0): boolean {
  return value.lt(OneEther.div(Math.pow(10, decimals)))
}

export function formatEth (eth: string, config?: NumberFormatConfig): string {
  if (!eth) {
    return ''
  }
  const { precision, fixed, addPositiveSign }: NumberFormatConfig = { precision: 5, fixed: 5, addPositiveSign: false, ...config }
  const split = eth.split('.')
  let dec = (split[1] ?? '').substring(0, precision)
  if (fixed) {
    while (dec.length < fixed) {
      dec += '0'
    }
  }
  const label = commify(dec.length ? `${split[0]}.${dec}` : split[0])
  return `${addPositiveSign ? addPlusSign(label) : label} ETH`
}

export function formatWeiToEth (wei: BigNumberish, config: NumberFormatConfig = {}): string {
  if (!wei) {
    return ''
  }
  return formatEth(formatEther(wei), config)
}
