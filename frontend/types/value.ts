import type { CryptoCurrency, CryptoUnits, Currency } from '~/types/currencies'

export type ValueConvertOptions = {
  sourceCurrency?: CryptoCurrency // source crypto currency - default: ETH
  sourceUnit?: CryptoUnits // source unit - default main unit (like eth)
  targetCurrency?: Currency // target currency - overrides the selected currency
  fixedDecimalCount?: number // can override the usual settings, but can't go over 2 for fiat
  minUnit?: CryptoUnits // if output should only be in higher units (for example GWEI -> then it will never go down to WEI)
  fixedUnit?: CryptoUnits // fixed output unit - overrides min unit
  addPlus?: boolean // add + sign if value is positive
}

export type ExtendedLabel = {
  label: string | number
  fullLabel?: string
}
