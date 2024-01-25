import { formatEther, commify } from '@ethersproject/units'
import { type BigNumberish } from '@ethersproject/bignumber'
import { type NumberFormatConfig, addPlusSign } from './format'

export function formatEth (eth: string, { precision, fixed, addPositiveSign }: NumberFormatConfig = { precision: 5, fixed: 5, addPositiveSign: false }): string {
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
  const label = commify(dec.length ? `${split[0]}.${dec}` : split[0])
  return addPositiveSign ? addPlusSign(label) : label
}

export function formatWeiToEth (wei: BigNumberish, config: NumberFormatConfig): string {
  if (!wei) {
    return ''
  }
  return formatEth(formatEther(wei), config)
}
