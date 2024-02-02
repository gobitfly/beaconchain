const FiatCurrencies = ['AUD', 'CAD', 'CNY', 'EUR', 'GBP', 'JPY', 'USD'] as const
type FiatCurrency = typeof FiatCurrencies[number]
const CryptoCurrenies = ['ETH', 'GNO', 'DAI', 'XDAI'] as const
type CryptoCurrency = typeof CryptoCurrenies[number]
const Native = 'NAT' as const
type Native = typeof Native
type Currency = FiatCurrency | CryptoCurrency | Native

type CryptoUnits = 'MAIN' | 'GWEI' | 'WEI'

export { type Currency, type CryptoUnits, type CryptoCurrency, type FiatCurrency, CryptoCurrenies, FiatCurrencies, Native }
