import type { ComposerTranslation } from 'vue-i18n'
import type { Locale } from '~/i18n/i18n.config'
import type { NumberOrString } from '~/types/value'

export const ONE_MINUTE = 60
export const ONE_HOUR = ONE_MINUTE * 60
export const ONE_DAY = ONE_HOUR * 24
export const ONE_WEEK = ONE_DAY * 7
export const ONE_YEAR = ONE_DAY * 365

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
