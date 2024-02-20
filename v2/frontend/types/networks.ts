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

/* **** TO DO ****
   Get informations from the endpoint GET /api/i/networks that returns a list of all network configs
   ***************
*/

import type { CryptoCurrency } from '~/types/currencies'

export enum ChainIDs {
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

export function getListOfImplementedChainIDs () : ChainIDs[] {
  return [ChainIDs.Ethereum, ChainIDs.ArbitrumOneEthereum, ChainIDs.OptimismEthereum, ChainIDs.BaseEthereum, ChainIDs.Gnosis]
}

interface ChainInfoFields {
  name: string,
  mainNet: ChainIDs, // if the network is a testnet, this field points to the non-test network
  L1: ChainIDs, // if the network is a L2, this field points to the L1
  clCurrency: CryptoCurrency,
  elCurrency: CryptoCurrency,
  path: string,
  priority: number // preference order when displaying data for several networks and no order is requested (ex: search bar)
}

export const ChainInfo: Record<ChainIDs, ChainInfoFields> = {
  [ChainIDs.Ethereum]: {
    name: 'Ethereum Mainnet',
    mainNet: ChainIDs.Ethereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/ethereum',
    priority: 1
  },
  [ChainIDs.Holesky]: {
    name: 'Holesky Testnet',
    mainNet: ChainIDs.Ethereum,
    L1: ChainIDs.Holesky,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/holesky',
    priority: 2
  },
  [ChainIDs.Sepolia]: {
    name: 'Sepolia Testnet',
    mainNet: ChainIDs.Ethereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/sepolia',
    priority: 3
  },

  [ChainIDs.ArbitrumOneEthereum]: {
    name: 'Arbitrum One L2',
    mainNet: ChainIDs.ArbitrumOneEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/arbitrum-one-ethereum',
    priority: 10
  },
  [ChainIDs.ArbitrumNovaEthereum]: {
    name: 'Arbitrum Nova L2',
    mainNet: ChainIDs.ArbitrumNovaEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/arbitrum-nova-ethereum',
    priority: 11
  },
  [ChainIDs.ArbitrumOneSepolia]: {
    name: 'Arbitrum One Sepolia Testnet',
    mainNet: ChainIDs.ArbitrumOneEthereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/arbitrum-one-sepolia',
    priority: 12
  },

  [ChainIDs.OptimismEthereum]: {
    name: 'Optimism L2',
    mainNet: ChainIDs.OptimismEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/optimism-ethereum',
    priority: 20
  },
  [ChainIDs.OptimismSepolia]: {
    name: 'Optimism Sepolia Testnet',
    mainNet: ChainIDs.OptimismEthereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/optimism-sepolia',
    priority: 21
  },

  [ChainIDs.BaseEthereum]: {
    name: 'Base L2',
    mainNet: ChainIDs.BaseEthereum,
    L1: ChainIDs.Ethereum,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/base-ethereum',
    priority: 30
  },
  [ChainIDs.BaseSepolia]: {
    name: 'Base Sepolia Testnet',
    mainNet: ChainIDs.BaseEthereum,
    L1: ChainIDs.Sepolia,
    clCurrency: 'ETH',
    elCurrency: 'ETH',
    path: '/base-sepolia',
    priority: 31
  },

  [ChainIDs.Gnosis]: {
    name: 'Gnosis',
    mainNet: ChainIDs.Gnosis,
    L1: ChainIDs.Gnosis,
    clCurrency: 'GNO',
    elCurrency: 'xDAI',
    path: '/gnosis',
    priority: 40
  },
  [ChainIDs.Chiado]: {
    name: 'Gnosis Chiado Testnet',
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

export function getListOfChainIDs () : ChainIDs[] {
  const list : ChainIDs[] = []

  for (const id in ChainIDs) {
    if (isNaN(Number(id))) {
      list.push(ChainIDs[id as keyof typeof ChainIDs])
    }
  }
  return list
}
