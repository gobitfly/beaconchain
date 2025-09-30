const CHAIN_ID = {
  Ethereum: 1,
  Hoodi: 560048,
} as const

export type ChainId = (typeof CHAIN_ID)[keyof typeof CHAIN_ID]

export const getNetworkShortName = (chainId: ChainId) => {
  if (chainId === CHAIN_ID.Ethereum) return 'Ethereum'
  if (chainId === CHAIN_ID.Hoodi) return 'Hoodi'
}
