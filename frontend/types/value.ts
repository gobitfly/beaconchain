import type { BigNumber } from '@ethersproject/bignumber'
import type {
  CryptoCurrency, CryptoUnits, Currency,
} from '~/types/currencies'

export type ValueConvertOptions = {
  addPlus?: boolean // add + sign if value is positive
  fixedDecimalCount?: number // can override the usual settings, but can't go over 2 for fiat
  fixedUnit?: CryptoUnits // fixed output unit - overrides min unit
  maxDecimalCount?: number // max decimal count
  minDecimalCount?: number // min decimal count
  minUnit?: CryptoUnits // if output should only be in higher units (e.g. GWEI -> then it will never go down to WEI)
  minUnitDecimalCount?: number // decimal count to check for value while unit conversion - defaults to max decimal count
  sourceCurrency?: CryptoCurrency // source crypto currency - default: ETH
  sourceUnit?: CryptoUnits // source unit - default main unit (like eth)
  targetCurrency?: Currency // target currency - overrides the selected currency
}

export type NumberOrString = number | string

export type ExtendedLabel = {
  fullLabel?: string
  label: NumberOrString
}

export const TimeFrames = [
  'last_24h',
  'last_7d',
  'last_30d',
  'all_time',
] as const
export type TimeFrame = (typeof TimeFrames)[number]

export type WeiToValue = (
  wei?: BigNumber | string,
  options?: ValueConvertOptions,
) => ExtendedLabel

export type CompareResult = 'equal' | 'higher' | 'lower'
