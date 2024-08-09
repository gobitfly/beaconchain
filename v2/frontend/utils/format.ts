import { commify } from '@ethersproject/units'
import {
  DateTime, type StringUnitLength,
} from 'luxon'
import { type ComposerTranslation } from 'vue-i18n'
import type { AgeFormat } from '~/types/settings'
import {
  type ChainIDs, epochToTs, slotToTs,
} from '~/types/network'

export const ONE_MINUTE = 60
export const ONE_HOUR = ONE_MINUTE * 60
export const ONE_DAY = ONE_HOUR * 24
export const ONE_WEEK = ONE_DAY * 7
export const ONE_YEAR = ONE_DAY * 365

export interface NumberFormatConfig {
  addPositiveSign?: boolean
  fixed?: number
  precision?: number
}

export function formatPercent(
  percent?: number,
  config?: NumberFormatConfig,
): string {
  if (percent === undefined) {
    return ''
  }
  const {
    addPositiveSign, fixed, precision,
  } = {
    ...{
      addPositiveSign: false,
      fixed: 2,
      precision: 2,
    },
    ...config,
  }
  let result = trim(percent, precision, fixed)
  if (addPositiveSign) {
    result = addPlusSign(result)
  }
  return `${result}%`
}

export function calculatePercent(value?: number, base?: number): number {
  if (!base) {
    return 0
  }
  return ((value ?? 0) * 100) / base
}

export function formatAndCalculatePercent(
  value?: number,
  base?: number,
  config?: NumberFormatConfig,
): string {
  if (!base) {
    return ''
  }
  return formatPercent(calculatePercent(value, base), config)
}

export function formatNumber(value?: number): string {
  return value?.toLocaleString('en-US') ?? ''
}

export function addPlusSign(value: string, add = true): string {
  if (!add || !value || value === '0' || value.startsWith('-')) {
    return value
  }
  return `+${value}`
}

export function withCurrency(value: string, currency: string): string {
  return `${value} ${currency}`
}

export function nZeros(count: number): string {
  return count > 0
    ? Array.from(Array(count))
      .map(() => '0')
      .join('')
    : ''
}

export function commmifyLeft(value: string): string {
  const formatted = commify(value)
  const i = formatted.lastIndexOf('.0')
  if (i >= 0 && i === formatted.length - 2) {
    return formatted.substring(0, formatted.length - 2)
  }
  return formatted
}

export function trim(
  value: number | string,
  maxDecimalCount: number,
  minDecimalCount?: number,
): string {
  if (typeof value !== 'string') {
    value = `${value}`
  }
  minDecimalCount
    = minDecimalCount === undefined
      ? maxDecimalCount
      : Math.min(minDecimalCount, maxDecimalCount)
  const split = value.split('.')
  let dec = split[1] ?? ''
  const hasTinyValue = !!dec && REGEXP_HAS_NUMBERS.test(dec)
  dec = dec.substring(0, maxDecimalCount)
  while (dec.length < minDecimalCount) {
    dec += '0'
  }
  if (split[0] === '0' && (!dec || parseInt(dec) === 0) && hasTinyValue) {
    if (maxDecimalCount === 0) {
      return '<1'
    }
    return `<0.${nZeros(maxDecimalCount - 1)}1`
  }
  const left = commmifyLeft(split[0])
  if (!dec?.length) {
    return left
  }
  return `${left}.${dec}`
}

function formatTs(
  ts?: number,
  timestamp?: number,
  format: AgeFormat = 'relative',
  style: StringUnitLength = 'narrow',
  locales: string = 'en-US',
  withTime = true,
) {
  if (ts === undefined) {
    return undefined
  }

  if (format === 'relative') {
    return formatTsToRelative(ts * 1000, timestamp, style, locales)
  }
  else {
    return formatTsToAbsolute(ts, locales, withTime)
  }
}

export function formatTsToAbsolute(
  ts: number,
  locales: string,
  includeTime?: boolean,
): string {
  const timeOptions: Intl.DateTimeFormatOptions = includeTime
    ? {
        hour: 'numeric',
        minute: 'numeric',
      }
    : {}
  const options: Intl.DateTimeFormatOptions = {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    ...timeOptions,
  }
  const date = new Date(ts * 1000)
  return includeTime
    ? date.toLocaleString(locales, options)
    : date.toLocaleDateString(locales, options)
}

export function formatTsToTime(ts: number, locales: string): string {
  const options: Intl.DateTimeFormatOptions = {
    hour: 'numeric',
    minute: 'numeric',
  }
  const date = new Date(ts * 1000)
  return date.toLocaleTimeString(locales, options)
}

