const FiatCurrencies = ['AUD', 'CAD', 'CNY', 'EUR', 'GBP', 'JPY', 'USD', 'RUB'] as const
type FiatCurrency = typeof FiatCurrencies[number]
const CryptoCurrencies = ['ETH', 'GNO', 'DAI', 'xDAI'] as const
type CryptoCurrency = typeof CryptoCurrencies[number]
const Native = 'NAT' as const
type Native = typeof Native
type Currency = FiatCurrency | CryptoCurrency | Native

type CryptoUnits = 'MAIN' | 'GWEI' | 'WEI'

export { type Currency, type CryptoUnits, type CryptoCurrency, type FiatCurrency, CryptoCurrencies, FiatCurrencies, Native }
