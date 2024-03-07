/* This file sets types and fundamental data about different networks.

   To use it, you need to write first:
     import { ChainIDs, ChainInfo } from '~/types/networks'

   First, the file defines identifiers equal to the chain IDs of the networks.
   In your code, you write ChainIDs.Ethereum whenever you want to represent the main Ethereum network,
   or ChainIDs.Sepolia for the Sepolia testnet and so on.
   Those constants are integers but for a safer code you should define your constants, fields, variables
   and parameters as the type ChainIDs.

   The most important feature of this file is to provide a mapping between those chain IDs and
   information about the networks.
   For example, when your variable myNetwork is equal to 10200 (namely ChainIDs.Chiado, a testnet of Gnosis) :
   * ChainInfo[myNetwork].path  is the beginning of the path used to address this network in API endpoints.
   * ChainInfo[myNetwork].mainNet  is equal to the chain ID of the mainnet of Gnosis (100).
     So, to check whether your network is a testnet, you can do  myNetwork != ChainInfo[myNetwork].mainNet
     or simply !isMainNet(myNetwork) whose implementation does the same test (first, add the function to your import list)
   * ChainInfo[myNetwork].elCurrency is equal to 'xDAI' whereas ChainInfo[myNetwork].clCurrency is 'GNO'

   To check whether a network is a L1, you can do myNetwork === ChainInfo[myNetwork].L1
   or simply isL1(myNetwork) whose implementation does the same test (first, add the function to your import list)
*/

import type { CryptoCurrency } from '~/types/currencies'

export type NetworkFamily = 'Any' | 'Ethereum' | 'Arbitrum' | 'Optimism' | 'Base' | 'Gnosis'

export enum ChainIDs {
  Any = 0, // to organize data internally (example of use: some ahead-results in the search bar belong to all networks)

  Ethereum = 1,
  Holesky = 17000,
  Sepolia = 11155111,

  ArbitrumOneEthereum = 42161,
  ArbitrumNovaEthereum= 42170,
  ArbitrumOneSepolia = 421614,

  OptimismEthereum = 10,
  OptimismSepolia = 11155420,

  BaseEthereum = 8453,
  BaseSepolia = 84532,

  Gnosis = 100,
  Chiado = 10200
}

// TODO: request it from the API
export function getListOfImplementedChainIDs (sortByPriority : boolean) : ChainIDs[] {
  const list = [ChainIDs.Ethereum, ChainIDs.ArbitrumOneEthereum, ChainIDs.OptimismEthereum, ChainIDs.BaseEthereum, ChainIDs.Gnosis]
  if (sortByPriority) {
    list.sort((a, b) => { return ChainInfo[a].priority - ChainInfo[b].priority })
  }
  return list
}

interface ChainInfoFields {
  name: string,
  family: NetworkFamily,
  mainNet: ChainIDs, // if the network is a testnet, this field points to the non-test network
  L1: ChainIDs, // if the network is a L2, this field points to the L1
  clCurrency: CryptoCurrency,
  elCurrency: CryptoCurrency,
  path: string,
  priority: number // default order of the networks on the screen (ex: in the drop-down of the search bar)
}

export const ChainInfo: Record<ChainIDs, ChainInfoFields> = {
  [ChainIDs.Any]: {
    name: 'Any',
    family: 'Any',
    mainNet: ChainIDs.Any,
    L1: ChainIDs.Any,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/undefined',
    priority: 0 // data belonging to all networks is displayed first by default
  },

  [ChainIDs.Ethereum]: {
    name: 'Ethereum Mainnet',
    family: 'Ethereum',
    mainNet: ChainIDs.Ethereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/ethereum',
    priority: 1
  },
  [ChainIDs.Holesky]: {
    name: 'Holesky Testnet',
    family: 'Ethereum',
    mainNet: ChainIDs.Ethereum,
    L1: ChainIDs.Holesky,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/holesky',
    priority: 2
  },
  [ChainIDs.Sepolia]: {
    name: 'Sepolia Testnet',
    family: 'Ethereum',
    mainNet: ChainIDs.Ethereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/sepolia',
    priority: 3
  },

  [ChainIDs.ArbitrumOneEthereum]: {
    name: 'Arbitrum One L2',
    family: 'Arbitrum',
    mainNet: ChainIDs.ArbitrumOneEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/arbitrum-one-ethereum',
    priority: 10
  },
  [ChainIDs.ArbitrumNovaEthereum]: {
    name: 'Arbitrum Nova L2',
    family: 'Arbitrum',
    mainNet: ChainIDs.ArbitrumNovaEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/arbitrum-nova-ethereum',
    priority: 11
  },
  [ChainIDs.ArbitrumOneSepolia]: {
    name: 'Arbitrum One Sepolia Testnet',
    family: 'Arbitrum',
    mainNet: ChainIDs.ArbitrumOneEthereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/arbitrum-one-sepolia',
    priority: 12
  },

  [ChainIDs.OptimismEthereum]: {
    name: 'Optimism L2',
    family: 'Optimism',
    mainNet: ChainIDs.OptimismEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/optimism-ethereum',
    priority: 20
  },
  [ChainIDs.OptimismSepolia]: {
    name: 'Optimism Sepolia Testnet',
    family: 'Optimism',
    mainNet: ChainIDs.OptimismEthereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/optimism-sepolia',
    priority: 21
  },

  [ChainIDs.BaseEthereum]: {
    name: 'Base L2',
    family: 'Base',
    mainNet: ChainIDs.BaseEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/base-ethereum',
    priority: 30
  },
  [ChainIDs.BaseSepolia]: {
    name: 'Base Sepolia Testnet',
    family: 'Base',
    mainNet: ChainIDs.BaseEthereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/base-sepolia',
    priority: 31
  },

  [ChainIDs.Gnosis]: {
    name: 'Gnosis',
    family: 'Gnosis',
    mainNet: ChainIDs.Gnosis,
    L1: ChainIDs.Gnosis,
    clCurrency: 'GNO',
    elCurrency: 'xDAI',
    path: '/gnosis',
    priority: 40
  },
  [ChainIDs.Chiado]: {
    name: 'Gnosis Chiado Testnet',
    family: 'Gnosis',
    mainNet: ChainIDs.Gnosis,
    L1: ChainIDs.Chiado,
    clCurrency: 'GNO',
    elCurrency: 'xDAI',
    path: '/chiado',
    priority: 41
  }
}

export function isMainNet (network: ChainIDs) : boolean {
  return (ChainInfo[network].mainNet === network)
}

export function isL1 (network: ChainIDs) : boolean {
  return (ChainInfo[network].L1 === network)
}

export function getListOfChainIDs (sortByPriority : boolean) : ChainIDs[] {
  const list : ChainIDs[] = []

  for (const id in ChainIDs) {
    if (isNaN(Number(id))) {
      list.push(ChainIDs[id as keyof typeof ChainIDs])
    }
  }
  if (sortByPriority) {
    list.sort((a, b) => { return ChainInfo[a].priority - ChainInfo[b].priority })
  }
  return list
}
