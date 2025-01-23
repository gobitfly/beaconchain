const Native = 'NAT' as const
type CryptoUnits = 'GWEI' | 'MAIN' | 'WEI'

type Native = typeof Native

export {
  type CryptoUnits,
  Native,
}
