const FiatCurrencyCodes = [
  'AUD',
  'CAD',
  'CNY',
  'EUR',
  'GBP',
  'JPY',
  'USD',
] as const
export type CurrencyCodeFiat = (typeof FiatCurrencyCodes)[number]
export const isFiat = (currencyCode: CurrencyCode): currencyCode is CurrencyCodeFiat =>
  FiatCurrencyCodes.includes(currencyCode as CurrencyCodeFiat)

const CryptoCurrencyCodes = [
  'ETH',
  'GNO',
  'DAI',
  'xDAI',
  'mGNO',
] as const
export type CurrencyCodeCrypto = (typeof CryptoCurrencyCodes)[number]
export const isCrypto = (currencyCode: CurrencyCode): currencyCode is CurrencyCodeCrypto =>
  CryptoCurrencyCodes.includes(currencyCode as CurrencyCodeCrypto)

export type CurrencyCode = CurrencyCodeCrypto | CurrencyCodeFiat
