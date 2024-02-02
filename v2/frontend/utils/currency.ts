import { FiatCurrencies, type FiatCurrency, CryptoCurrenies, type CryptoCurrency, type Currency } from '~/types/currencies'

const isFiat = (value?:Currency) => !!value && FiatCurrencies.includes(value as FiatCurrency)
const isCrypto = (value?:Currency) => !!value && CryptoCurrenies.includes(value as CryptoCurrency)
const isNative = (value?:Currency) => value === 'NAT'

export { isFiat, isCrypto, isNative }