function formatTsToRelative(
  targetTimestamp?: number,
  baseTimestamp?: number,
  style: StringUnitLength = 'narrow',
  locales: string = 'en-US',
): null | string | undefined {
  if (!targetTimestamp) {
    return undefined
  }

  const date = baseTimestamp
    ? DateTime.fromMillis(baseTimestamp)
    : DateTime.now()
  return DateTime.fromMillis(targetTimestamp)
    .setLocale(locales)
    .toRelative({
      base: date,
      style,
    })
}

export function formatGoTimestamp(
  timestamp: number | string,
  compareTimestamp?: number,
  format?: AgeFormat,
  style?: StringUnitLength,
  locales?: string,
  withTime?: boolean,
) {
  if (typeof timestamp === 'number') {
    timestamp *= 1000
  }
  const dateTime = new Date(timestamp).getTime()
  return formatTs(
    dateTime / 1000,
    compareTimestamp,
    format,
    style,
    locales,
    withTime,
  )
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `formatEpochToDateTime(currentNetwork.value, ...)`
 * you should rather use `formatEpochToDateTime(...)` from `useFormat.ts`.
 */
export function formatEpochToDateTime(
  chainId: ChainIDs,
  epoch: number,
  timestamp?: number,
  format?: AgeFormat,
  style?: StringUnitLength,
  locales?: string,
  withTime?: boolean,
): null | string | undefined {
  return formatTs(
    epochToTs(chainId, epoch),
    timestamp,
    format,
    style,
    locales,
    withTime,
  )
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `formatSlotToDateTime(currentNetwork.value, ...)`
 * you should rather use `formatSlotToDateTime(...)` from `useFormat.ts`.
 */
export function formatSlotToDateTime(
  chainId: ChainIDs,
  slot: number,
  timestamp?: number,
  format?: AgeFormat,
  style?: StringUnitLength,
  locales?: string,
  withTime?: boolean,
): null | string | undefined {
  return formatTs(
    slotToTs(chainId, slot),
    timestamp,
    format,
    style,
    locales,
    withTime,
  )
}

/**
 * Should be used only when you work with a network different from the current one.
 * Wherever you would write `formatEpochToDate(currentNetwork.value, ...)` you
 * should rather use `formatEpochToDate(...)` from `useFormat.ts`.
 */
export function formatEpochToDate(
  chainId: ChainIDs,
  epoch: number,
  locales: string,
): null | string | undefined {
  return formatEpochToDateTime(
    chainId,
    epoch,
    undefined,
    'absolute',
    undefined,
    locales,
    false,
  )
}

export function formattedNumberToHtml(value?: string): string | undefined {
  return value?.split(',').join('<span class=\'comma\' />')
}

export function formatTimeDuration(
  seconds: number | undefined,
  t: ComposerTranslation,
): string | undefined {
  if (seconds === undefined) {
    return undefined
  }

  let translationId = 'time_duration.years'
  let divider = ONE_YEAR

  if (seconds < ONE_MINUTE) {
    translationId = 'time_duration.seconds'
    divider = 1
  }
  else if (seconds < ONE_HOUR) {
    translationId = 'time_duration.minutes'
    divider = ONE_MINUTE
  }
  else if (seconds < ONE_DAY) {
    translationId = 'time_duration.hours'
    divider = ONE_HOUR
  }
  else if (seconds < ONE_YEAR) {
    translationId = 'time_duration.days'
    divider = ONE_DAY
  }

  const amount = Math.floor(seconds / divider)

  return t(translationId, { amount }, amount === 1 ? 1 : 2)
}

export function formatNanoSecondDuration(
  nano: number | undefined,
  t: ComposerTranslation,
): string | undefined {
  if (nano === undefined) {
    return undefined
  }
  const seconds = Math.floor(nano / 1000000000)
  return formatTimeDuration(seconds, t)
}

export function formatFiat(
  value: number,
  currency: string,
  locales: string,
  minimumFractionDigits?: number,
  maximumFractionDigits?: number,
) {
  const formatter = new Intl.NumberFormat(locales, {
    currency,
    maximumFractionDigits,
    minimumFractionDigits,
    style: 'currency',
  })

  return formatter.format(value)
}

export const formatPremiumProductPrice = (
  t: ComposerTranslation,
  price: number,
  digits?: number,
) => {
  return formatFiat(
    price,
    'EUR',
    t('locales.currency'),
    digits ?? 2,
    digits ?? 2,
  )
}
