import {
  CryptoCurrencies,
  type CryptoCurrency,
  type Currency,
  FiatCurrencies,
  type FiatCurrency,
} from '~/types/currencies'

const isFiat = (value?: Currency) =>
  !!value && FiatCurrencies.includes(value as FiatCurrency)
const isCrypto = (value?: Currency) =>
  !!value && CryptoCurrencies.includes(value as CryptoCurrency)
const isNative = (value?: Currency) => value === 'NAT'

export {
  isCrypto, isFiat, isNative,
}
