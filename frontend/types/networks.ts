/* This file sets types and fundamental data about different networks.

   To use it, you need to write first:
     import { ChainIDs, ChainInfo } from '~/types/networks'

   First, the file defines identifiers equal to the chain IDs of the networks.
   In your code, you write ChainIDs.Mainnet whenever you want to represent the main Ethereum network,
   or ChainIDs.Sepolia for the Sepolia testnet and so on.
   Those constants are integers but for a safer code you should define your constants, fields, variables
   and parameters as the type ChainIDs.

   The most important feature of this file is to provide a mapping between those chain IDs and
   information about the networks.
   For example, when your variable myNetwork is equal to 10200 (namely ChainIDs.GnosisChiado, a testnet of Gnosis) :
   * ChainInfo[myNetwork].path  is the beginning of the path used to address this network in API endpoints.
   * ChainInfo[myNetwork].mainNet  is equal to the chain ID of the mainnet of Gnosis (100)
     /!\ not the ID of the testnet! if you want the testnet ID, it is simply myNetwork.
     So, to test whether your network is a testnet, you can do  myNetwork != ChainInfo[myNetwork].mainNet
   * ChainInfo[myNetwork].elCurrency is equal to 'xDAI' whereas ChainInfo[myNetwork].clCurrency is 'GNO'
*/

import type { CryptoCurrency } from '~/types/currencies'

export const enum ChainIDs {
  Ethereum = 1,
  Holesky = 17000,
  Sepolia = 11155111,
  Goerli = 5,

  ArbitrumOne = 42161,
  ArbitrumSepolia = 421614,

  Optimism = 10,
  OptimismGoerli = 420,

  Base = 8453,
  BaseSepolia = 84532,

  Gnosis = 100,
  GnosisChiado = 10200
}

interface ChainInfoFields {
  name: string,
  mainNet: ChainIDs,
  clCurrency: CryptoCurrency,
  elCurrency: CryptoCurrency,
  path: string
}

export const ChainInfo: Record<ChainIDs, ChainInfoFields> = {
  [ChainIDs.Ethereum]: { name: 'Ethereum Mainnet', mainNet: ChainIDs.Ethereum, clCurrency: 'ETH', elCurrency: 'ETH', path: '/ethereum' },
  [ChainIDs.Holesky]: { name: 'Holesky Testnet', mainNet: ChainIDs.Ethereum, clCurrency: 'ETH', elCurrency: 'ETH', path: '/holesky' },
  [ChainIDs.Sepolia]: { name: 'Sepolia Testnet', mainNet: ChainIDs.Ethereum, clCurrency: 'ETH', elCurrency: 'ETH', path: '/sepolia' },
  [ChainIDs.Goerli]: { name: 'Görli Testnet', mainNet: ChainIDs.Ethereum, clCurrency: 'ETH', elCurrency: 'ETH', path: '/goerli' },

  [ChainIDs.ArbitrumOne]: { name: 'Arbitrum One L2', mainNet: ChainIDs.ArbitrumOne, clCurrency: 'ETH', elCurrency: 'ETH', path: '/arbitrumone' },
  [ChainIDs.ArbitrumSepolia]: { name: 'Arbitrum Sepolia Testnet', mainNet: ChainIDs.ArbitrumOne, clCurrency: 'ETH', elCurrency: 'ETH', path: '/arbitrumsepolia' },

  [ChainIDs.Optimism]: { name: 'Optimism L2', mainNet: ChainIDs.Optimism, clCurrency: 'ETH', elCurrency: 'ETH', path: '/optimism' },
  [ChainIDs.OptimismGoerli]: { name: 'Optimism Görli Testnet', mainNet: ChainIDs.Optimism, clCurrency: 'ETH', elCurrency: 'ETH', path: '/optimismgoerli' },

  [ChainIDs.Base]: { name: 'Base L2', mainNet: ChainIDs.Base, clCurrency: 'ETH', elCurrency: 'ETH', path: '/base' },
  [ChainIDs.BaseSepolia]: { name: 'Base Sepolia Testnet', mainNet: ChainIDs.Base, clCurrency: 'ETH', elCurrency: 'ETH', path: '/basesepolia' },

  [ChainIDs.Gnosis]: { name: 'Gnosis', mainNet: ChainIDs.Gnosis, clCurrency: 'GNO', elCurrency: 'xDAI', path: '/gnosis' },
  [ChainIDs.GnosisChiado]: { name: 'Gnosis Chiado Testnet', mainNet: ChainIDs.Gnosis, clCurrency: 'GNO', elCurrency: 'xDAI', path: '/gnosischiado' }
}
