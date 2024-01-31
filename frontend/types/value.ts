import type { CryptoCurrency, CryptoUnits, Currency } from '~/types/currencies'

export type ValueConvertOptions = {
  sourceCurrency?: CryptoCurrency
  targetCurrency?: Currency
  fixedDecimalCount?: number
  minUnit?: CryptoUnits
}
