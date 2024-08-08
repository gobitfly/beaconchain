import {
  FiatCurrencies,
  type FiatCurrency,
  CryptoCurrencies,
  type CryptoCurrency,
  type Currency,
} from '~/types/currencies'

const isFiat = (value?: Currency) =>
  !!value && FiatCurrencies.includes(value as FiatCurrency)
const isCrypto = (value?: Currency) =>
  !!value && CryptoCurrencies.includes(value as CryptoCurrency)
const isNative = (value?: Currency) => value === 'NAT'

export { isFiat, isCrypto, isNative }
