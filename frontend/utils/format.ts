import { commify } from '@ethersproject/units'
import { DateTime, type StringUnitLength } from 'luxon'
import type { AgeFormat } from '~/types/settings'

const { epochToTs } = useNetwork()

const REGEXP_HAS_NUMBERS = /^(?!0+$)\d+$/

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
  let result = trim(percent, precision, fixed)
  if (addPositiveSign) {
    result = addPlusSign(result)
  }
  return `${result}%`
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

export function trim (value:string | number, maxDecimalCount: number, minDecimalCount?: number):string {
  if (typeof value !== 'string') {
    value = `${value}`
  }
  minDecimalCount = minDecimalCount === undefined ? maxDecimalCount : Math.min(minDecimalCount, maxDecimalCount)
  const split = value.split('.')
  let dec = (split[1] ?? '')
  const hasTinyValue = !!dec && REGEXP_HAS_NUMBERS.test(dec)
  dec = dec.substring(0, maxDecimalCount)
  while (dec.length < minDecimalCount) {
    dec += '0'
  }
  if (split[0] === '0' && (!dec || parseInt(dec) === 0) && hasTinyValue) {
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

export function formatTs (ts: number, locales: string, includeTime?: boolean): string {
  const timeOptions: Intl.DateTimeFormatOptions = includeTime
    ? {
        hour: 'numeric',
        minute: 'numeric'
      }
    : {}
  const options: Intl.DateTimeFormatOptions = {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    ...timeOptions
  }
  const date = new Date(ts * 1000)
  return includeTime ? date.toLocaleString(locales, options) : date.toLocaleDateString(locales, options)
}

export function formatToRelative (targetTimestamp?: number, baseTimestamp?: number, style: StringUnitLength = 'narrow', locales: string = 'en-US') {
  if (!targetTimestamp) {
    return undefined
  }
  const date = baseTimestamp ? DateTime.fromMillis(baseTimestamp) : DateTime.now()
  return DateTime.fromMillis(targetTimestamp).setLocale(locales).toRelative({ base: date, style })
}

export function formatEpochToDateTime (epoch: number, timestamp?: number, format: AgeFormat = 'relative', style: StringUnitLength = 'narrow', locales: string = 'en-US', withTime = true) {
  const ts = epochToTs(epoch)
  if (ts === undefined) {
    return undefined
  }
  if (format === 'relative') {
    return formatToRelative(ts * 1000, timestamp, style, locales)
  } else {
    return formatTs(ts, locales, withTime)
  }
}

export function formatEpochToDate (epoch: number, locales: string): string | null |undefined {
  return formatEpochToDateTime(epoch, undefined, 'absolute', undefined, locales, false)
}

export function formattedNumberToHtml (value?:string):string | undefined {
  return value?.split(',').join("<span class='comma' />")
}
