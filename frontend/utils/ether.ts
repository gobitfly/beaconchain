import { formatEther, commify } from '@ethersproject/units'
import { BigNumberish } from '@ethersproject/bignumber'
import { NumberFormatConfig } from './format'

export function formatEth (eth: string, { precision = 5, fixed = 5 }: NumberFormatConfig): string {
  if (!eth) {
    return ''
  }
  const split = eth.split('.')
  let dec = (split[1] ?? '').substring(0, precision)
  if (fixed) {
    while (dec.length < fixed) {
      dec += '0'
    }
  }
  return commify(dec.length ? `${split[0]}.${dec}` : split[0])
}

export function formatWeiToEth (wei: BigNumberish, config: NumberFormatConfig): string {
  if (!wei) {
    return ''
  }
  return formatEth(formatEther(wei), config)
}
