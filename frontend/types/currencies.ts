const FiatCurrencies = [
  'AUD',
  'CAD',
  'CNY',
  'EUR',
  'GBP',
  'JPY',
  'USD',
  'RUB',
] as const
type FiatCurrency = (typeof FiatCurrencies)[number]
const CryptoCurrencies = [
  'ETH',
  'GNO',
  'DAI',
  'xDAI',
] as const
type CryptoCurrency = (typeof CryptoCurrencies)[number]
const Native = 'NAT' as const
type CryptoUnits = 'GWEI' | 'MAIN' | 'WEI'
type Currency = CryptoCurrency | FiatCurrency | Native

type Native = typeof Native

export {
  CryptoCurrencies,
  type CryptoCurrency,
  type CryptoUnits,
  type Currency,
  FiatCurrencies,
  type FiatCurrency,
  Native,
}
