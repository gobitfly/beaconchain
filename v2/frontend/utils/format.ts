import { commify } from '@ethersproject/units'
import { round } from 'lodash-es'
export interface NumberFormatConfig {
  precision?: number
  fixed?:number
  addPositiveSign?: boolean
}

export function formatPercent (percent?: number, config?: NumberFormatConfig):string {
  if (percent === undefined) {
    return ''
  }
  const { precision, fixed, addPositiveSign } = { ...{ precision: 2, fixed: 2, addPositiveSign: false }, ...config }
  let result = percent
  if (precision !== undefined) {
    result = round(result, precision)
  }
  const label = fixed !== undefined ? `${result.toFixed(fixed)}%` : `${result}%`
  if (fixed !== undefined) {
    return `${result.toFixed(fixed)}%`
  }

  return addPositiveSign ? addPlusSign(label) : label
}

export function calculatePercent (value?: number, base?: number):number {
  if (!base) {
    return 0
  }
  return (value ?? 0) * 100 / base
}

export function formatAndCalculatePercent (value?: number, base?: number, config?: NumberFormatConfig):string {
  if (!base) {
    return ''
  }
  return formatPercent(calculatePercent(value, base), config)
}

export function formatNumber (value?: number):string {
  return value?.toLocaleString('en-US') ?? ''
}

export function addPlusSign (value: string, add = true): string {
  if (!add || !value || value === '0' || value.startsWith('-')) {
    return value
  }
  return `+${value}`
}

export function withCurrency (value: string, currency: string):string {
  return `${value} ${currency}`
}

export function nZeros (count: number):string {
  return count > 0 ? Array.from(Array(count)).map(() => '0').join('') : ''
}

export function commmifyLeft (value: string):string {
  const formatted = commify(value)
  const i = formatted.lastIndexOf('.0')
  if (i >= 0 && i === formatted.length - 2) {
    return formatted.substring(0, formatted.length - 2)
  }
  return formatted
}

export function trim (value:string, maxDecimalCount: number, minDecimalCount?: number):string {
  minDecimalCount = minDecimalCount === undefined ? maxDecimalCount : Math.min(minDecimalCount, maxDecimalCount)
  const split = value.split('.')
  let dec = (split[1] ?? '').substring(0, maxDecimalCount)
  while (dec.length < minDecimalCount) {
    dec += '0'
  }
  if (split[0] === '0' && (!dec || parseInt(dec) === 0)) {
    if (maxDecimalCount === 0) {
      return '<1'
    }
    return `<0.${nZeros(minDecimalCount - 1)}1`
  }
  const left = commmifyLeft(split[0])
  if (!dec?.length) {
    return left
  }
  return `${left}.${dec}`
}

export function formatTs (ts: number): string {
  const options: Intl.DateTimeFormatOptions = {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  }
  return new Date(ts * 1000).toLocaleDateString(undefined, options)
}
