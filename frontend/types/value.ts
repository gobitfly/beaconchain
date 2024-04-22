import type { BigNumber } from '@ethersproject/bignumber'
import type { CryptoCurrency, CryptoUnits, Currency } from '~/types/currencies'

export type ValueConvertOptions = {
  sourceCurrency?: CryptoCurrency // source crypto currency - default: ETH
  sourceUnit?: CryptoUnits // source unit - default main unit (like eth)
  targetCurrency?: Currency // target currency - overrides the selected currency
  fixedDecimalCount?: number // can override the usual settings, but can't go over 2 for fiat
  minUnit?: CryptoUnits // if output should only be in higher units (for example GWEI -> then it will never go down to WEI)
  minUnitDecimalCount?: number // decimal count to check for value while unit conversion - defaults to max decimal count
  fixedUnit?: CryptoUnits // fixed output unit - overrides min unit
  addPlus?: boolean // add + sign if value is positive
}

export type NumberOrString = number | string

export type ExtendedLabel = {
  label: NumberOrString
  fullLabel?: string
}

export const TimeFrames = ['last_24h', 'last_7d', 'last_30d', 'all_time'] as const
export type TimeFrame = typeof TimeFrames[number]

export type VaiToValue = (wei?: string | BigNumber, options?: ValueConvertOptions) => ExtendedLabel
