import type { Locale } from '~/i18n/i18n.config.ts'

export function formatFiatCurrency(
  value: number | string,
  options: {
    currency?: CurrencyCodeFiat,
    locale?: Locale,
    maximumFractionDigits?: number,
    minimumFractionDigits?: number,
    trailingZeroDisplay?: 'auto' | 'stripIfInteger',
  } = {},
) {
  const {
    currency = 'EUR',
    locale = 'en-US',
    maximumFractionDigits,
    minimumFractionDigits,
    trailingZeroDisplay = 'auto',
  } = options

  return new Intl.NumberFormat(locale, {
    currency,
    maximumFractionDigits,
    minimumFractionDigits,
    style: 'currency',
    trailingZeroDisplay,
  }).format(value as `${number}`)
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
