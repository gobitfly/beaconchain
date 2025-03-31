import {
  DateTime, type StringUnitLength,
} from 'luxon'
import type { ComposerTranslation } from 'vue-i18n'
import type { Locale } from '~/i18n/i18n.config'
import type { AgeFormat } from '~/types/settings'
import type { NumberOrString } from '~/types/value'

export const ONE_MINUTE = 60
export const ONE_HOUR = ONE_MINUTE * 60
export const ONE_DAY = ONE_HOUR * 24
export const ONE_WEEK = ONE_DAY * 7
export const ONE_YEAR = ONE_DAY * 365

export interface NumberFormatConfig {
  addPositiveSign?: boolean,
  fixed?: number,
  precision?: number,
}

export function addPlusSign(value: string, add = true): string {
  if (!add || !value || value === '0' || value.startsWith('-')) {
    return value
  }
  return `+${value}`
}

export function calculatePercent(value?: number, base?: number): number {
  if (!base) {
    return 0
  }
  return ((value ?? 0) * 100) / base
}

export function formatFiatCurrency(
  value: number | string,
  options: {
    currency?: CurrencyCodeFiat,
    locale?: Locale,
    maximumFractionDigits?: number,
    minimumFractionDigits?: number,
  } = {},
) {
  const {
    currency = 'EUR',
    locale = 'en-US',
    maximumFractionDigits,
    minimumFractionDigits,
  } = options

  return new Intl.NumberFormat(locale, {
    currency,
    maximumFractionDigits,
    minimumFractionDigits,
    style: 'currency',
  }).format(value as `${number}`)
}

/**
 * This should convert 0.2069 to 20
 */
export function formatFraction(value: NumberOrString, option?: { locale?: Locale }) {
  const {
    locale = 'en-US',
  } = option ?? {}
  const number = Number(value)
  return new Intl.NumberFormat(locale, {
    maximumFractionDigits: 0,
    minimumFractionDigits: 0,
  }).format(number * 100)
}
export function formatGoTimestamp(
  timestamp: number | string,
  compareTimestamp?: number,
  format?: AgeFormat,
  style?: StringUnitLength,
  locales?: Locale,
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

export function formatNumber(value: number | string, {
  hasRoundingIndication,
  locale = 'en-US',
  maximumFractionDigits,
  minimumFractionDigits,
  scaleBy = 0,
  signDisplay,
  useGrouping,
}: {
  hasRoundingIndication?: boolean,
  locale?: Locale,
  maximumFractionDigits?: number,
  minimumFractionDigits?: number,
  scaleBy?: number,
  signDisplay?: Intl.NumberFormatOptions['signDisplay'],
  useGrouping?: Intl.NumberFormatOptions['useGrouping'],
} = {}) {
  const [
    number,
    exponent = 0,
  ] = `${value}`.toLowerCase().split('e')
  const numberInScientificNotation = `${number}e${Number(exponent) + scaleBy}`
  const numberInScientificNotationAbsolute = Number(numberInScientificNotation)
  const isPositive = numberInScientificNotationAbsolute > 0
  const isRoundedToZero = numberInScientificNotationAbsolute < Number(`1e-${maximumFractionDigits || 1}`)
  const shouldShowRoundingIndication = hasRoundingIndication && isPositive && isRoundedToZero
  const formattedValue = new Intl.NumberFormat(locale, {
    maximumFractionDigits,
    minimumFractionDigits,
    roundingMode: shouldShowRoundingIndication
      ? 'expand'
      : 'halfExpand',
    signDisplay,
    useGrouping,
  }).format(numberInScientificNotation as `${number}`)
  return `${shouldShowRoundingIndication ? '<' : ''}${formattedValue}`
}

/**
 * Format number | string (fraction or number) to percent.
 *
 * @example 0.12346 to 12.346%
 * @example (isFraction: false) 98 to 98%
 *
 */
export function formatPercent(value: NumberOrString, option?: {
  isFraction?: boolean,
  locale?: Locale,
  maximumFractionDigits?: number,
  minimumFractionDigits?: number,
}) {
  const {
    isFraction = true,
    locale = 'en-US',
    maximumFractionDigits,
    minimumFractionDigits,
  } = option ?? {}
  const number = isFraction ? Number(value) * 100 : Number(value)
  return new Intl.NumberFormat(locale, {
    maximumFractionDigits,
    minimumFractionDigits,
    style: 'unit',
    unit: 'percent',
  }).format(number)
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

/**
 * This should convert 20 to 0.2
 */
export function formatToFraction(value: NumberOrString, option?: { locale?: Locale }) {
  const {
    locale = 'en-US',
  } = option ?? {}
  const number = Number(value ?? 0)
  return new Intl.NumberFormat(locale, {
    // maximumFractionDigits: 0,
    // minimumFractionDigits: 0,
  }).format(number / 100)
}

export function formatTs(
  ts?: number,
  timestamp?: number,
  format: AgeFormat = 'relative',
  style: StringUnitLength = 'narrow',
  locales: Locale = 'en-US',
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
  locales: Locale,
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

export function formatTsToTime(ts: number, locales: Locale): string {
  const options: Intl.DateTimeFormatOptions = {
    hour: 'numeric',
    minute: 'numeric',
  }
  const date = new Date(ts * 1000)
  return date.toLocaleTimeString(locales, options)
}

export function nZeros(count: number): string {
  return count > 0
    ? Array.from(Array(count))
        .map(() => '0')
        .join('')
    : ''
}

export function withCurrency(value: string, currency: string): string {
  return `${value} ${currency}`
}

function formatTsToRelative(
  targetTimestamp?: number,
  baseTimestamp?: number,
  style: StringUnitLength = 'narrow',
  locales: Locale = 'en-US',
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

export const formatValue = (value: string, {
  from = 'wei',
  maximumFractionDigits = 0,
  minimumFractionDigits = 0,
  to,
  useGrouping = false,
}: {
  from?: CryptoUnit,
  maximumFractionDigits?: number,
  minimumFractionDigits?: number,
  to: CryptoUnit,
  useGrouping?: boolean,
}) => {
  const scaleBy = unitFactorCrypto[from] - unitFactorCrypto[to]
  return formatNumber(value, {
    maximumFractionDigits,
    minimumFractionDigits,
    scaleBy,
    useGrouping,
  })
}
